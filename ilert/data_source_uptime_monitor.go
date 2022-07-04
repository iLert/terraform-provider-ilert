package ilert

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v2"
)

func dataSourceUptimeMonitor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUptimeMonitorRead,

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

func dataSourceUptimeMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert uptime monitor")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.GetUptimeMonitors(&ilert.GetUptimeMonitorsInput{})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for uptime monitor with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an uptime monitor with ID %s", d.Id()))
		}

		var found *ilert.UptimeMonitor

		for _, uptimeMonitor := range resp.UptimeMonitors {
			if uptimeMonitor.Name == searchName {
				found = uptimeMonitor
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any uptime monitor with the name: %s", searchName),
			)
		}

		// Fetch uptime monitor again because the list route does not return report urls
		result, err := client.GetUptimeMonitor(&ilert.GetUptimeMonitorInput{UptimeMonitorID: ilert.Int64(found.ID)})
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		}

		d.SetId(strconv.FormatInt(result.UptimeMonitor.ID, 10))
		d.Set("name", result.UptimeMonitor.Name)
		d.Set("status", result.UptimeMonitor.Status)
		d.Set("embed_url", result.UptimeMonitor.EmbedURL)
		d.Set("share_url", result.UptimeMonitor.ShareURL)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
