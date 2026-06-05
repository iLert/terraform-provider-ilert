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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v3"
)

func dataSourceEventFlowIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEventFlowIntegrationRead,

		Schema: map[string]*schema.Schema{
			"event_flow_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"integration_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.AlertSourceIntegrationTypesAll, false),
			},
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"integration_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceEventFlowIntegrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ilert.Client)

	eventFlowID := int64(d.Get("event_flow_id").(int))
	integrationType := d.Get("integration_type").(string)

	log.Printf("[DEBUG] Reading ilert event flow integration for event flow %d and integration type %s", eventFlowID, integrationType)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		resp, err := client.GetEventFlowIntegrations(&ilert.GetEventFlowIntegrationsInput{EventFlowID: ilert.Int64(eventFlowID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for event flow integrations on event flow %d to be read, error: %s", eventFlowID, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not list event flow integrations for event flow %d, error: %s", eventFlowID, err.Error()))
		}

		for _, integration := range resp.EventFlowIntegrations {
			if integration == nil || integration.IntegrationType != integrationType {
				continue
			}
			d.SetId(strconv.FormatInt(integration.ID, 10))
			d.Set("event_flow_id", int(integration.EventFlowID))
			d.Set("integration_type", integration.IntegrationType)
			d.Set("integration_key", integration.IntegrationKey)
			d.Set("integration_url", integration.IntegrationURL)
			return nil
		}

		return resource.NonRetryableError(
			fmt.Errorf("unable to locate any event flow integration of type %s on event flow %d", integrationType, eventFlowID),
		)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
