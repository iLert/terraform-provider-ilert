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

func dataSourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEscalationPolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceEscalationPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert escalation policy")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchEscalationPolicy(&ilert.SearchEscalationPolicyInput{EscalationPolicyName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an escalation policy with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.EscalationPolicy

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any escalation policy with the name: %s", searchName),
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
