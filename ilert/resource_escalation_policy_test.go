package ilert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v3"
)

func TestTransformEscalationPolicyResource_DoesNotPanicWhenAPIReturnsMoreTeamsThanState(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceEscalationPolicy().Schema, map[string]any{
		"name": "test-escalation-policy",
		"escalation_rule": []any{
			map[string]any{
				"escalation_timeout": 0,
			},
		},
		"team": []any{
			map[string]any{
				"id":   1,
				"name": "Team 1",
			},
		},
	})

	escalationPolicy := &ilertapi.EscalationPolicy{
		Name: "test-escalation-policy",
		Teams: []ilertapi.TeamShort{
			{ID: 1, Name: "Team 1"},
			{ID: 2, Name: "Team 2"},
		},
	}

	if err := transformEscalationPolicyResource(escalationPolicy, d); err != nil {
		t.Fatalf("unexpected error transforming escalation policy: %v", err)
	}

	teams := d.Get("team").([]any)
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams in state, got %d", len(teams))
	}
}
