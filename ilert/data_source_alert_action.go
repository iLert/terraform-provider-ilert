package ilert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

func dataSourceAlertAction() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAlertActionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"trigger_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAlertActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert alert action")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchAlertAction(&ilert.SearchAlertActionInput{AlertActionName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action with name '%s' to be read", searchName))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a alert action with name: %s", searchName))
		}

		found := resp.AlertAction

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any alert action with the name: %s", searchName),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("trigger_mode", found.TriggerMode)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
