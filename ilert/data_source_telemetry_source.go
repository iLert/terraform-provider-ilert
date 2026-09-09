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

func dataSourceTelemetrySource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTelemetrySourceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceTelemetrySourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	log.Printf("[DEBUG] Reading ilert telemetry source")

	searchName := d.Get("name").(string)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.SearchTelemetrySource(&ilert.SearchTelemetrySourceInput{TelemetrySourceName: &searchName})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for telemetry source with name '%s' to be read, error: %s", searchName, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a telemetry source with name: %s, error: %s", searchName, err.Error()))
		}

		found := resp.TelemetrySource

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("unable to locate any telemetry source with the name: %s", searchName),
			)
		}

		d.SetId(strconv.FormatInt(found.ID, 10))
		d.Set("name", found.Name)
		d.Set("type", found.Type)
		d.Set("description", found.Description)
		d.Set("service_name_prefix", found.ServiceNamePrefix)
		d.Set("service_name_suffix", found.ServiceNameSuffix)
		d.Set("status", found.Status)
		d.Set("integration_key", found.IntegrationKey)
		// the server-stamped discovery label is derived metadata, not something the user
		// set, so it stays out of the data source too
		labels := flattenLabelsAll(found.Labels)
		delete(labels, telemetrySourceDiscoveredByLabel)
		if err := d.Set("labels", labels); err != nil {
			return resource.NonRetryableError(fmt.Errorf("could not set labels of telemetry source with name: %s, error: %s", searchName, err.Error()))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
