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

func dataSourceCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCredentialRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sub_category": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCredentialRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert credential")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchCredential(&ilert.SearchCredentialInput{CredentialName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for credential with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a credential with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.Credential

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any credential with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("type", found.TypeHint)
		d.Set("host_url", found.HostURL)
		d.Set("category", found.Category)
		d.Set("sub_category", found.SubCategory)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
