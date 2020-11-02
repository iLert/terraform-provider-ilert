package ilert

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go"
)

func dataSourceConnector() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConnectorRead,

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

func dataSourceConnectorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert connector")

	searchName := d.Get("name").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := client.GetConnectors(&ilert.GetConnectorsInput{})
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
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
				fmt.Errorf("Unable to locate any connector with the name: %s", searchName),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("type", found.Type)

		return nil
	})
}
