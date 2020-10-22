package ilert

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go"
)

func dataSourceUptimeMonitor() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUptimeMonitorRead,

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

func dataSourceUptimeMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert uptime monitor")

	searchName := d.Get("name").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := client.GetUptimeMonitors(&ilert.GetUptimeMonitorsInput{})
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
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
				fmt.Errorf("Unable to locate any uptime monitor with the name: %s", searchName),
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
}
