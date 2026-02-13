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

func resourceService() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ilert.ServiceStatus.Operational,
				ValidateFunc: validation.StringInSlice(ilert.ServiceStatusAll, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"one_open_incident_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"show_uptime_history": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
		CreateContext: resourceServiceCreate,
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		Exists:        resourceServiceExists,
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

func buildService(d *schema.ResourceData) (*ilert.Service, error) {
	name := d.Get("name").(string)

	service := &ilert.Service{
		Name: name,
	}

	if val, ok := d.GetOk("status"); ok {
		service.Status = val.(string)
	}

	if val, ok := d.GetOk("description"); ok {
		service.Description = val.(string)
	}

	if val, ok := d.GetOk("one_open_incident_only"); ok {
		service.OneOpenIncidentOnly = val.(bool)
	}

	if val, ok := d.GetOk("show_uptime_history"); ok {
		service.ShowUptimeHistory = val.(bool)
	}

	if val, ok := d.GetOk("team"); ok {
		vL := val.([]any)
		tms := make([]ilert.TeamShort, 0)
		for _, m := range vL {
			v := m.(map[string]any)
			tm := ilert.TeamShort{
				ID: int64(v["id"].(int)),
			}
			if v["name"] != nil && v["name"].(string) != "" {
				tm.Name = v["name"].(string)
			}
			tms = append(tms, tm)
		}
		service.Teams = tms
	}

	return service, nil
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	service, err := buildService(d)
	if err != nil {
		log.Printf("[ERROR] Building service error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating service %s", service.Name)

	result := &ilert.CreateServiceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateService(&ilert.CreateServiceInput{Service: service})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert service error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert service error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.Service == nil {
		log.Printf("[ERROR] Creating ilert service error: empty response")
		return diag.Errorf("service response is empty")
	}

	d.SetId(strconv.FormatInt(result.Service.ID, 10))

	return resourceServiceRead(ctx, d, m)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	serviceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse service id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading service: %s", d.Id())
	result := &ilert.GetServiceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetService(&ilert.GetServiceInput{ServiceID: ilert.Int64(serviceID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing service %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an service with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert service error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.Service == nil {
		log.Printf("[ERROR] Reading ilert service error: empty response")
		return diag.Errorf("service response is empty")
	}

	err = transformServiceResource(result.Service, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	service, err := buildService(d)
	if err != nil {
		log.Printf("[ERROR] Building service error %s", err.Error())
		return diag.FromErr(err)
	}

	serviceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse service id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating service: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateService(&ilert.UpdateServiceInput{Service: service, ServiceID: ilert.Int64(serviceID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an service with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert service error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceServiceRead(ctx, d, m)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	serviceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse service id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting service: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteService(&ilert.DeleteServiceInput{ServiceID: ilert.Int64(serviceID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an service with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert service error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceServiceExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	serviceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse service id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading service: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetService(&ilert.GetServiceInput{ServiceID: ilert.Int64(serviceID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert service error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for service to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a service with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert service error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformServiceResource(service *ilert.Service, d *schema.ResourceData) error {
	d.Set("name", service.Name)
	d.Set("status", service.Status)
	d.Set("description", service.Description)
	d.Set("one_open_incident_only", service.OneOpenIncidentOnly)
	d.Set("show_uptime_history", service.ShowUptimeHistory)

	teams, err := flattenTeamShortList(service.Teams, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	return nil
}

func flattenTeamShortList(list []ilert.TeamShort, d *schema.ResourceData) ([]any, error) {
	if list == nil {
		return make([]any, 0), nil
	}
	results := make([]any, 0)
	if val, ok := d.GetOk("team"); ok && val != nil {
		vL := val.([]any)
		for i, item := range list {
			if vL != nil && i < len(vL) && vL[i] != nil {
				result := make(map[string]any)
				v := vL[i].(map[string]any)
				result["id"] = item.ID
				if item.Name != "" && v["name"] != nil && v["name"].(string) != "" {
					result["name"] = item.Name
				}
				results = append(results, result)
			}
		}
	} else if d.Id() == "" {
		for _, item := range list {
			result := map[string]any{
				"id": item.ID,
			}
			if item.Name != "" {
				result["name"] = item.Name
			}
			results = append(results, result)
		}
	}
	return results, nil
}
