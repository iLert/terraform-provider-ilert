package ilert

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ilert discontinued uptime monitoring after 30.06.2024, so there is no uptime monitor left to look up.
// The data source keeps its schema so existing configurations still parse, and fails with an error
// that says so instead of the 404 the API answers with.
func dataSourceUptimeMonitor() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: fmt.Sprintf(uptimeMonitorDiscontinued, "ilert_uptime_monitor data source"),
		ReadContext:        dataSourceUptimeMonitorRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"embed_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUptimeMonitorRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Errorf(uptimeMonitorDiscontinued, "ilert_uptime_monitor data source")
}
