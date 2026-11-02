package ilert

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v4"
)

// deleteStatusPageGroup runs the delete of group 2 on status page 1 against an
// API that answers with the given status and body.
func deleteStatusPageGroup(t *testing.T, status int, body string) (*schema.ResourceData, diag.Diagnostics) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/status-pages/1/groups/2" {
			t.Errorf("got %s %s, want DELETE /api/status-pages/1/groups/2", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	d := schema.TestResourceDataRaw(t, resourceStatusPageGroup().Schema, map[string]any{
		"name":        "test-group",
		"status_page": []any{map[string]any{"id": 1}},
	})
	d.SetId("2")

	client := ilert.NewClient(ilert.WithAPIEndpoint(server.URL), ilert.WithAPIToken("test"))
	return d, resourceStatusPageGroupDelete(context.Background(), d, client)
}

// The API deletes the groups of a status page along with it, so a destroy that
// deletes the page first finds its groups gone. That is the outcome the delete
// wants, so it is not an error. See iLert/engineering-tasks#2835.
func TestResourceStatusPageGroupDelete_TreatsNotFoundAsDeleted(t *testing.T) {
	d, diags := deleteStatusPageGroup(t, http.StatusNotFound, `{"status":404,"code":"ERROR","message":"Could not find status page with id 1"}`)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("id = %q, want it removed from the state", d.Id())
	}
}

func TestResourceStatusPageGroupDelete_ReportsOtherErrors(t *testing.T) {
	d, diags := deleteStatusPageGroup(t, http.StatusBadRequest, `{"status":400,"code":"ERROR","message":"bad request"}`)
	if !diags.HasError() {
		t.Fatal("expected an error for a 400")
	}
	if d.Id() != "2" {
		t.Errorf("id = %q, want it kept in the state", d.Id())
	}
}
