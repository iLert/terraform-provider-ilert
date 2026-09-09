package ilert

import (
	"encoding/json"
	"strings"
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

// Links and labels follow the same contract as the team blocks: removing every block has
// to reach the API as an explicit empty array or object, or they stay on the server
// while Terraform reports the apply as successful.
func TestBuildService_RemovingAllLinksAndLabelsSendsEmptyValues(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceService(),
		map[string]any{
			"name":   "test-service",
			"labels": map[string]any{"env": "production"},
			"link": []any{
				map[string]any{"href": "https://example.com", "text": "Runbook"},
			},
		},
		map[string]any{"name": "test-service"},
	)

	service, err := buildService(d)
	if err != nil {
		t.Fatalf("unexpected error building service: %v", err)
	}
	if service.Labels == nil || len(*service.Labels) != 0 {
		t.Fatalf("Labels = %v, want a non-nil empty map", service.Labels)
	}
	if service.Links == nil || len(*service.Links) != 0 {
		t.Fatalf("Links = %v, want a non-nil empty slice", service.Links)
	}

	encoded, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("unexpected error marshalling payload: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":{}`) || !strings.Contains(string(encoded), `"links":[]`) {
		t.Fatalf("expected payload to clear both labels and links, got %s", encoded)
	}
}

// Labels are the half of this the API really does leave alone: a config that never
// declared them omits the field, and the labels assigned in the web app survive. Links
// are not, so they are always sent — see the test below.
func TestBuildService_KeepsLabelsWhenNeverDeclared(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceService(),
		map[string]any{"name": "test-service"},
		map[string]any{"name": "test-service-renamed"},
	)

	service, err := buildService(d)
	if err != nil {
		t.Fatalf("unexpected error building service: %v", err)
	}
	if service.Labels != nil {
		t.Fatalf("Labels = %v, want the field omitted", *service.Labels)
	}

	encoded, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("unexpected error marshalling payload: %v", err)
	}
	if strings.Contains(string(encoded), `"labels"`) {
		t.Fatalf("expected payload to omit labels entirely, got %s", encoded)
	}
}

// Verified against the API on 09.09.2026: a PUT that omits "links" clears the links on
// the server, exactly as an explicit empty array does. Omitting them therefore protects
// nothing, so the payload always carries them and Terraform owns them outright. A build
// that left the field out would silently wipe links instead of preserving them, which is
// what this provider claimed to do before the behaviour was measured.
func TestBuildService_AlwaysSendsLinks(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceService(),
		map[string]any{"name": "test-service"},
		map[string]any{"name": "test-service-renamed"},
	)

	service, err := buildService(d)
	if err != nil {
		t.Fatalf("unexpected error building service: %v", err)
	}
	if service.Links == nil {
		t.Fatalf("Links = nil, want a non-nil empty slice so the payload always carries them")
	}

	encoded, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("unexpected error marshalling payload: %v", err)
	}
	if !strings.Contains(string(encoded), `"links":[]`) {
		t.Fatalf("expected payload to contain \"links\":[], got %s", encoded)
	}
}

// Links keep the order they were declared in: the API preserves it and displays them in
// it, so a list that came back reordered would be a real change, not noise.
func TestBuildService_KeepsLinkOrder(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceService().Schema, map[string]any{
		"name": "test-service",
		"link": []any{
			map[string]any{"href": "https://example.com/first", "text": "First"},
			map[string]any{"href": "https://example.com/second", "text": "Second"},
		},
	})

	service, err := buildService(d)
	if err != nil {
		t.Fatalf("unexpected error building service: %v", err)
	}
	if service.Links == nil || len(*service.Links) != 2 {
		t.Fatalf("Links = %v, want the two declared links", service.Links)
	}
	if (*service.Links)[0].Href != "https://example.com/first" || (*service.Links)[1].Text != "Second" {
		t.Fatalf("Links = %+v, want them in the declared order", *service.Links)
	}
}

// The new service fields reach the state on a read, and a service without links or
// labels reads back as empty rather than null so a config declaring none has no diff.
func TestTransformService_SetsNewFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceService().Schema, map[string]any{
		"name":   "test-service",
		"labels": map[string]any{"env": "declared"},
	})
	d.SetId("1")

	labels := map[string]string{"env": "production"}
	links := []ilertapi.ServiceLink{{Href: "https://example.com", Text: "Runbook"}}
	if err := transformServiceResource(&ilertapi.Service{
		Name:         "test-service",
		IconUrl:      "https://example.com/icon.png",
		PublicStatus: "OPERATIONAL",
		Labels:       &labels,
		Links:        &links,
	}, d); err != nil {
		t.Fatalf("unexpected error transforming service: %v", err)
	}

	if got := d.Get("icon_url").(string); got != "https://example.com/icon.png" {
		t.Errorf("icon_url = %q, want the url from the API", got)
	}
	if got := d.Get("public_status").(string); got != "OPERATIONAL" {
		t.Errorf("public_status = %q, want OPERATIONAL", got)
	}
	if got := d.Get("labels").(map[string]any); got["env"] != "production" {
		t.Errorf("labels = %v, want the label from the API", got)
	}
	gotLinks := d.Get("link").([]any)
	if len(gotLinks) != 1 {
		t.Fatalf("link = %v, want the single link from the API", gotLinks)
	}
	if v := gotLinks[0].(map[string]any); v["href"] != "https://example.com" || v["text"] != "Runbook" {
		t.Errorf("link[0] = %v, want href and text from the API", v)
	}

	if err := transformServiceResource(&ilertapi.Service{Name: "test-service"}, d); err != nil {
		t.Fatalf("unexpected error transforming service: %v", err)
	}
	if got := d.Get("link").([]any); len(got) != 0 {
		t.Errorf("link = %v, want empty when the API returns none", got)
	}
	if got := d.Get("labels").(map[string]any); len(got) != 0 {
		t.Errorf("labels = %v, want empty when the API returns none", got)
	}
}

// A service whose configuration never declared labels leaves them unmanaged, so an
// unrelated apply does not take over labels set in the web app. Same rule as the team
// blocks, and the reason the API's "absent means untouched" behaviour is worth having.
func TestTransformService_LeavesUndeclaredLabelsUnmanaged(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceService().Schema, map[string]any{
		"name": "test-service",
	})
	d.SetId("1")

	labels := map[string]string{"owner": "webapp-team"}
	if err := transformServiceResource(&ilertapi.Service{Name: "test-service", Labels: &labels}, d); err != nil {
		t.Fatalf("unexpected error transforming service: %v", err)
	}

	if got := d.Get("labels").(map[string]any); len(got) != 0 {
		t.Fatalf("labels = %v, want them left out of state so they stay unmanaged", got)
	}
}
