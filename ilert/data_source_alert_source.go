package ilert

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/iLert/ilert-go"
)

func dataSourceAlertSource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlertSourceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAlertSourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert alert source")

	searchName := d.Get("name").(string)
	o := &ilert.GetAlertSourcesInput{}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := client.GetAlertSources(o)
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		}

		var found *ilert.AlertSource

		for _, alertSource := range resp.AlertSources {
			if alertSource.Name == searchName {
				found = alertSource
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any service with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)

		return nil
	})
}
