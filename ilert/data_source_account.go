package ilert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAccountRead,

		Schema: map[string]*schema.Schema{
			"organization_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"timezone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enforce_mobile_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_admin_seat_purchase": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allow_ai": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ai_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"application_features": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subscription": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert account")

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.GetCurrentAccount(&ilert.GetCurrentAccountInput{})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for account to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read the account, error: %s", err.Error()))
		}

		found := resp.Account

		if found == nil {
			return resource.NonRetryableError(fmt.Errorf("unable to read the account of the configured api token"))
		}

		// the account id is a slug such as "acme", not a numeric id like every other entity
		d.SetId(found.ID)
		d.Set("organization_name", found.OrganizationName)
		d.Set("timezone", found.Timezone)
		d.Set("language", found.Language)
		d.Set("region", found.Region)
		d.Set("enforce_mobile_protection", found.EnforceMobileProtection)
		d.Set("allow_admin_seat_purchase", found.AllowAdminSeatPurchase)
		d.Set("allow_ai", found.AllowAI)
		d.Set("ai_mode", found.AiMode)
		if err := d.Set("application_features", found.ApplicationFeatures); err != nil {
			return resource.NonRetryableError(fmt.Errorf("could not set application features of the account, error: %s", err.Error()))
		}
		if err := d.Set("subscription", flattenAccountSubscription(found.Subscription)); err != nil {
			return resource.NonRetryableError(fmt.Errorf("could not set subscription of the account, error: %s", err.Error()))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenAccountSubscription(subscription *ilert.AccountSubscription) []any {
	if subscription == nil {
		return make([]any, 0)
	}

	return []any{
		map[string]any{
			"name":   subscription.Name,
			"status": subscription.Status,
		},
	}
}
