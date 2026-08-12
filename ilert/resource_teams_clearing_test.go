package ilert

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v3"
)

// teamsBuildCase describes one resource whose team blocks are managed by
// Terraform: the schema, a minimal configuration without team blocks, and how to
// build the API payload and read the teams back off it.
type teamsBuildCase struct {
	name     string
	resource *schema.Resource
	config   map[string]any
	// build returns the payload and its teams. teamsSet reports whether the field
	// was assigned at all, which for the pointer-typed fields is the difference
	// between an omitted field and an empty array.
	build func(d *schema.ResourceData) (payload any, teams []ilertapi.TeamShort, teamsSet bool, err error)
	// unsetOmitsField is true for the resources whose payload drops the teams
	// field entirely when it is not assigned, instead of marshalling it to null.
	// Their endpoints clear the teams on an explicit null.
	unsetOmitsField bool
}

func teamsBuildCases() []teamsBuildCase {
	return []teamsBuildCase{
		{
			name:     "service",
			resource: resourceService(),
			config:   map[string]any{"name": "test-service"},
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildService(d)
				if v == nil {
					return nil, nil, false, err
				}
				return v, v.Teams, v.Teams != nil, err
			},
		},
		{
			name:     "heartbeat monitor",
			resource: resourceHeartbeatMonitor(),
			config:   map[string]any{"name": "test-heartbeat-monitor", "interval_sec": 3600},
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildHeartbeatMonitor(d)
				if v == nil {
					return nil, nil, false, err
				}
				return v, v.Teams, v.Teams != nil, err
			},
		},
		{
			name:     "escalation policy",
			resource: resourceEscalationPolicy(),
			config:   map[string]any{"name": "test-escalation-policy"},
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildEscalationPolicy(d)
				if v == nil {
					return nil, nil, false, err
				}
				return v, v.Teams, v.Teams != nil, err
			},
		},
		{
			name:     "support hour",
			resource: resourceSupportHour(),
			config:   map[string]any{"name": "test-support-hour", "timezone": "Europe/Berlin"},
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildSupportHour(d)
				if v == nil {
					return nil, nil, false, err
				}
				return v, v.Teams, v.Teams != nil, err
			},
		},
		{
			name:     "incident template",
			resource: resourceIncidentTemplate(),
			config: map[string]any{
				"name":    "test-incident-template",
				"summary": "summary",
				"message": "message",
				"status":  "INVESTIGATING",
			},
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildIncidentTemplate(d)
				if v == nil {
					return nil, nil, false, err
				}
				return v, v.Teams, v.Teams != nil, err
			},
		},
		{
			name:     "schedule",
			resource: resourceSchedule(),
			config: map[string]any{
				"name":     "test-schedule",
				"timezone": "Europe/Berlin",
				"type":     "STATIC",
			},
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildSchedule(d)
				if v == nil {
					return nil, nil, false, err
				}
				return v, v.Teams, v.Teams != nil, err
			},
		},
		{
			name:     "status page",
			resource: resourceStatusPage(),
			config: map[string]any{
				"name":       "test-status-page",
				"subdomain":  "test-status-page",
				"visibility": "PUBLIC",
			},
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildStatusPage(d)
				if v == nil {
					return nil, nil, false, err
				}
				return v, v.Teams, v.Teams != nil, err
			},
		},
		{
			name:     "event flow",
			resource: resourceEventFlow(),
			config: map[string]any{
				"name":      "test-event-flow",
				"root_node": []any{map[string]any{"node_type": "ROOT"}},
			},
			unsetOmitsField: true,
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildEventFlow(d)
				if v == nil {
					return nil, nil, false, err
				}
				if v.Teams == nil {
					return v, nil, false, err
				}
				return v, *v.Teams, true, err
			},
		},
		{
			name:     "call flow",
			resource: resourceCallFlow(),
			config: map[string]any{
				"name":      "test-call-flow",
				"language":  "en",
				"root_node": []any{map[string]any{"node_type": "ROOT"}},
			},
			unsetOmitsField: true,
			build: func(d *schema.ResourceData) (any, []ilertapi.TeamShort, bool, error) {
				v, err := buildCallFlow(d)
				if v == nil {
					return nil, nil, false, err
				}
				if v.Teams == nil {
					return v, nil, false, err
				}
				return v, *v.Teams, true, err
			},
		},
	}
}

func withTeams(config map[string]any) map[string]any {
	withTeams := make(map[string]any, len(config)+1)
	for k, v := range config {
		withTeams[k] = v
	}
	withTeams["team"] = []any{
		map[string]any{"id": 1, "name": "Team 1"},
		map[string]any{"id": 2, "name": "Team 2"},
	}
	return withTeams
}

// Regression for iLert/engineering-tasks#2360, the same bug fixed for alert
// sources in #151: with every team block removed the API only clears the teams on
// an explicit empty array. An omitted or null teams field leaves them untouched,
// so the built payload must carry a non-nil empty slice.
func TestBuildResources_RemovingAllTeamsSendsEmptyArray(t *testing.T) {
	for _, tc := range teamsBuildCases() {
		t.Run(tc.name, func(t *testing.T) {
			d := testResourceDataForUpdate(t, tc.resource, withTeams(tc.config), tc.config)

			payload, teams, teamsSet, err := tc.build(d)
			if err != nil {
				t.Fatalf("unexpected error building payload: %v", err)
			}
			if !teamsSet {
				t.Fatalf("expected teams to be set to an empty array, got no teams field")
			}
			if len(teams) != 0 {
				t.Fatalf("expected teams to be empty, got %v", teams)
			}

			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("unexpected error marshalling payload: %v", err)
			}
			if !strings.Contains(string(encoded), `"teams":[]`) {
				t.Fatalf("expected payload to contain \"teams\":[], got %s", encoded)
			}
		})
	}
}

// The empty array must stay scoped to resources whose team blocks were actually
// removed. A config that never declared one is not a statement that the resource
// has no teams, so an unrelated update must leave whatever is assigned to it
// elsewhere alone: the field has to be omitted, not sent empty.
func TestBuildResources_KeepsTeamsWhenNeverDeclared(t *testing.T) {
	for _, tc := range teamsBuildCases() {
		t.Run(tc.name, func(t *testing.T) {
			renamed := make(map[string]any, len(tc.config))
			for k, v := range tc.config {
				renamed[k] = v
			}
			renamed["name"] = tc.config["name"].(string) + "-renamed"

			d := testResourceDataForUpdate(t, tc.resource, tc.config, renamed)

			payload, teams, teamsSet, err := tc.build(d)
			if err != nil {
				t.Fatalf("unexpected error building payload: %v", err)
			}
			if teamsSet {
				t.Fatalf("expected teams to be omitted, got %v", teams)
			}

			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("unexpected error marshalling payload: %v", err)
			}
			if tc.unsetOmitsField {
				// the call flow and event flow endpoints clear the teams on an
				// explicit null, so the field must not appear at all
				if strings.Contains(string(encoded), `"teams"`) {
					t.Fatalf("expected payload to omit the teams field, got %s", encoded)
				}
				return
			}
			if !strings.Contains(string(encoded), `"teams":null`) {
				t.Fatalf("expected payload to contain \"teams\":null, got %s", encoded)
			}
		})
	}
}

// The escalation policy's deprecated top-level "teams" field must keep working:
// the empty-array fallback must not clobber teams supplied through it.
func TestBuildEscalationPolicy_DeprecatedTeamsFieldNotClobbered(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceEscalationPolicy(),
		map[string]any{"name": "test-escalation-policy", "team": []any{
			map[string]any{"id": 1, "name": "Team 1"},
		}},
		map[string]any{"name": "test-escalation-policy", "teams": []any{1, 2}})

	escalationPolicy, err := buildEscalationPolicy(d)
	if err != nil {
		t.Fatalf("unexpected error building escalation policy: %v", err)
	}

	if len(escalationPolicy.Teams) != 2 {
		t.Fatalf("expected the deprecated teams field to be kept, got %v", escalationPolicy.Teams)
	}
}

// Removing a subset of the team blocks must keep sending the remaining ones.
func TestBuildResources_RemovingSomeTeamsSendsTheRest(t *testing.T) {
	for _, tc := range teamsBuildCases() {
		t.Run(tc.name, func(t *testing.T) {
			remaining := make(map[string]any, len(tc.config)+1)
			for k, v := range tc.config {
				remaining[k] = v
			}
			remaining["team"] = []any{map[string]any{"id": 2, "name": "Team 2"}}

			d := testResourceDataForUpdate(t, tc.resource, withTeams(tc.config), remaining)

			_, teams, _, err := tc.build(d)
			if err != nil {
				t.Fatalf("unexpected error building payload: %v", err)
			}
			if len(teams) != 1 || teams[0].ID != 2 {
				t.Fatalf("expected only team 2 to remain, got %v", teams)
			}
		})
	}
}
