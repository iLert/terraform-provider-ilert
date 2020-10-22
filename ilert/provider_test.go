package ilert

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"ilert": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ILERT_ORGANIZATION"); v == "" {
		t.Fatal("ILERT_ORGANIZATION must be set for acceptance tests")
	}
	if v := os.Getenv("ILERT_USERNAME"); v == "" {
		t.Fatal("ILERT_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("ILERT_PASSWORD"); v == "" {
		t.Fatal("ILERT_PASSWORD must be set for acceptance tests")
	}
}
