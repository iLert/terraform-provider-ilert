package ilert

import (
	"testing"

	ilertapi "github.com/iLert/ilert-go/v3"
)

// The account data source describes the caller's own account, so it takes no arguments:
// a required field here would ask the user for something the API never accepts.
func TestDataSourceAccount_TakesNoArguments(t *testing.T) {
	for name, attribute := range dataSourceAccount().Schema {
		if attribute.Required || attribute.Optional {
			t.Errorf("expected %q to be computed only, got required=%t optional=%t", name, attribute.Required, attribute.Optional)
		}
	}
}

func TestFlattenAccountSubscription(t *testing.T) {
	if got := flattenAccountSubscription(nil); len(got) != 0 {
		t.Errorf("flattenAccountSubscription(nil) = %v, want an empty list", got)
	}

	got := flattenAccountSubscription(&ilertapi.AccountSubscription{
		Name:   "ilert Scale",
		Status: ilertapi.SubscriptionStatus.Active,
	})
	if len(got) != 1 {
		t.Fatalf("flattenAccountSubscription() = %v, want a single block", got)
	}
	block := got[0].(map[string]any)
	if block["name"] != "ilert Scale" {
		t.Errorf("name = %v, want \"ilert Scale\"", block["name"])
	}
	if block["status"] != "ACTIVE" {
		t.Errorf("status = %v, want ACTIVE", block["status"])
	}
}
