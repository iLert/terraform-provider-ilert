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

func dataSourceIncidentTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIncidentTemplateRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"summary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceIncidentTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert incident template")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchIncidentTemplate(&ilert.SearchIncidentTemplateInput{IncidentTemplateName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for incident template with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a incident template with name %s, error: %s", searchName, err.Error()))
		}

		found := resp.IncidentTemplate

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any incident template with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("status", found.Status)
		d.Set("summary", found.Summary)
		d.Set("message", found.Message)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
