package ilert

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ilert discontinued uptime monitoring, so the resource never calls the API: a nil meta here would
// panic on the first client call.
func TestResourceUptimeMonitor_Discontinued(t *testing.T) {
	r := resourceUptimeMonitor()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"name":              "checkout",
		"check_type":        "http",
		"escalation_policy": "1",
		"check_params":      []any{map[string]any{"url": "https://example.com"}},
	})

	diags := r.CreateContext(context.Background(), d, nil)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "discontinued uptime monitoring after 30.06.2024") {
		t.Fatalf("create = %v, want the discontinued error", diags)
	}
	if d.Id() != "" {
		t.Errorf("id = %q after a failed create, want none", d.Id())
	}

	// an uptime monitor left in the state is dropped on refresh, as the API's 404 did before
	d.SetId("123")
	if diags := r.ReadContext(context.Background(), d, nil); diags.HasError() || d.Id() != "" {
		t.Errorf("read = %v with id %q, want no error and the id cleared", diags, d.Id())
	}

	d.SetId("123")
	if diags := r.DeleteContext(context.Background(), d, nil); diags.HasError() || d.Id() != "" {
		t.Errorf("delete = %v with id %q, want no error and the id cleared", diags, d.Id())
	}
}

func TestDataSourceUptimeMonitor_Discontinued(t *testing.T) {
	r := dataSourceUptimeMonitor()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"name": "checkout"})

	diags := r.ReadContext(context.Background(), d, nil)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "ilert_uptime_monitor data source") {
		t.Fatalf("read = %v, want the discontinued error naming the data source", diags)
	}
}

// the warning Terraform prints at plan has to say the resource no longer works, not only that it is deprecated
func TestUptimeMonitor_DeprecationMessages(t *testing.T) {
	for name, r := range map[string]*schema.Resource{"resource": resourceUptimeMonitor(), "data source": dataSourceUptimeMonitor()} {
		if !strings.Contains(r.DeprecationMessage, "discontinued uptime monitoring") {
			t.Errorf("%s deprecation message = %q, want it to say uptime monitoring was discontinued", name, r.DeprecationMessage)
		}
	}
}
