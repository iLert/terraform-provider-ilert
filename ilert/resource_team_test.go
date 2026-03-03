package ilert

import (
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

func TestFlattenMembersListSorted_UsesConfigOrder(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTeam().Schema, map[string]any{
		"name": "test-team",
		"member": []any{
			map[string]any{
				"user": "200",
				"role": ilert.TeamMemberRoles.Responder,
			},
			map[string]any{
				"user": "100",
				"role": ilert.TeamMemberRoles.User,
			},
		},
	})

	serverMembers := []ilert.TeamMember{
		newTeamMember(100, ilert.TeamMemberRoles.User),
		newTeamMember(200, ilert.TeamMemberRoles.Responder),
	}

	got, err := flattenMembersListSorted(serverMembers, d)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	wantOrder := []string{"200", "100"}
	gotOrder := teamMemberUserOrder(got)
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("expected user order %v, got %v", wantOrder, gotOrder)
	}
}

func TestFlattenMembersListSorted_AppendsUnmatchedServerMember(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTeam().Schema, map[string]any{
		"name": "test-team",
		"member": []any{
			map[string]any{
				"user": "200",
				"role": ilert.TeamMemberRoles.Responder,
			},
		},
	})

	serverMembers := []ilert.TeamMember{
		newTeamMember(100, ilert.TeamMemberRoles.User),
		newTeamMember(200, ilert.TeamMemberRoles.Responder),
	}

	got, err := flattenMembersListSorted(serverMembers, d)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	gotOrder := teamMemberUserOrder(got)
	if len(gotOrder) != 2 {
		t.Fatalf("expected 2 users, got %d (%v)", len(gotOrder), gotOrder)
	}
	if gotOrder[0] != "200" {
		t.Fatalf("expected configured user first, got %v", gotOrder)
	}
	if !slices.Contains(gotOrder, "100") {
		t.Fatalf("expected unmatched server user 100 to be preserved, got %v", gotOrder)
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

func teamMemberUserOrder(members []any) []string {
	order := make([]string, 0, len(members))
	for _, member := range members {
		memberMap, ok := member.(map[string]any)
		if !ok {
			continue
		}
		user, ok := memberMap["user"].(string)
		if ok {
			order = append(order, user)
		}
	}
	return order
}
