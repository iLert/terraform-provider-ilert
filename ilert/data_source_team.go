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
	"github.com/iLert/ilert-go"
)

func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"trigger_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert team")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.GetTeams(&ilert.GetTeamsInput{})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for team with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a team with ID %s", d.Id()))
		}

		var found *ilert.Team

		for _, team := range resp.Teams {
			if team.Name == searchName {
				found = team
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any team with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("visibility", found.Visibility)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
