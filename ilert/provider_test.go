package ilert

import (
	"context"
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

// early_access_enabled opts in to each early access resource on its own, and
// nothing is enabled by default
func TestProviderConfiguresEarlyAccess(t *testing.T) {
	for _, tc := range []struct {
		name    string
		enabled []any
		want    map[string]bool
	}{
		{"off by default", nil, map[string]bool{"ilert_agent": false, "ilert_credential": false}},
		{"one resource", []any{"ilert_agent"}, map[string]bool{"ilert_agent": true, "ilert_credential": false}},
		{"both", []any{"ilert_agent", "ilert_credential"}, map[string]bool{"ilert_agent": true, "ilert_credential": true}},
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

	// a name that is not an early access resource is rejected at plan time
	config := terraform.NewResourceConfigRaw(map[string]any{"api_token": "test-token", "early_access_enabled": []any{"ilert_agents"}})
	if diags := Provider().Validate(config); !diags.HasError() {
		t.Error("early_access_enabled accepted ilert_agents, want only the early access resources")
	}
}

// Every name early_access_enabled takes is an early access resource with a data
// source of the same name, mapped to the ilert-go API it calls
func TestEarlyAccessResourcesAreMapped(t *testing.T) {
	p := Provider()
	if len(earlyAccessAPIs) != len(earlyAccessResources) {
		t.Errorf("%d resources take early_access_enabled, %d are mapped to an API", len(earlyAccessResources), len(earlyAccessAPIs))
	}
	for _, typeName := range earlyAccessResources {
		api, ok := earlyAccessAPIs[typeName]
		if !ok || !slices.Contains(ilertapi.EarlyAccessAll, api) {
			t.Errorf("%s is mapped to %q, want an early access API of ilert-go", typeName, api)
		}
		if p.ResourcesMap[typeName] == nil || p.DataSourcesMap[typeName] == nil {
			t.Errorf("%s is not registered as both a resource and a data source", typeName)
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
		want := fmt.Sprintf("must be enabled by adding %q to early_access_enabled", typeName)
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

	for typeName, r := range map[string]*schema.Resource{"ilert_agent": resourceAgent(), "ilert_credential": resourceCredential()} {
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
	// the agent import passes the id through to the read, the credential import
	// reads the credential itself
	d := resourceCredential().TestResourceData()
	d.SetId("1")
	_, err := resourceCredential().Importer.StateContext(ctx, d, client)
	expectRefused("ilert_credential import", "ilert_credential", err)

	for typeName, r := range map[string]*schema.Resource{"ilert_agent": dataSourceAgent(), "ilert_credential": dataSourceCredential()} {
		d := r.TestResourceData()
		d.Set("name", "triage")
		expectRefused("data source "+typeName, typeName, diagErr(r.ReadContext(ctx, d, client)))
	}

	if n := requests.Load(); n != 0 {
		t.Errorf("%d requests were sent without early access, want none", n)
	}

	// enabling one does not enable the other
	ilertapi.WithEarlyAccess(ilertapi.EarlyAccess.Agents)(client)
	if err := requireEarlyAccess(client, "ilert_agent"); err != nil {
		t.Errorf("ilert_agent enabled: err = %v, want none", err)
	}
	expectRefused("ilert_credential while only ilert_agent is enabled", "ilert_credential", requireEarlyAccess(client, "ilert_credential"))

	if err := requireEarlyAccess(nil, "ilert_agent"); err != nil {
		t.Errorf("no configured client: err = %v, want none, ilert-go refuses the call itself", err)
	}
}

type diagError struct{ summary string }

func (e *diagError) Error() string { return e.summary }
