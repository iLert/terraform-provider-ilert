package ilert

import (
	"strings"
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

func TestBuildEscalationPolicy_WithEscalationRuleTeams(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceEscalationPolicy().Schema, map[string]any{
		"name": "test-escalation-policy",
		"escalation_rule": []any{
			map[string]any{
				"escalation_timeout": 15,
				"teams": []any{
					map[string]any{
						"id":   "123",
						"name": "Platform Team",
					},
				},
			},
		},
	})

	escalationPolicy, err := buildEscalationPolicy(d)
	if err != nil {
		t.Fatalf("unexpected error building escalation policy: %v", err)
	}

	if len(escalationPolicy.EscalationRules) != 1 {
		t.Fatalf("expected 1 escalation rule, got %d", len(escalationPolicy.EscalationRules))
	}

	rule := escalationPolicy.EscalationRules[0]
	if len(rule.Teams) != 1 {
		t.Fatalf("expected 1 escalation rule team, got %d", len(rule.Teams))
	}
	if rule.Teams[0].ID != 123 {
		t.Fatalf("expected escalation rule team id 123, got %d", rule.Teams[0].ID)
	}
	if rule.Teams[0].Name != "Platform Team" {
		t.Fatalf("expected escalation rule team name 'Platform Team', got %s", rule.Teams[0].Name)
	}
}

func TestCheckEscalationRuleSchema_TeamsConflictsWithUserOrSchedule(t *testing.T) {
	tests := []struct {
		name string
		rule map[string]any
	}{
		{
			name: "teams conflicts with user",
			rule: map[string]any{
				"user":      "123",
				"schedule":  "",
				"users":     []any{},
				"schedules": []any{},
				"teams": []any{
					map[string]any{"id": "1"},
				},
			},
		},
		{
			name: "teams conflicts with schedule",
			rule: map[string]any{
				"user":      "",
				"schedule":  "456",
				"users":     []any{},
				"schedules": []any{},
				"teams": []any{
					map[string]any{"id": "1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkEscalationRuleSchema(tt.rule)
			if err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), "teams") {
				t.Fatalf("expected validation error to mention teams, got: %s", err.Error())
			}
		})
	}
}

func TestFlattenEscalationRulesList_WithRuleTeams(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceEscalationPolicy().Schema, map[string]any{
		"name": "test-escalation-policy",
		"escalation_rule": []any{
			map[string]any{
				"escalation_timeout": 15,
				"teams": []any{
					map[string]any{
						"id":   "1",
						"name": "Team 1",
					},
				},
			},
		},
	})

	rules := []ilertapi.EscalationRule{
		{
			EscalationTimeout: 15,
			Teams: []ilertapi.TeamShort{
				{ID: 1, Name: "Team 1"},
			},
		},
	}

	flattened, err := flattenEscalationRulesList(rules, d)
	if err != nil {
		t.Fatalf("unexpected error flattening escalation rules: %v", err)
	}
	if len(flattened) != 1 {
		t.Fatalf("expected 1 flattened escalation rule, got %d", len(flattened))
	}

	flattenedRule := flattened[0].(map[string]any)
	teams := flattenedRule["teams"].([]any)
	if len(teams) != 1 {
		t.Fatalf("expected 1 flattened team, got %d", len(teams))
	}

	team := teams[0].(map[string]any)
	if team["id"].(string) != "1" {
		t.Fatalf("expected flattened team id '1', got %s", team["id"].(string))
	}
	if team["name"].(string) != "Team 1" {
		t.Fatalf("expected flattened team name 'Team 1', got %s", team["name"].(string))
	}
}
