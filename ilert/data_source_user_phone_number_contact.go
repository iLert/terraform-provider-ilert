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

func dataSourceUserPhoneNumberContact() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserPhoneNumberContactRead,

		Schema: map[string]*schema.Schema{
			"target": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUserPhoneNumberContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert user phone number contact")

	searchTarget := d.Get("target").(string)

	user := d.Get("user").([]interface{})
	userId := int64(-1)
	if len(user) > 0 && user[0] != nil {
		usr := user[0].(map[string]interface{})
		id := int64(usr["id"].(int))
		userId = id
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchUserPhoneNumberContact(&ilert.SearchUserPhoneNumberContactInput{UserPhoneNumberContactTarget: &searchTarget, UserID: ilert.Int64(userId)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for user contact with phone number '%s' to be read", searchTarget))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a user contact with phone number: %s", searchTarget))
		}

		found := resp.UserPhoneNumberContact

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any user contact with the phone number: %s", searchTarget),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("target", found.Target)
		d.Set("status", found.Status)

		usr := make([]interface{}, 0)
		u := make(map[string]interface{}, 0)
		u["id"] = int(userId)
		usr = append(usr, u)
		d.Set("user", usr)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
