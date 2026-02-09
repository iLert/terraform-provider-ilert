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

func dataSourceMetricDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMetricDataSourceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMetricDataSourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert metric data source")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchMetricDataSource(&ilert.SearchMetricDataSourceInput{MetricDataSourceName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric data source with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a metric data source with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.MetricDataSource

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any metric data source with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("type", found.Type)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
