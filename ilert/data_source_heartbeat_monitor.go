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
	"github.com/iLert/ilert-go/v3"
)

func dataSourceHeartbeatMonitor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHeartbeatMonitorRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
			},
			"integration_url": {
				Type:      schema.TypeString,
				Computed:  true,
			},
		},
	}
}

func dataSourceHeartbeatMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert heartbeat monitor")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchHeartbeatMonitor(&ilert.SearchHeartbeatMonitorInput{HeartbeatMonitorName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for heartbeat monitor with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a heartbeat monitor with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.HeartbeatMonitor

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any heartbeat monitor with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("state", found.State)
		d.Set("integration_key", found.IntegrationKey)
		d.Set("integration_url", found.IntegrationUrl)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
