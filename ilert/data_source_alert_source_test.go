package ilert

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAlertSourceDataSource_basic(t *testing.T) {
	rName := acctest.RandString(32)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertSourceDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ilert_alert_source.test", "name"),
					resource.TestCheckResourceAttrSet("data.ilert_alert_source.test", "status"),
				),
			},
		},
	})
}

func testAccAlertSourceDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_alert_source" "test" {
  name              = "%s"
  integration_type  = "API"
  escalation_policy = data.ilert_escalation_policy.default.id
}

data "ilert_alert_source" "test" {
  name = ilert_alert_source.test.name
}
`, rName)
}
