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

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert user")

	searchEmail := d.Get("email").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchUser(&ilert.SearchUserInput{UserEmail: &searchEmail})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user with email '%s' to be read, error: %s", searchEmail, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an user with email: %s, error: %s", searchEmail, err.Error()))
		}

		found := resp.User

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any user with the email: %s", searchEmail),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("email", found.Email)
		d.Set("first_name", found.Username)
		d.Set("last_name", found.Username)
		d.Set("username", found.Username)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
