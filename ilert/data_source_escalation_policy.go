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

func dataSourceEscalationPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert escalation policy")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchEscalationPolicy(&ilert.SearchEscalationPolicyInput{})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an escalation policy with ID %s", d.Id()))
		}

		var found *ilert.EscalationPolicy

		if resp.escalationPolicy.Name == searchName {
			found = resp.escalationPolicy
		}

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
