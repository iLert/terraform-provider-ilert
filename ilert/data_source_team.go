package ilert

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go"
)

func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTeamRead,

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

func dataSourceTeamRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert team")

	searchName := d.Get("name").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := client.GetTeams(&ilert.GetTeamsInput{})
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
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
				fmt.Errorf("Unable to locate any team with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("visibility", found.Visibility)

		return nil
	})
}
