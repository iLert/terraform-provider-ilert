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

func dataSourceStatusPageGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceStatusPageGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status_page_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func dataSourceStatusPageGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert status page group")

	searchName := d.Get("name").(string)
	statusPageID := int64(d.Get("status_page_id").(int))

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchStatusPageGroup(&ilert.SearchStatusPageGroupInput{StatusPageGroupName: &searchName, StatusPageID: &statusPageID})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page group with name '%s' to be read", searchName))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a status page group with name: %s", searchName))
		}

		found := resp.StatusPageGroup

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any status page group with the name: %s", searchName),
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
