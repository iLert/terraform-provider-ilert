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
