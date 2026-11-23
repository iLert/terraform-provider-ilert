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
	"github.com/iLert/ilert-go/v4"
)

func dataSourceAgent() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAgentRead,

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

func dataSourceAgentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if err := requireEarlyAccess(meta, "ilert_agent"); err != nil {
		return diag.FromErr(err)
	}

	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert agent")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchAgent(&ilert.SearchAgentInput{AgentName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for agent with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an agent with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.Agent

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any agent with the name: %s", searchName),
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
