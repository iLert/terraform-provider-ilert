package ilert

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ilertapi "github.com/iLert/ilert-go/v3"
)

func TestTransformAlertSourceResource_DoesNotPanicWhenAPIReturnsMoreTeamsThanState(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		"team": []any{
			map[string]any{
				"id":   1,
				"name": "Team 1",
			},
		},
	})

	alertSource := &ilertapi.AlertSource{
		Name: "test-alert-source",
		EscalationPolicy: &ilertapi.EscalationPolicy{
			ID: 1,
		},
		Teams: []ilertapi.TeamShort{
			{ID: 1, Name: "Team 1"},
			{ID: 2, Name: "Team 2"},
		},
	}

	if err := transformAlertSourceResource(alertSource, d); err != nil {
		t.Fatalf("unexpected error transforming alert source: %v", err)
	}

	teams := d.Get("team").(*schema.Set).List()
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams in state, got %d", len(teams))
	}
}

// Regression for #150: the API returns teams sorted by id, which need not match
// the order they were declared in. The read must bind each configured name to its
// own team id rather than zipping the two lists by position.
func TestTransformAlertSourceResource_BindsTeamNamesByIDNotPosition(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		// declared in the reverse of the order the API returns
		"team": []any{
			map[string]any{"id": 2, "name": "Team 2"},
			map[string]any{"id": 1, "name": "Team 1"},
		},
	})

	alertSource := &ilertapi.AlertSource{
		Name:             "test-alert-source",
		EscalationPolicy: &ilertapi.EscalationPolicy{ID: 1},
		Teams: []ilertapi.TeamShort{
			{ID: 1, Name: "Team 1"},
			{ID: 2, Name: "Team 2"},
		},
	}

	if err := transformAlertSourceResource(alertSource, d); err != nil {
		t.Fatalf("unexpected error transforming alert source: %v", err)
	}

	got := make(map[int]string)
	for _, item := range d.Get("team").(*schema.Set).List() {
		v := item.(map[string]any)
		got[v["id"].(int)] = v["name"].(string)
	}

	if got[1] != "Team 1" || got[2] != "Team 2" {
		t.Fatalf("team names bound to the wrong ids: %v", got)
	}
}

// testResourceDataForUpdate builds the ResourceData an update sees: the prior
// state diffed against the new configuration. TestResourceDataRaw has no prior
// state, so it cannot express an attribute that was removed from the config.
func testResourceDataForUpdate(t *testing.T, resource *schema.Resource, priorRaw, raw map[string]any) *schema.ResourceData {
	t.Helper()

	prior := schema.TestResourceDataRaw(t, resource.Schema, priorRaw)
	prior.SetId("1")

	state := prior.State()
	resourceSchema := schema.InternalMap(resource.Schema)

	diff, err := resourceSchema.Diff(context.Background(), state, terraform.NewResourceConfigRaw(raw), nil, nil, true)
	if err != nil {
		t.Fatalf("unexpected error diffing: %v", err)
	}

	d, err := resourceSchema.Data(state, diff)
	if err != nil {
		t.Fatalf("unexpected error building resource data: %v", err)
	}

	return d
}

// Regression for #151: with every team block removed the API only clears teams on
// an explicit empty array. An omitted or null "teams" field leaves them untouched,
// so the built payload must carry a non-nil empty slice.
func TestBuildAlertSource_RemovingAllTeamsSendsEmptyArray(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceAlertSource(),
		map[string]any{
			"name":              "test-alert-source",
			"integration_type":  "API",
			"escalation_policy": "1",
			"team": []any{
				map[string]any{"id": 1, "name": "Team 1"},
				map[string]any{"id": 2, "name": "Team 2"},
			},
		},
		map[string]any{
			"name":              "test-alert-source",
			"integration_type":  "API",
			"escalation_policy": "1",
		})

	alertSource, err := buildAlertSource(d)
	if err != nil {
		t.Fatalf("unexpected error building alert source: %v", err)
	}

	if alertSource.Teams == nil {
		t.Fatal("expected non-nil empty teams slice, got nil (marshals to null, which the API ignores)")
	}
	if len(alertSource.Teams) != 0 {
		t.Fatalf("expected 0 teams, got %d", len(alertSource.Teams))
	}

	payload, err := json.Marshal(alertSource)
	if err != nil {
		t.Fatalf("unexpected error marshalling alert source: %v", err)
	}
	if !strings.Contains(string(payload), `"teams":[]`) {
		t.Fatalf("expected payload to contain \"teams\":[], got %s", payload)
	}
}

// The empty array must stay scoped to alert sources whose team blocks were
// actually removed. A config that never declared one is not a statement that the
// alert source has no teams, so an unrelated update must leave whatever is
// assigned to it elsewhere alone: the field has to be omitted, not sent empty.
func TestBuildAlertSource_KeepsTeamsWhenNeverDeclared(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceAlertSource(),
		map[string]any{
			"name":              "test-alert-source",
			"integration_type":  "API",
			"escalation_policy": "1",
		},
		map[string]any{
			"name":              "test-alert-source-renamed",
			"integration_type":  "API",
			"escalation_policy": "1",
		})

	alertSource, err := buildAlertSource(d)
	if err != nil {
		t.Fatalf("unexpected error building alert source: %v", err)
	}

	if alertSource.Teams != nil {
		t.Fatalf("expected teams to be omitted, got %v", alertSource.Teams)
	}

	payload, err := json.Marshal(alertSource)
	if err != nil {
		t.Fatalf("unexpected error marshalling alert source: %v", err)
	}
	if !strings.Contains(string(payload), `"teams":null`) {
		t.Fatalf("expected payload to contain \"teams\":null, got %s", payload)
	}
}

// The deprecated top-level "teams" field must keep working: the empty-array fallback
// above must not clobber teams supplied through it.
func TestBuildAlertSource_DeprecatedTeamsFieldNotClobbered(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		"teams":             []any{1, 2},
	})

	alertSource, err := buildAlertSource(d)
	if err != nil {
		t.Fatalf("unexpected error building alert source: %v", err)
	}

	if len(alertSource.Teams) != 2 {
		t.Fatalf("expected 2 teams from the deprecated field, got %d", len(alertSource.Teams))
	}
}

func TestTransformAlertSourceResource_DoesNotClobberFilterOperatorDefaultWhenAPIReturnsEmpty(t *testing.T) {
	// filter_operator/resolve_filter_operator are deprecated fields with a schema
	// default of "AND". The API returns them empty for non-email sources, so the
	// read must not overwrite the "AND" already in state with "". Doing so causes
	// a perpetual "+ filter_operator = AND" diff on every plan.
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":                    "test-alert-source",
		"integration_type":        "CLOUDWATCH",
		"escalation_policy":       "1",
		"filter_operator":         "AND",
		"resolve_filter_operator": "AND",
	})

	alertSource := &ilertapi.AlertSource{
		Name: "test-alert-source",
		EscalationPolicy: &ilertapi.EscalationPolicy{
			ID: 1,
		},
		FilterOperator:        "",
		ResolveFilterOperator: "",
	}

	if err := transformAlertSourceResource(alertSource, d); err != nil {
		t.Fatalf("unexpected error transforming alert source: %v", err)
	}

	if got := d.Get("filter_operator").(string); got != "AND" {
		t.Fatalf("expected filter_operator to remain \"AND\", got %q", got)
	}
	if got := d.Get("resolve_filter_operator").(string); got != "AND" {
		t.Fatalf("expected resolve_filter_operator to remain \"AND\", got %q", got)
	}
}

func TestTransformAlertSourceResource_FillsFilterOperatorDefaultOnImport(t *testing.T) {
	// On import there is no prior state to fall back on. Leaving the deprecated
	// filter operators unset leaves them null in state while the schema default
	// fills the configuration with "AND", which reads as a permanent
	// "+ filter_operator = AND" diff on every plan after the import.
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{})

	alertSource := &ilertapi.AlertSource{
		Name: "test-alert-source",
		EscalationPolicy: &ilertapi.EscalationPolicy{
			ID: 1,
		},
		IntegrationType:       "EMAIL2",
		FilterOperator:        "",
		ResolveFilterOperator: "",
	}

	if err := transformAlertSourceResource(alertSource, d); err != nil {
		t.Fatalf("unexpected error transforming alert source: %v", err)
	}

	if got := d.Get("filter_operator").(string); got != "AND" {
		t.Fatalf("expected filter_operator to default to \"AND\", got %q", got)
	}
	if got := d.Get("resolve_filter_operator").(string); got != "AND" {
		t.Fatalf("expected resolve_filter_operator to default to \"AND\", got %q", got)
	}
}

func TestTransformAlertSourceResource_KeepsFilterOperatorReturnedByAPI(t *testing.T) {
	// Email sources do carry the deprecated operators, and a value the API
	// returns always wins over the schema default.
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{})

	alertSource := &ilertapi.AlertSource{
		Name: "test-alert-source",
		EscalationPolicy: &ilertapi.EscalationPolicy{
			ID: 1,
		},
		IntegrationType:       "EMAIL",
		FilterOperator:        "OR",
		ResolveFilterOperator: "OR",
	}

	if err := transformAlertSourceResource(alertSource, d); err != nil {
		t.Fatalf("unexpected error transforming alert source: %v", err)
	}

	if got := d.Get("filter_operator").(string); got != "OR" {
		t.Fatalf("expected filter_operator to stay \"OR\", got %q", got)
	}
	if got := d.Get("resolve_filter_operator").(string); got != "OR" {
		t.Fatalf("expected resolve_filter_operator to stay \"OR\", got %q", got)
	}
}

func TestFlattenSeverityTemplate(t *testing.T) {
	// nil severity template flattens to an empty result
	empty, err := flattenSeverityTemplate(nil)
	if err != nil {
		t.Fatalf("unexpected error flattening nil severity template: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty result for nil severity template, got %d", len(empty))
	}

	// populated severity template round-trips into the schema shape
	severityTemplate := &ilertapi.SeverityTemplate{
		ValueTemplate: &ilertapi.Template{TextTemplate: "{{ event.customDetails.sev }}"},
		Mappings: []ilertapi.SeverityMapping{
			{Value: "critical", Severity: 1},
			{Value: "warning", Severity: 3},
		},
	}

	result, err := flattenSeverityTemplate(severityTemplate)
	if err != nil {
		t.Fatalf("unexpected error flattening severity template: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 flattened block, got %d", len(result))
	}
	block := result[0].(map[string]any)

	valueTemplate := block["value_template"].([]any)
	if len(valueTemplate) != 1 {
		t.Fatalf("expected 1 value_template, got %d", len(valueTemplate))
	}
	if got := valueTemplate[0].(map[string]any)["text_template"]; got != "{{ event.customDetails.sev }}" {
		t.Fatalf("unexpected text_template: %v", got)
	}

	mappings := block["mapping"].([]any)
	if len(mappings) != 2 {
		t.Fatalf("expected 2 mappings, got %d", len(mappings))
	}
	if first := mappings[0].(map[string]any); first["value"] != "critical" || first["severity"] != 1 {
		t.Fatalf("unexpected first mapping: %v", first)
	}
	if second := mappings[1].(map[string]any); second["value"] != "warning" || second["severity"] != 3 {
		t.Fatalf("unexpected second mapping: %v", second)
	}
}

func TestBuildAlertSourceSeverity(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		"severity":          4,
		"severity_template": []any{
			map[string]any{
				"value_template": []any{
					map[string]any{"text_template": "{{ event.customDetails.sev }}"},
				},
				"mapping": []any{
					map[string]any{"value": "critical", "severity": 1},
					map[string]any{"value": "warning", "severity": 3},
				},
			},
		},
	})

	alertSource, err := buildAlertSource(d)
	if err != nil {
		t.Fatalf("unexpected error building alert source: %v", err)
	}

	if alertSource.Severity != 4 {
		t.Fatalf("expected severity 4, got %d", alertSource.Severity)
	}

	st := alertSource.SeverityTemplate
	if st == nil {
		t.Fatal("expected severity template, got nil")
	}
	if st.ValueTemplate == nil || st.ValueTemplate.TextTemplate != "{{ event.customDetails.sev }}" {
		t.Fatalf("unexpected value template: %#v", st.ValueTemplate)
	}
	if len(st.Mappings) != 2 {
		t.Fatalf("expected 2 mappings, got %d", len(st.Mappings))
	}
	if st.Mappings[0].Value != "critical" || st.Mappings[0].Severity != 1 {
		t.Fatalf("unexpected first mapping: %#v", st.Mappings[0])
	}
	if st.Mappings[1].Value != "warning" || st.Mappings[1].Severity != 3 {
		t.Fatalf("unexpected second mapping: %#v", st.Mappings[1])
	}
}

func TestBuildCreateAlertSource_SetsSetupStatusFinished(t *testing.T) {
	// Sources created via Terraform must be sent with setupStatus FINISHED so the
	// API does not default them to CREATED (which surfaces a "Finish setup" prompt
	// in the UI for an already complete alert source).
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "CLOUDWATCH",
		"escalation_policy": "1",
	})

	alertSource, err := buildCreateAlertSource(d)
	if err != nil {
		t.Fatalf("unexpected error building alert source: %v", err)
	}

	if alertSource.SetupStatus != ilertapi.AlertSourceSetupStatuses.Finished {
		t.Fatalf("expected setup status %q, got %q", ilertapi.AlertSourceSetupStatuses.Finished, alertSource.SetupStatus)
	}
}

func TestTransformAlertSourceResource_FlattensServicesAndServicesTemplate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		"services": []any{
			map[string]any{
				"id": 1,
			},
		},
		"services_template": []any{
			map[string]any{
				"text_template": "{{ event.service }}",
			},
		},
		"auto_create_services": true,
	})

	alertSource := &ilertapi.AlertSource{
		Name: "test-alert-source",
		EscalationPolicy: &ilertapi.EscalationPolicy{
			ID: 1,
		},
		// API returns more services than were declared in state; the resource keeps all of them,
		// matching the behavior asserted for teams above.
		Services: []ilertapi.AlertSourceService{
			{ID: 1, Name: "Service 1"},
			{ID: 2, Name: "Service 2"},
		},
		ServicesTemplate: []ilertapi.Template{
			{TextTemplate: "{{ event.service }}"},
		},
		AutoCreateServices: true,
	}

	if err := transformAlertSourceResource(alertSource, d); err != nil {
		t.Fatalf("unexpected error transforming alert source: %v", err)
	}

	services := d.Get("services").([]any)
	if len(services) != 2 {
		t.Fatalf("expected 2 services in state, got %d", len(services))
	}
	// name must not be written when the user did not configure one, to avoid a perpetual diff
	if name, ok := services[0].(map[string]any)["name"].(string); ok && name != "" {
		t.Fatalf("expected no service name in state for id-only config, got %q", name)
	}

	servicesTemplate := d.Get("services_template").([]any)
	if len(servicesTemplate) != 1 {
		t.Fatalf("expected 1 services_template in state, got %d", len(servicesTemplate))
	}
	if tt := servicesTemplate[0].(map[string]any)["text_template"].(string); tt != "{{ event.service }}" {
		t.Fatalf("unexpected services_template text_template: %q", tt)
	}

	if !d.Get("auto_create_services").(bool) {
		t.Fatalf("expected auto_create_services to be true in state")
	}
}

func TestBuildAlertSource_ServicesAndServicesTemplate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		"services": []any{
			map[string]any{"id": 1},
			map[string]any{"id": 2, "name": "Service 2"},
		},
		"services_template": []any{
			map[string]any{"text_template": "{{ event.service }}"},
		},
		"auto_create_services": true,
	})

	alertSource, err := buildAlertSource(d)
	if err != nil {
		t.Fatalf("unexpected error building alert source: %v", err)
	}

	if len(alertSource.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(alertSource.Services))
	}
	if alertSource.Services[0].ID != 1 {
		t.Fatalf("expected first service id 1, got %d", alertSource.Services[0].ID)
	}
	// name only sent when configured: first has none, second does
	if alertSource.Services[0].Name != "" {
		t.Fatalf("expected first service name empty, got %q", alertSource.Services[0].Name)
	}
	if alertSource.Services[1].Name != "Service 2" {
		t.Fatalf("expected second service name 'Service 2', got %q", alertSource.Services[1].Name)
	}

	if len(alertSource.ServicesTemplate) != 1 || alertSource.ServicesTemplate[0].TextTemplate != "{{ event.service }}" {
		t.Fatalf("unexpected services template: %+v", alertSource.ServicesTemplate)
	}

	if !alertSource.AutoCreateServices {
		t.Fatalf("expected auto_create_services to be true")
	}
}

// Regression for #147: on update the computed integration_key still holds the
// previous address from state. It must not overwrite the new value coming from
// the "email" field, otherwise the email update silently no-ops.
func TestBuildAlertSource_EmailUpdateNotClobberedByStaleIntegrationKey(t *testing.T) {
	for _, integrationType := range []string{"EMAIL", "EMAIL2"} {
		t.Run(integrationType, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
				"name":              "test-alert-source",
				"integration_type":  integrationType,
				"escalation_policy": "1",
				"email":             "new-address",
				// simulates the computed value carried over from prior state
				"integration_key": "old-address@example.ilertnotify.dev",
			})

			alertSource, err := buildAlertSource(d)
			if err != nil {
				t.Fatalf("unexpected error building alert source: %v", err)
			}

			if alertSource.IntegrationKey != "new-address" {
				t.Fatalf("expected integration key %q, got %q", "new-address", alertSource.IntegrationKey)
			}
		})
	}
}

// Non-email sources must still honor integration_key (its address is server-computed
// and sent back on update); the #147 guard only skips EMAIL/EMAIL2.
func TestBuildAlertSource_NonEmailKeepsIntegrationKey(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertSource().Schema, map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		"integration_key":   "server-generated-key",
	})

	alertSource, err := buildAlertSource(d)
	if err != nil {
		t.Fatalf("unexpected error building alert source: %v", err)
	}

	if alertSource.IntegrationKey != "server-generated-key" {
		t.Fatalf("expected integration key %q, got %q", "server-generated-key", alertSource.IntegrationKey)
	}
}
