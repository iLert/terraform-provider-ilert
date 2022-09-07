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

func dataSourceAlertSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAlertSourceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"integration_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceAlertSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert alert source")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchAlertSource(&ilert.SearchAlertSourceInput{AlertSourceName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an alert source with ID %s", d.Id()))
		}

		found := resp.AlertSource

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any alert source with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("status", found.Status)
		d.Set("integration_key", found.IntegrationKey)
		d.Set("integration_url", found.IntegrationURL)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
