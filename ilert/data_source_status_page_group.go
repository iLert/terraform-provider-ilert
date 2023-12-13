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

func dataSourceStatusPageGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceStatusPageGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status_page": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceStatusPageGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert status page group")

	searchName := d.Get("name").(string)
	spL := d.Get("status_page").([]interface{})
	statusPageID := int64(-1)
	if len(spL) > 0 && spL[0] != nil {
		sp := spL[0].(map[string]interface{})
		id := int64(sp["id"].(int))
		statusPageID = id
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchStatusPageGroup(&ilert.SearchStatusPageGroupInput{StatusPageGroupName: &searchName, StatusPageID: &statusPageID})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for status page group with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a status page group with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.StatusPageGroup

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any status page group with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)

		sp := make([]interface{}, 0)
		s := make(map[string]interface{}, 0)
		s["id"] = int(statusPageID)
		sp = append(sp, s)
		d.Set("status_page", sp)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
