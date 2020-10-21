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

func dataSourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEscalationPolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceEscalationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert escalation policy")

	searchName := d.Get("name").(string)
	o := &ilert.GetEscalationPoliciesInput{}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := client.GetEscalationPolicies(o)
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		}

		var found *ilert.EscalationPolicy

		for _, escalationPolicy := range resp.EscalationPolicies {
			if escalationPolicy.Name == searchName {
				found = escalationPolicy
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any escalation policy with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)

		return nil
	})
}
