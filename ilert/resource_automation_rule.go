package ilert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v3"
)

// Legacy API - please use alert-actions of type 'automation_rule' - for more information see https://api.ilert.com/api-docs/#tag/Alert-Actions/paths/~1alert-actions/post
func resourceAutomationRule() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "The resource automation rule is deprecated! Please use alert actions of type 'automation_rule' instead.",
		Schema: map[string]*schema.Schema{
			"alert_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.AlertTypeAll, false),
			},
			"resolve_incident": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"resolve_service": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"service_status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.ServiceStatusAll, false),
			},
			"template": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"service": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"alert_source": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"send_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CreateContext: resourceAutomationRuleCreate,
		ReadContext:   resourceAutomationRuleRead,
		UpdateContext: resourceAutomationRuleUpdate,
		DeleteContext: resourceAutomationRuleDelete,
		Exists:        resourceAutomationRuleExists,
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

func buildAutomationRule(d *schema.ResourceData) (*ilert.AutomationRule, error) {
	alertType := d.Get("alert_type").(string)
	serviceStatus := d.Get("service_status").(string)

	automationRule := &ilert.AutomationRule{
		AlertType:     alertType,
		ServiceStatus: serviceStatus,
	}

	if val, ok := d.GetOk("resolve_incident"); ok {
		automationRule.ResolveIncident = val.(bool)
	}

	if val, ok := d.GetOk("resolve_service"); ok {
		automationRule.ResolveService = val.(bool)
	}

	if val, ok := d.GetOk("template"); ok {
		if vL, ok := val.([]any); ok && len(vL) > 0 && vL[0] != nil {
			tmp := &ilert.IncidentTemplate{}
			if v, ok := vL[0].(map[string]any); ok && len(v) > 0 {
				tmp.ID = int64(v["id"].(int))
				if name, ok := v["name"].(string); ok && name != "" {
					tmp.Name = name
				}
				automationRule.Template = tmp
			} else {
				automationRule.Template = nil
			}
		}
	}

	if val, ok := d.GetOk("service"); ok {
		if vL, ok := val.([]any); ok && len(vL) > 0 && vL[0] != nil {
			svc := &ilert.Service{}
			if v, ok := vL[0].(map[string]any); ok && len(v) > 0 {
				svc.ID = int64(v["id"].(int))
				if name, ok := v["name"].(string); ok && name != "" {
					svc.Name = name
				}
				automationRule.Service = svc
			} else {
				automationRule.Service = nil
			}
		}
	}

	if val, ok := d.GetOk("alert_source"); ok {
		if vL, ok := val.([]any); ok && len(vL) > 0 && vL[0] != nil {
			asc := &ilert.AlertSource{}
			if v, ok := vL[0].(map[string]any); ok && len(v) > 0 {
				asc.ID = int64(v["id"].(int))
				if name, ok := v["name"].(string); ok && name != "" {
					asc.Name = name
				}
				automationRule.AlertSource = asc
			} else {
				automationRule.AlertSource = nil
			}
		}
	}

	if val, ok := d.GetOk("send_notification"); ok {
		automationRule.SendNotification = val.(bool)
	}

	return automationRule, nil
}

func resourceAutomationRuleCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	automationRule, err := buildAutomationRule(d)
	if err != nil {
		log.Printf("[ERROR] Building automation rule error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating automation rule")

	result := &ilert.CreateAutomationRuleOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateAutomationRule(&ilert.CreateAutomationRuleInput{AutomationRule: automationRule})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert automation rule error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for automation rule to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert automation rule error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.AutomationRule == nil {
		log.Printf("[ERROR] Creating ilert automation rule error: empty response ")
		return diag.Errorf("automation rule response is empty")
	}

	d.SetId(result.AutomationRule.ID)

	return resourceAutomationRuleRead(ctx, d, m)
}

func resourceAutomationRuleRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	automationRuleID := d.Id()
	log.Printf("[DEBUG] Reading automation rule: %s", d.Id())
	result := &ilert.GetAutomationRuleOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetAutomationRule(&ilert.GetAutomationRuleInput{AutomationRuleID: ilert.String(automationRuleID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing automation rule %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for automation rule with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an automation rule with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.AutomationRule == nil {
		log.Printf("[ERROR] Reading ilert automation rule error: empty response ")
		return diag.Errorf("automation rule response is empty")
	}

	d.Set("alert_type", result.AutomationRule.AlertType)
	d.Set("resolve_incident", result.AutomationRule.ResolveIncident)
	d.Set("resolve_service", result.AutomationRule.ResolveService)
	d.Set("service_status", result.AutomationRule.ServiceStatus)
	d.Set("send_notification", result.AutomationRule.SendNotification)

	if result.AutomationRule.Template != nil {
		d.Set("template", []any{
			map[string]any{
				"id":   result.AutomationRule.Template.ID,
				"name": result.AutomationRule.Template.Name,
			},
		})
	} else {
		d.Set("template", []any{})
	}

	service := make(map[string]any)
	service["id"] = result.AutomationRule.Service.ID
	service["name"] = result.AutomationRule.Service.Name
	d.Set("service", service)

	alertSource := make(map[string]any)
	alertSource["id"] = result.AutomationRule.AlertSource.ID
	alertSource["name"] = result.AutomationRule.AlertSource.Name
	d.Set("alert_source", alertSource)

	return nil
}

func resourceAutomationRuleUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	automationRule, err := buildAutomationRule(d)
	if err != nil {
		log.Printf("[ERROR] Building automation rule error %s", err.Error())
		return diag.FromErr(err)
	}

	automationRuleID := d.Id()
	if err != nil {
		log.Printf("[ERROR] Could not parse automation rule id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating automation rule: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateAutomationRule(&ilert.UpdateAutomationRuleInput{AutomationRule: automationRule, AutomationRuleID: ilert.String(automationRuleID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for automation rule with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an automation rule with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert automation rule error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceAutomationRuleRead(ctx, d, m)
}

func resourceAutomationRuleDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	automationRuleID := d.Id()
	log.Printf("[DEBUG] Deleting automation rule: %s", d.Id())
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := client.DeleteAutomationRule(&ilert.DeleteAutomationRuleInput{AutomationRuleID: ilert.String(automationRuleID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for automation rule with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an automation rule with ID %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert automation rule error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceAutomationRuleExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	automationRuleID := d.Id()
	log.Printf("[DEBUG] Reading automation rule: %s", d.Id())
	ctx := context.Background()
	result := false
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetAutomationRule(&ilert.GetAutomationRuleInput{AutomationRuleID: ilert.String(automationRuleID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert automation rule error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for automation rule to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = true
		return nil
	})

	if err != nil {
		return false, err
	}
	return result, nil
}
