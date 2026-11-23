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

func dataSourceAction() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceActionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"entity_kind": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceActionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if err := requireEarlyAccess(meta, "ilert_action"); err != nil {
		return diag.FromErr(err)
	}

	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert action")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchAction(&ilert.SearchActionInput{ActionName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for action with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an action with name: %s, error: %s", searchName, actionsUnavailable(err).Error()))
		}

		found := resp.Action

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any action with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("entity_kind", found.EntityKind)
		d.Set("template_key", found.TemplateKey)
		d.Set("enabled", found.Enabled == nil || *found.Enabled)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
