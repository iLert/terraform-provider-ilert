package ilert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

func TestTeamMembersRoundTrip(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTeam().Schema, map[string]any{})
	team := &ilert.Team{
		Name:       "test-team",
		Visibility: ilert.TeamVisibility.Public,
		Members: []ilert.TeamMember{
			newTeamMember(200, ilert.TeamMemberRoles.Responder),
			newTeamMember(100, ilert.TeamMemberRoles.User),
		},
	}

	if err := transformTeamResource(team, d); err != nil {
		t.Fatalf("unexpected error transforming team: %v", err)
	}
	got, err := buildTeam(d)
	if err != nil {
		t.Fatalf("unexpected error building team: %v", err)
	}

	wantMembers := map[int64]string{
		100: ilert.TeamMemberRoles.User,
		200: ilert.TeamMemberRoles.Responder,
	}
	if len(got.Members) != len(wantMembers) {
		t.Fatalf("expected %d members, got %d", len(wantMembers), len(got.Members))
	}
	for _, member := range got.Members {
		if wantRole, ok := wantMembers[member.User.ID]; !ok || member.Role != wantRole {
			t.Fatalf("unexpected member %d with role %q", member.User.ID, member.Role)
		}
	}
}

func newTeamMember(userID int64, role string) ilert.TeamMember {
	return ilert.TeamMember{
		User: ilert.User{
			ID: userID,
		},
		Role: role,
	}
}
