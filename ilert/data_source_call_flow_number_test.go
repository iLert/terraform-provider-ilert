package ilert

import (
	"testing"

	ilertapi "github.com/iLert/ilert-go/v3"
)

func TestFlattenCallFlowNumberPhoneNumber(t *testing.T) {
	// a number without a phone number has to flatten to an empty list rather than a
	// block of empty strings, otherwise the data source reports a number that is not there
	if got := flattenCallFlowNumberPhoneNumber(nil); len(got) != 0 {
		t.Errorf("flattenCallFlowNumberPhoneNumber(nil) = %v, want an empty list", got)
	}

	got := flattenCallFlowNumberPhoneNumber(&ilertapi.PhoneNumber{RegionCode: "DE", Number: "+4930123456"})
	if len(got) != 1 {
		t.Fatalf("flattenCallFlowNumberPhoneNumber() = %v, want a single block", got)
	}
	block := got[0].(map[string]any)
	if block["region_code"] != "DE" {
		t.Errorf("region_code = %v, want DE", block["region_code"])
	}
	if block["number"] != "+4930123456" {
		t.Errorf("number = %v, want +4930123456", block["number"])
	}
}

// Only the name is an input: state and the phone number come from the API, and the
// assigned call flow is deliberately absent because the lookup endpoint never returns it.
func TestDataSourceCallFlowNumber_Schema(t *testing.T) {
	dataSourceSchema := dataSourceCallFlowNumber().Schema

	name, ok := dataSourceSchema["name"]
	if !ok {
		t.Fatal("schema is missing \"name\"")
	}
	if !name.Required {
		t.Error("expected \"name\" to be required")
	}

	for _, attribute := range []string{"state", "phone_number"} {
		attributeSchema, ok := dataSourceSchema[attribute]
		if !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
		if !attributeSchema.Computed {
			t.Errorf("expected %q to be computed", attribute)
		}
	}

	if _, ok := dataSourceSchema["assigned_to"]; ok {
		t.Error("assigned_to is exposed, but the name lookup endpoint never returns it")
	}
}
