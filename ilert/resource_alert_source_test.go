package ilert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	teams := d.Get("team").([]any)
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams in state, got %d", len(teams))
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
