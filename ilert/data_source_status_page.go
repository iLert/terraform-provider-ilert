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

func dataSourceStatusPage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceStatusPageRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceStatusPageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert status page")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchStatusPage(&ilert.SearchStatusPageInput{StatusPageName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a status page with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.StatusPage

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any service with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("status", found.Status)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
