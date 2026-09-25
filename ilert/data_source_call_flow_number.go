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

func dataSourceCallFlowNumber() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCallFlowNumberRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"phone_number": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"number": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCallFlowNumberRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert call flow number")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchCallFlowNumber(&ilert.SearchCallFlowNumberInput{CallFlowNumberName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for call flow number with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a call flow number with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.CallFlowNumber

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any call flow number with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		// the API resolves a number matched by its phone number without a state, so this
		// stays empty rather than guessing one
		d.Set("state", found.State)
		if err := d.Set("phone_number", flattenCallFlowNumberPhoneNumber(found.PhoneNumber)); err != nil {
			return resource.NonRetryableError(fmt.Errorf("could not set phone number of call flow number with name: %s, error: %s", searchName, err.Error()))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenCallFlowNumberPhoneNumber(phoneNumber *ilert.PhoneNumber) []any {
	if phoneNumber == nil {
		return make([]any, 0)
	}

	return []any{
		map[string]any{
			"region_code": phoneNumber.RegionCode,
			"number":      phoneNumber.Number,
		},
	}
}
