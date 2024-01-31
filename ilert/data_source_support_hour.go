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

func dataSourceSupportHour() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSupportHourRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceSupportHourRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert support hour")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchSupportHour(&ilert.SearchSupportHourInput{SupportHourName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for support hour with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a support hour with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.SupportHour

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any support hour with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
