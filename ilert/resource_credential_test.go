package ilert

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ilertapi "github.com/iLert/ilert-go/v4"
)

// fakePlan stands in for the *schema.ResourceDiff of a plan: values holds the known
// arguments, unknown the ones whose value is only known on apply, id is set for a
// credential that exists and changed for one the plan changes
type fakePlan struct {
	values  map[string]string
	unknown map[string]bool
	id      string
	changed bool
}

func (p fakePlan) Id() string { return p.id }

func (p fakePlan) Get(key string) any { return p.values[key] }

func (p fakePlan) NewValueKnown(key string) bool { return !p.unknown[key] }

func (p fakePlan) HasChanges(...string) bool { return p.changed }

func TestValidateCredentialArguments(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]string
		unknown map[string]bool
		wantErr string
	}{
		{name: "bearer", values: map[string]string{"type": "BEARER", "secret_value": "s"}},
		{name: "basic", values: map[string]string{"type": "BASIC", "secret_value": "s", "username": "u"}},
		{name: "api key with username", values: map[string]string{"type": "API_KEY", "secret_value": "s", "username": "u"}},
		{name: "none", values: map[string]string{"type": "NONE", "host_url": "https://hooks.example.com/T0/secret"}},
		{name: "aws with access key id", values: map[string]string{"type": "AWS_SIGV4", "secret_value": "s", "username": "AKIA"}},
		{name: "aws assuming a role", values: map[string]string{"type": "AWS_SIGV4", "secret_value": "s", "role_arn": "arn:aws:iam::1:role/r"}},
		{name: "oauth2 refresh token", values: map[string]string{"type": "OAUTH2", "secret_value": "s", "idp_url": "https://idp.example.com/token", "oauth_method": "REFRESH_TOKEN"}},
		{name: "oauth2 client credentials", values: map[string]string{"type": "OAUTH2", "secret_value": "s", "idp_url": "https://idp.example.com/token", "oauth_method": "CLIENT_SECRET_POST", "idp_client": "c"}},
		{name: "secret known on apply", values: map[string]string{"type": "BEARER"}, unknown: map[string]bool{"secret_value": true}},
		{name: "type known on apply", values: map[string]string{}, unknown: map[string]bool{"type": true}},

		{name: "missing secret", values: map[string]string{"type": "BEARER"}, wantErr: "needs a secret_value"},
		{name: "blank secret", values: map[string]string{"type": "BEARER", "secret_value": "  "}, wantErr: "needs a secret_value"},
		{name: "basic without username", values: map[string]string{"type": "BASIC", "secret_value": "s"}, wantErr: "needs a username"},
		{name: "bearer with username", values: map[string]string{"type": "BEARER", "secret_value": "s", "username": "u"}, wantErr: "takes no username"},
		{name: "oauth2 with username", values: map[string]string{"type": "OAUTH2", "secret_value": "s", "idp_url": "https://idp.example.com/token", "username": "u"}, wantErr: "takes no username"},
		{name: "none without url", values: map[string]string{"type": "NONE"}, wantErr: "whole URL"},
		{name: "none with secret", values: map[string]string{"type": "NONE", "host_url": "https://hooks.example.com/x", "secret_value": "s"}, wantErr: "takes no secret_value"},
		{name: "aws without key or role", values: map[string]string{"type": "AWS_SIGV4", "secret_value": "s"}, wantErr: "access key id"},
		{name: "oauth2 without token endpoint", values: map[string]string{"type": "OAUTH2", "secret_value": "s"}, wantErr: "idp_url"},
		{name: "oauth2 client credentials without client", values: map[string]string{"type": "OAUTH2", "secret_value": "s", "idp_url": "https://idp.example.com/token", "oauth_method": "CLIENT_SECRET_BASIC"}, wantErr: "idp_client"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCredentialArguments(fakePlan{values: tc.values, unknown: tc.unknown})
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want one containing %q", err, tc.wantErr)
			}
		})
	}
}

// An existing credential the plan does not change is not created again, so it is
// not checked: an imported one has no secret in its configuration, and that must not
// fail the plan or replace it. Changed, it is created again and checked.
func TestValidateCredentialArguments_ExistingCredential(t *testing.T) {
	withoutSecret := map[string]string{"type": "BEARER"}
	if err := validateCredentialArguments(fakePlan{values: withoutSecret, id: "1"}); err != nil {
		t.Errorf("unchanged credential: unexpected error: %v", err)
	}
	if err := validateCredentialArguments(fakePlan{values: withoutSecret, id: "1", changed: true}); err == nil {
		t.Errorf("changed credential: expected the missing secret to be rejected")
	}
}

func TestCredentialURLOrigin(t *testing.T) {
	tests := map[string]string{
		"https://hooks.example.com/services/T0/B0/x": "https://hooks.example.com",
		"HTTPS://Hooks.Example.com/x?token=y":        "https://hooks.example.com",
		"http://example.com:8080/x":                  "http://example.com:8080",
		"http://[::1]:8080/x":                        "http://[::1]:8080",
		"https://example.com":                        "https://example.com",
	}
	for input, want := range tests {
		if got := credentialURLOrigin(input); got != want {
			t.Errorf("credentialURLOrigin(%q) = %q, want %q", input, got, want)
		}
	}
}

// The API keeps the path and query of a NONE credential's URL as its secret and
// returns only the origin. Taking that into the state would replace the credential on
// every plan, so the configured URL is kept while the origin matches.
func TestTransformCredentialResource_KeepsNoneURL(t *testing.T) {
	const configured = "https://hooks.example.com/services/T0/B0/secret"
	tests := []struct {
		name     string
		returned string
		want     string
	}{
		{"same origin", "https://hooks.example.com", configured},
		{"different origin", "https://other.example.com", "https://other.example.com"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceCredential().Schema, map[string]any{
				"name":     "webhook",
				"type":     "NONE",
				"host_url": configured,
			})
			d.SetId("1")
			err := transformCredentialResource(&ilertapi.Credential{ID: 1, Name: "webhook", TypeHint: "NONE", HostURL: tc.returned}, d)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := d.Get("host_url").(string); got != tc.want {
				t.Errorf("host_url = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTransformCredentialResource_KeepsSecretTheAPIDoesNotReturn(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceCredential().Schema, map[string]any{
		"name":         "github",
		"type":         "BEARER",
		"secret_value": "ghp_secret",
	})
	d.SetId("1")
	err := transformCredentialResource(&ilertapi.Credential{ID: 1, Name: "github", TypeHint: "BEARER"}, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := d.Get("secret_value").(string); got != "ghp_secret" {
		t.Errorf("secret_value = %q, want the configured secret kept", got)
	}
}

func TestValidateCredentialHostURL(t *testing.T) {
	valid := []string{
		"https://example.com",
		"http://example.com:8080/path?x=1&y=%20z",
		"https://hooks.slack.com/services/T0/B0/xyz",
		"http://[::1]:8080/x",
	}
	invalid := []string{
		"example.com",
		"hooks.slack.com/services/T0/B0/secret-token",
		"ftp://example.com",
		"https://user:pass@example.com",
		"https://",
		" https://example.com",
		"https://example.com/path ",
		"https://my_host.example.com/x",
		"https://example.com/a b",
		"https://example.com/x?q={a}",
		"https://example.com/x|y",
		"https://example.com/x?a[]=1",
	}
	for _, value := range valid {
		if _, errs := validateCredentialHostURL(value, "host_url"); len(errs) != 0 {
			t.Errorf("%q rejected: %v", value, errs)
		}
	}
	for _, value := range invalid {
		_, errs := validateCredentialHostURL(value, "host_url")
		if len(errs) == 0 {
			t.Errorf("%q accepted, want it rejected", value)
			continue
		}
		// the URL of a NONE credential is its secret, and Terraform prints the message
		if strings.Contains(errs[0].Error(), "secret-token") {
			t.Errorf("error for %q repeats the URL: %v", value, errs[0])
		}
	}
}

// Every change replaces a credential, so a team renamed in ilert must not show as a
// change: the configured name is kept.
func TestFlattenCredentialTeams_KeepsConfiguredName(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceCredential().Schema, map[string]any{
		"name": "github", "type": "BEARER", "secret_value": "s",
		"team": []any{map[string]any{"id": 1, "name": "ops"}, map[string]any{"id": 2}},
	})
	d.SetId("1")
	got := flattenCredentialTeams([]ilertapi.TeamShort{{ID: 1, Name: "operations"}, {ID: 2, Name: "platform"}}, d)
	want := []any{map[string]any{"id": int64(1), "name": "ops"}, map[string]any{"id": int64(2)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("teams = %v, want %v", got, want)
	}
}

// A credential cannot be updated, so a changed team has to replace it. ForceNew on
// the set alone only covers the number of teams, not a team swapped for another.
func TestResourceCredential_ChangedTeamForcesNew(t *testing.T) {
	base := map[string]any{"name": "github", "type": "BEARER", "secret_value": "s"}
	withTeam := func(team map[string]any) map[string]any {
		config := map[string]any{"team": []any{team}}
		for k, v := range base {
			config[k] = v
		}
		return config
	}
	tests := []struct {
		name         string
		prior, after map[string]any
	}{
		{"other team", withTeam(map[string]any{"id": 1}), withTeam(map[string]any{"id": 2})},
		{"team name added", withTeam(map[string]any{"id": 1}), withTeam(map[string]any{"id": 1, "name": "ops"})},
		{"team removed", withTeam(map[string]any{"id": 1}), base},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := testResourceDataForUpdate(t, resourceCredential(), tc.prior, tc.after)
			if !d.HasChange("team") {
				t.Fatalf("expected a team change")
			}
			prior := schema.TestResourceDataRaw(t, resourceCredential().Schema, tc.prior)
			prior.SetId("1")
			diff, err := schema.InternalMap(resourceCredential().Schema).Diff(context.Background(), prior.State(), terraform.NewResourceConfigRaw(tc.after), nil, nil, true)
			if err != nil {
				t.Fatalf("unexpected error diffing: %v", err)
			}
			if !diff.RequiresNew() {
				t.Errorf("diff does not require a new credential, Terraform would plan an update the resource cannot do")
			}
		})
	}
}
