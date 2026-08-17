package ilert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v3"
)

// Regression for the team block ordering bug: the API returns teams sorted by id,
// which need not match the order they were declared in. flattenTeamShortList is
// shared by 14 resources, so binding each configured name to its own team id here
// covers all of them.
func TestFlattenTeamShortList_BindsTeamNamesByIDNotPosition(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceService().Schema, map[string]any{
		"name": "test-service",
		// declared in the reverse of the order the API returns
		"team": []any{
			map[string]any{"id": 2, "name": "Team 2"},
			map[string]any{"id": 1, "name": "Team 1"},
		},
	})

	service := &ilertapi.Service{
		Name: "test-service",
		Teams: []ilertapi.TeamShort{
			{ID: 1, Name: "Team 1"},
			{ID: 2, Name: "Team 2"},
		},
	}

	if err := transformServiceResource(service, d); err != nil {
		t.Fatalf("unexpected error transforming service: %v", err)
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

// A team assigned on the server but not declared in the config must reach the
// state so the drift shows up in the plan. The read used to keep only as many
// teams as the config declared blocks, which hid it.
func TestFlattenTeamShortList_KeepsTeamsMissingFromConfig(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceService().Schema, map[string]any{
		"name": "test-service",
		"team": []any{
			map[string]any{"id": 1, "name": "Team 1"},
		},
	})

	service := &ilertapi.Service{
		Name: "test-service",
		Teams: []ilertapi.TeamShort{
			{ID: 1, Name: "Team 1"},
			{ID: 2, Name: "Team 2"},
		},
	}

	if err := transformServiceResource(service, d); err != nil {
		t.Fatalf("unexpected error transforming service: %v", err)
	}

	teams := d.Get("team").(*schema.Set).List()
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams in state, got %d", len(teams))
	}
}

// The name is only stored for teams the user named, so a config that declares a
// bare id keeps a bare id in state regardless of what the API reports.
func TestFlattenTeamShortList_OmitsNameWhenConfigDeclaresBareID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceService().Schema, map[string]any{
		"name": "test-service",
		"team": []any{
			map[string]any{"id": 2, "name": "Team 2"},
			map[string]any{"id": 1},
		},
	})

	service := &ilertapi.Service{
		Name: "test-service",
		Teams: []ilertapi.TeamShort{
			{ID: 1, Name: "Team 1"},
			{ID: 2, Name: "Team 2"},
		},
	}

	if err := transformServiceResource(service, d); err != nil {
		t.Fatalf("unexpected error transforming service: %v", err)
	}

	got := make(map[int]string)
	for _, item := range d.Get("team").(*schema.Set).List() {
		v := item.(map[string]any)
		got[v["id"].(int)] = v["name"].(string)
	}

	if got[1] != "" {
		t.Fatalf("expected no name for team 1, got %q", got[1])
	}
	if got[2] != "Team 2" {
		t.Fatalf("expected name 'Team 2' for team 2, got %q", got[2])
	}
}

func TestBuildService_ReadsTeamSet(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceService().Schema, map[string]any{
		"name": "test-service",
		"team": []any{
			map[string]any{"id": 2, "name": "Team 2"},
			map[string]any{"id": 1},
		},
	})

	service, err := buildService(d)
	if err != nil {
		t.Fatalf("unexpected error building service: %v", err)
	}

	got := make(map[int64]string)
	for _, tm := range service.Teams {
		got[tm.ID] = tm.Name
	}

	if len(got) != 2 || got[1] != "" || got[2] != "Team 2" {
		t.Fatalf("unexpected teams on the built service: %v", got)
	}
}
