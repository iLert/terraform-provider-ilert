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

func resourceEventFlowIntegration() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"integration_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.AlertSourceIntegrationTypesAll, false),
			},
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"event_flow_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"integration_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
		CreateContext: resourceEventFlowIntegrationCreate,
		ReadContext:   resourceEventFlowIntegrationRead,
		UpdateContext: resourceEventFlowIntegrationUpdate,
		DeleteContext: resourceEventFlowIntegrationDelete,
		Exists:        resourceEventFlowIntegrationExists,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func buildEventFlowIntegration(d *schema.ResourceData) *ilert.EventFlowIntegration {
	return &ilert.EventFlowIntegration{
		IntegrationType: d.Get("integration_type").(string),
		EventFlowID:     int64(d.Get("event_flow_id").(int)),
	}
}

func resourceEventFlowIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	integration := buildEventFlowIntegration(d)

	log.Printf("[INFO] Creating event flow integration of type %s for event flow %d", integration.IntegrationType, integration.EventFlowID)

	result := &ilert.CreateEventFlowIntegrationOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateEventFlowIntegration(&ilert.CreateEventFlowIntegrationInput{EventFlowIntegration: integration})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert event flow integration error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for event flow integration to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create an event flow integration, error: %s", err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert event flow integration error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.EventFlowIntegration == nil {
		log.Printf("[ERROR] Creating ilert event flow integration error: empty response")
		return diag.Errorf("event flow integration response is empty")
	}

	d.SetId(strconv.FormatInt(result.EventFlowIntegration.ID, 10))

	return resourceEventFlowIntegrationRead(ctx, d, m)
}

func resourceEventFlowIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	integrationID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse event flow integration id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading event flow integration: %s", d.Id())

	result := &ilert.GetEventFlowIntegrationOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetEventFlowIntegration(&ilert.GetEventFlowIntegrationInput{EventFlowIntegrationID: ilert.Int64(integrationID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing event flow integration %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for event flow integration with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an event flow integration with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert event flow integration error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.EventFlowIntegration == nil {
		log.Printf("[ERROR] Reading ilert event flow integration error: empty response")
		return diag.Errorf("event flow integration response is empty")
	}

	return diag.FromErr(transformEventFlowIntegrationResource(result.EventFlowIntegration, d))
}

func resourceEventFlowIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	integration := buildEventFlowIntegration(d)

	integrationID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse event flow integration id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating event flow integration: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateEventFlowIntegration(&ilert.UpdateEventFlowIntegrationInput{EventFlowIntegration: integration, EventFlowIntegrationID: ilert.Int64(integrationID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for event flow integration with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an event flow integration with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert event flow integration error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceEventFlowIntegrationRead(ctx, d, m)
}

func resourceEventFlowIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	integrationID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse event flow integration id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting event flow integration: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteEventFlowIntegration(&ilert.DeleteEventFlowIntegrationInput{EventFlowIntegrationID: ilert.Int64(integrationID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for event flow integration with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an event flow integration with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert event flow integration error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceEventFlowIntegrationExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	integrationID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse event flow integration id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading event flow integration: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetEventFlowIntegration(&ilert.GetEventFlowIntegrationInput{EventFlowIntegrationID: ilert.Int64(integrationID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert event flow integration error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for event flow integration to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an event flow integration with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert event flow integration error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformEventFlowIntegrationResource(integration *ilert.EventFlowIntegration, d *schema.ResourceData) error {
	d.Set("integration_type", integration.IntegrationType)
	d.Set("integration_key", integration.IntegrationKey)
	d.Set("event_flow_id", int(integration.EventFlowID))
	d.Set("integration_url", integration.IntegrationURL)
	return nil
}
