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
