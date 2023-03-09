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

func dataSourceMetric() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMetricRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aggregation_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMetricRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert metric")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchMetric(&ilert.SearchMetricInput{MetricName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for metric with name '%s' to be read", searchName))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a metric with name: %s", searchName))
		}

		found := resp.Metric

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any metric with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("aggregation_type", found.AggregationType)
		d.Set("display_type", found.DisplayType)

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
