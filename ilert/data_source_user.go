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

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading iLert user")

	searchEmail := d.Get("email").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, err := client.GetUsers(&ilert.GetUsersInput{})
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		}

		var found *ilert.User

		for _, user := range resp.Users {
			if user.Email == searchEmail {
				found = user
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any user with the email: %s", searchEmail),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("email", found.Email)
		d.Set("username", found.Username)

		return nil
	})
}
