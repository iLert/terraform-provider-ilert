package ilert

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

func resourceAlertActionAlertSourceRelation() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"alert_action_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alert_source_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		CreateContext: resourceAlertActionAlertSourceRelationCreate,
		ReadContext:   resourceAlertActionAlertSourceRelationRead,
		DeleteContext: resourceAlertActionAlertSourceRelationDelete,
		Exists:        resourceAlertActionAlertSourceRelationExists,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAlertActionAlertSourceRelationImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAlertActionAlertSourceRelationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertActionID := d.Get("alert_action_id").(string)
	alertSourceIDStr := d.Get("alert_source_id").(string)
	alertSourceID, err := strconv.ParseInt(alertSourceIDStr, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid alert_source_id %q: %s", alertSourceIDStr, err.Error()))
	}

	log.Printf("[INFO] Attaching alert source %d to alert action %s", alertSourceID, alertActionID)

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.AddAlertSourceToAlertAction(&ilert.AddAlertSourceToAlertActionInput{
			AlertActionID: ilert.String(alertActionID),
			AlertSourceID: ilert.Int64(alertSourceID),
		})
		if err != nil {
			if isAlreadyAttachedError(err) {
				log.Printf("[WARN] Alert source %d already attached to alert action %s, treating as success", alertSourceID, alertActionID)
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Attaching alert source error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source %d to be attached to alert action %s, error: %s", alertSourceID, alertActionID, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not attach alert source %d to alert action %s, error: %s", alertSourceID, alertActionID, err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Attaching alert source to alert action error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%d", alertActionID, alertSourceID))

	return resourceAlertActionAlertSourceRelationRead(ctx, d, m)
}

func resourceAlertActionAlertSourceRelationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertActionID, alertSourceID, err := parseAlertActionAlertSourceRelationID(d.Id())
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}

	log.Printf("[DEBUG] Reading alert action / alert source relation: %s", d.Id())

	result := &ilert.GetAlertActionOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetAlertAction(&ilert.GetAlertActionInput{AlertActionID: ilert.String(alertActionID), Version: ilert.Int(2)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing alert action / alert source relation %s from state because alert action no longer exists", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action with id '%s' to be read, error: %s", alertActionID, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read alert action with id %s, error: %s", alertActionID, err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading alert action / alert source relation error: %s", err.Error())
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		return nil
	}

	if result == nil || result.AlertAction == nil {
		return diag.Errorf("alert action response is empty")
	}

	if !alertActionContainsSource(result.AlertAction, alertSourceID) {
		log.Printf("[WARN] Removing alert action / alert source relation %s from state because alert source is no longer attached", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("alert_action_id", alertActionID)
	d.Set("alert_source_id", strconv.FormatInt(alertSourceID, 10))

	return nil
}

func resourceAlertActionAlertSourceRelationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertActionID, alertSourceID, err := parseAlertActionAlertSourceRelationID(d.Id())
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}

	log.Printf("[DEBUG] Detaching alert source %d from alert action %s", alertSourceID, alertActionID)

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := client.RemoveAlertSourceFromAlertAction(&ilert.RemoveAlertSourceFromAlertActionInput{
			AlertActionID: ilert.String(alertActionID),
			AlertSourceID: ilert.Int64(alertSourceID),
		})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Alert action %s or alert source %d not found, treating detach as success", alertActionID, alertSourceID)
				return nil
			}
			if isAlreadyDetachedError(err) {
				log.Printf("[WARN] Alert source %d already detached from alert action %s, treating as success", alertSourceID, alertActionID)
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source %d to be detached from alert action %s, error: %s", alertSourceID, alertActionID, err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not detach alert source %d from alert action %s, error: %s", alertSourceID, alertActionID, err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Detaching alert source from alert action error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceAlertActionAlertSourceRelationExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	alertActionID, alertSourceID, err := parseAlertActionAlertSourceRelationID(d.Id())
	if err != nil {
		return false, unconvertibleIDErr(d.Id(), err)
	}

	log.Printf("[DEBUG] Checking alert action / alert source relation exists: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.GetAlertAction(&ilert.GetAlertActionInput{AlertActionID: ilert.String(alertActionID), Version: ilert.Int(2)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading alert action error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read alert action with id %s, error: %s", alertActionID, err.Error()))
		}
		result = r != nil && alertActionContainsSource(r.AlertAction, alertSourceID)
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading alert action / alert source relation error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func resourceAlertActionAlertSourceRelationImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	alertActionID, alertSourceID, err := parseAlertActionAlertSourceRelationID(d.Id())
	if err != nil {
		return nil, err
	}
	d.Set("alert_action_id", alertActionID)
	d.Set("alert_source_id", strconv.FormatInt(alertSourceID, 10))
	return []*schema.ResourceData{d}, nil
}

func parseAlertActionAlertSourceRelationID(id string) (alertActionID string, alertSourceID int64, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, fmt.Errorf("expected ID in the form '<alert_action_id>/<alert_source_id>', got %q", id)
	}
	alertSourceID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid alert_source_id %q in ID %q: %s", parts[1], id, err.Error())
	}
	return parts[0], alertSourceID, nil
}

func isAlreadyAttachedError(err error) bool {
	bre, ok := err.(*ilert.BadRequestAPIError)
	if !ok {
		return false
	}
	return strings.Contains(bre.Message, "already attached to this alert source")
}

func isAlreadyDetachedError(err error) bool {
	bre, ok := err.(*ilert.BadRequestAPIError)
	if !ok {
		return false
	}
	return strings.Contains(bre.Message, "not attached to this alert source")
}

func alertActionContainsSource(alertAction *ilert.AlertActionOutput, alertSourceID int64) bool {
	if alertAction == nil {
		return false
	}
	if alertAction.AlertSources != nil {
		for _, s := range *alertAction.AlertSources {
			if s.ID == alertSourceID {
				return true
			}
		}
	}
	return slices.Contains(alertAction.AlertSourceIDs, alertSourceID)
}
