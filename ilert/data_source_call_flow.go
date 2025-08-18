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

func dataSourceCallFlow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCallFlowRead,

        Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"assigned_number": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
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
				},
			},
		},
	}
}

func dataSourceCallFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert call flow")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchCallFlow(&ilert.SearchCallFlowInput{CallFlowName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for call flow with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a call flow with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.CallFlow

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any call flow with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)

		if found.AssignedNumber != nil {
			assigned := make(map[string]interface{})
			assigned["id"] = found.AssignedNumber.ID
			assigned["name"] = found.AssignedNumber.Name
			if found.AssignedNumber.PhoneNumber != nil {
				assigned["phone_number"] = []interface{}{
					map[string]interface{}{
						"region_code": found.AssignedNumber.PhoneNumber.RegionCode,
						"number":      found.AssignedNumber.PhoneNumber.Number,
					},
				}
			} else {
				assigned["phone_number"] = []interface{}{}
			}
			if err := d.Set("assigned_number", []interface{}{assigned}); err != nil {
				return resource.NonRetryableError(fmt.Errorf("error setting assigned_number: %s", err))
			}
		} else {
			if err := d.Set("assigned_number", []interface{}{}); err != nil {
				return resource.NonRetryableError(fmt.Errorf("error setting assigned_number: %s", err))
			}
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
