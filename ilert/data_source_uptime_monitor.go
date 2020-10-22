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
	o := &ilert.GetUptimeMonitorsInput{}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := client.GetUptimeMonitors(o)
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

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("status", found.Status)
		d.Set("embed_url", found.EmbedURL)
		d.Set("share_url", found.ShareURL)

		return nil
	})
}
