package ilert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go"
)

func dataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectorRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert connector")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.GetConnectors(&ilert.GetConnectorsInput{})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connector with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a connector with ID %s", d.Id()))
		}

		var found *ilert.ConnectorOutput

		for _, connector := range resp.Connectors {
			if connector.Name == searchName {
				found = connector
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any connector with the name: %s", searchName),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("type", found.Type)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
