package ilert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v4"
)

// The names have to come from their own fields: read from the username, every
// user came back with the username as both the first and the last name.
func TestTransformUserDataSource(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceUser().Schema, map[string]any{
		"email": "jane.doe@example.com",
	})

	err := transformUserDataSource(&ilertapi.User{
		ID:        1,
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "jdoe",
		Email:     "jane.doe@example.com",
	}, d)
	if err != nil {
		t.Fatalf("unexpected error transforming user: %v", err)
	}

	want := map[string]string{
		"email":      "jane.doe@example.com",
		"first_name": "Jane",
		"last_name":  "Doe",
		"username":   "jdoe",
	}
	for attribute, value := range want {
		if got := d.Get(attribute); got != value {
			t.Errorf("%s = %v, want %q", attribute, got, value)
		}
	}
}
