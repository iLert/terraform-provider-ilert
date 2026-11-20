package ilert

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ilertapi "github.com/iLert/ilert-go/v4"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"ilert": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	// an API token is enough, the provider prefers it over the basic credentials
	if v := os.Getenv("ILERT_API_TOKEN"); v != "" {
		return
	}
	if v := os.Getenv("ILERT_ORGANIZATION"); v == "" {
		t.Fatal("ILERT_ORGANIZATION must be set for acceptance tests")
	}
	if v := os.Getenv("ILERT_USERNAME"); v == "" {
		t.Fatal("ILERT_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("ILERT_PASSWORD"); v == "" {
		t.Fatal("ILERT_PASSWORD must be set for acceptance tests")
	}
}

// early_access_enabled opts in to each early access API on its own, and nothing is
// enabled by default
func TestProviderConfiguresEarlyAccess(t *testing.T) {
	for _, tc := range []struct {
		name    string
		enabled []any
		want    map[string]bool
	}{
		{"off by default", nil, map[string]bool{"ilert_action": false, "ilert_agent": false, "ilert_credential": false}},
		{"one API", []any{"agents"}, map[string]bool{"ilert_action": false, "ilert_agent": true, "ilert_credential": false}},
		{"all", []any{"actions", "agents", "credentials"}, map[string]bool{"ilert_action": true, "ilert_agent": true, "ilert_credential": true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := map[string]any{"api_token": "test-token"}
			if tc.enabled != nil {
				config["early_access_enabled"] = tc.enabled
			}

			p := Provider()
			if diags := p.Configure(context.Background(), terraform.NewResourceConfigRaw(config)); diags.HasError() {
				t.Fatalf("configure: %v", diags)
			}
			client := p.Meta().(*ilertapi.Client)
			for typeName, want := range tc.want {
				if got := requireEarlyAccess(client, typeName) == nil; got != want {
					t.Errorf("%s enabled = %t, want %t", typeName, got, want)
				}
			}
		})
	}

	// a name that is not an early access API is rejected at plan time, the names of
	// the resources included
	for _, name := range []string{"agent", "ilert_agent"} {
		config := terraform.NewResourceConfigRaw(map[string]any{"api_token": "test-token", "early_access_enabled": []any{name}})
		if diags := Provider().Validate(config); !diags.HasError() {
			t.Errorf("early_access_enabled accepted %s, want only the early access APIs", name)
		}
	}
}

// Every early access resource has a data source of the same name, and is mapped to
// a name early_access_enabled takes, the one ilert-go refuses its calls under: the
// path of its API under /api/beta
func TestEarlyAccessResourcesAreMapped(t *testing.T) {
	p := Provider()
	for typeName, api := range earlyAccessResources {
		if !slices.Contains(earlyAccessAPIs, api) {
			t.Errorf("%s is mapped to %q, which early_access_enabled does not take", typeName, api)
		}
		if p.ResourcesMap[typeName] == nil || p.DataSourcesMap[typeName] == nil {
			t.Errorf("%s is not registered as both a resource and a data source", typeName)
		}
	}

	client := ilertapi.NewClient(ilertapi.WithAPIToken("test-token"))
	calls := map[string]func() error{
		"ilert_action":     func() error { _, err := client.GetActions(nil); return err },
		"ilert_agent":      func() error { _, err := client.GetAgents(nil); return err },
		"ilert_credential": func() error { _, err := client.GetCredentials(nil); return err },
	}
	if len(calls) != len(earlyAccessResources) {
		t.Errorf("%d early access resources, %d are checked against ilert-go", len(earlyAccessResources), len(calls))
	}
	for typeName, call := range calls {
		var disabled *ilertapi.EarlyAccessDisabledError
		if err := call(); !errors.As(err, &disabled) || disabled.API != earlyAccessResources[typeName] {
			t.Errorf("%s: ilert-go refused with %v, want the API %q", typeName, err, earlyAccessResources[typeName])
		}
	}
}

// Without early_access_enabled every operation of the early access resources and
// data sources fails, the plan included, and nothing is sent to the API
func TestEarlyAccessResourcesNeedTheProviderOptIn(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := ilertapi.NewClient(ilertapi.WithAPIEndpoint(srv.URL), ilertapi.WithAPIToken("test-token"))

	expectRefused := func(what, typeName string, err error) {
		t.Helper()
		want := fmt.Sprintf("must be enabled by adding %q to early_access_enabled", earlyAccessResources[typeName])
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err = %v, want it to say it %s", what, err, want)
		}
	}
	diagErr := func(diags diag.Diagnostics) error {
		for _, d := range diags {
			if d.Severity == diag.Error {
				return &diagError{d.Summary}
			}
		}
		return nil
	}

	for typeName, r := range map[string]*schema.Resource{"ilert_action": resourceAction(), "ilert_agent": resourceAgent(), "ilert_credential": resourceCredential()} {
		d := r.TestResourceData()
		d.SetId("1")

		expectRefused(typeName+" plan", typeName, r.CustomizeDiff(ctx, nil, client))
		expectRefused(typeName+" create", typeName, diagErr(r.CreateContext(ctx, d, client)))
		expectRefused(typeName+" read", typeName, diagErr(r.ReadContext(ctx, d, client)))
		if r.UpdateContext != nil {
			expectRefused(typeName+" update", typeName, diagErr(r.UpdateContext(ctx, d, client)))
		}
		expectRefused(typeName+" delete", typeName, diagErr(r.DeleteContext(ctx, d, client)))
		_, err := r.Exists(d, client)
		expectRefused(typeName+" exists", typeName, err)
	}
	// the agent import passes the id through to the read, the credential and action
	// imports read it themselves
	for typeName, r := range map[string]*schema.Resource{"ilert_action": resourceAction(), "ilert_credential": resourceCredential()} {
		d := r.TestResourceData()
		d.SetId("1")
		_, err := r.Importer.StateContext(ctx, d, client)
		expectRefused(typeName+" import", typeName, err)
	}

	for typeName, r := range map[string]*schema.Resource{"ilert_action": dataSourceAction(), "ilert_agent": dataSourceAgent(), "ilert_credential": dataSourceCredential()} {
		d := r.TestResourceData()
		d.Set("name", "triage")
		expectRefused("data source "+typeName, typeName, diagErr(r.ReadContext(ctx, d, client)))
	}

	if n := requests.Load(); n != 0 {
		t.Errorf("%d requests were sent without early access, want none", n)
	}

	// enabling one does not enable the other
	ilertapi.WithEarlyAccess("agents")(client)
	if err := requireEarlyAccess(client, "ilert_agent"); err != nil {
		t.Errorf("agents enabled: err = %v, want none", err)
	}
	expectRefused("ilert_credential while only agents are enabled", "ilert_credential", requireEarlyAccess(client, "ilert_credential"))

	if err := requireEarlyAccess(nil, "ilert_agent"); err != nil {
		t.Errorf("no configured client: err = %v, want none, ilert-go refuses the call itself", err)
	}
}

type diagError struct{ summary string }

func (e *diagError) Error() string { return e.summary }
