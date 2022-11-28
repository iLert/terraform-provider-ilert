package ilert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v2"
)

// Legacy API - please use alert-actions - for more information see https://docs.ilert.com/rest-api/api-version-history#renaming-connections-to-alert-actions
func dataSourceConnection() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "The data source connection is deprecated! Please use data source alert action instead.",
		ReadContext:        dataSourceConnectionRead,

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

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert connection")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.GetConnections(&ilert.GetConnectionsInput{})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connection with name '%s' to be read", searchName))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a connection with name: %s", searchName))
		}

		var found *ilert.ConnectionOutput

		for _, connection := range resp.Connections {
			if connection.Name == searchName {
				found = connection
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any connection with the name: %s", searchName),
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
