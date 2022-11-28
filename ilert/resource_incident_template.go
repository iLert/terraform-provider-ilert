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
	"github.com/iLert/ilert-go/v2"
)

func resourceIncidentTemplate() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"summary": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.IncidentStatusAll, false),
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"send_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"team": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
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
		},
		CreateContext: resourceIncidentTemplateCreate,
		ReadContext:   resourceIncidentTemplateRead,
		UpdateContext: resourceIncidentTemplateUpdate,
		DeleteContext: resourceIncidentTemplateDelete,
		Exists:        resourceIncidentTemplateExists,
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

func buildIncidentTemplate(d *schema.ResourceData) (*ilert.IncidentTemplate, error) {
	name := d.Get("name").(string)
	status := d.Get("status").(string)

	incidentTemplate := &ilert.IncidentTemplate{
		Name:   name,
		Status: status,
	}

	if val, ok := d.GetOk("summary"); ok {
		incidentTemplate.Summary = val.(string)
	}

	if val, ok := d.GetOk("message"); ok {
		incidentTemplate.Message = val.(string)
	}

	if val, ok := d.GetOk("send_notification"); ok {
		incidentTemplate.SendNotification = val.(bool)
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.([]interface{})
		tms := make([]ilert.TeamShort, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			tm := ilert.TeamShort{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				tm.Name = v["name"].(string)
			}
			tms = append(tms, tm)
		}
		incidentTemplate.Teams = tms
	}

	return incidentTemplate, nil
}

func resourceIncidentTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	incidentTemplate, err := buildIncidentTemplate(d)
	if err != nil {
		log.Printf("[ERROR] Building incident template error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating incident template %s", incidentTemplate.Name)

	result := &ilert.CreateIncidentTemplateOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateIncidentTemplate(&ilert.CreateIncidentTemplateInput{IncidentTemplate: incidentTemplate})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert incident template error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for incident template to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert incident template error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.IncidentTemplate == nil {
		log.Printf("[ERROR] Creating ilert incident template error: empty response ")
		return diag.Errorf("incident template response is empty")
	}

	d.SetId(strconv.FormatInt(result.IncidentTemplate.ID, 10))

	return resourceIncidentTemplateRead(ctx, d, m)
}

func resourceIncidentTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	incidentTemplateID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse incident template id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading incident template: %s", d.Id())
	result := &ilert.GetIncidentTemplateOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetIncidentTemplate(&ilert.GetIncidentTemplateInput{IncidentTemplateID: ilert.Int64(incidentTemplateID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing incident template %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for incident template with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an incident template with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.IncidentTemplate == nil {
		log.Printf("[ERROR] Reading ilert incident template error: empty response ")
		return diag.Errorf("incident template response is empty")
	}

	d.Set("name", result.IncidentTemplate.Name)
	d.Set("summary", result.IncidentTemplate.Summary)
	d.Set("status", result.IncidentTemplate.Status)
	d.Set("message", result.IncidentTemplate.Message)
	d.Set("send_notification", result.IncidentTemplate.SendNotification)

	teams, err := flattenTeamShortList(result.IncidentTemplate.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	return nil
}

func resourceIncidentTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	incidentTemplate, err := buildIncidentTemplate(d)
	if err != nil {
		log.Printf("[ERROR] Building incident template error %s", err.Error())
		return diag.FromErr(err)
	}

	incidentTemplateID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse incident template id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating incident template: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateIncidentTemplate(&ilert.UpdateIncidentTemplateInput{IncidentTemplate: incidentTemplate, IncidentTemplateID: ilert.Int64(incidentTemplateID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for incident template with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an incident template with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert incident template error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceIncidentTemplateRead(ctx, d, m)
}

func resourceIncidentTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	incidentTemplateID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse incident template id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting incident template: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteIncidentTemplate(&ilert.DeleteIncidentTemplateInput{IncidentTemplateID: ilert.Int64(incidentTemplateID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for incident template with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an incident template with ID %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert incident template error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceIncidentTemplateExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	incidentTemplateID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse incident template id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading incident template: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetIncidentTemplate(&ilert.GetIncidentTemplateInput{IncidentTemplateID: ilert.Int64(incidentTemplateID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert incident template error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for incident template to be read, error: %s", err.Error()))
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
