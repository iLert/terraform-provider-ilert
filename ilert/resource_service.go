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
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"icon_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"link": {
				// TypeList, not TypeSet: the API preserves the order the links were
				// sent in and displays them in it, so the order is meaningful here.
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"href": {
							Type:     schema.TypeString,
							Required: true,
						},
						"text": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"public_status": {
				// the status shown on status pages, derived by the API from the service
				// status
				Type:     schema.TypeString,
				Computed: true,
			},
			"team": {
				// TypeSet, not TypeList: the API returns teams sorted by id, so an
				// ordered list produces a permanent diff whenever the config order
				// differs from that.
				Type:     schema.TypeSet,
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

	if val, ok := d.GetOk("alias"); ok {
		service.Alias = val.(string)
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

	if val, ok := d.GetOk("icon_url"); ok {
		service.IconUrl = val.(string)
	}

	if val, ok := d.GetOk("labels"); ok {
		labels := make(map[string]string)
		for k, v := range val.(map[string]any) {
			labels[k] = v.(string)
		}
		service.Labels = &labels
	} else if d.HasChange("labels") {
		// Every label removed. A nil map disappears from the payload and leaves the
		// labels untouched, so clearing them takes an explicit empty object. HasChange
		// keeps this to services whose labels were actually dropped from the config, so
		// an update on a configuration that never declared labels does not clear labels
		// assigned elsewhere.
		labels := make(map[string]string)
		service.Labels = &labels
	}

	// Links are always sent, unlike the labels and teams above. The API clears them when
	// the field is absent from the payload, verified against the API on 09.09.2026, so
	// omitting them protects nothing: it is the clear. Terraform therefore owns them
	// outright, and a link added in the web app is removed on the next apply.
	links := make([]ilert.ServiceLink, 0)
	if val, ok := d.GetOk("link"); ok {
		for _, m := range val.([]any) {
			v := m.(map[string]any)
			link := ilert.ServiceLink{
				Href: v["href"].(string),
			}
			if v["text"] != nil && v["text"].(string) != "" {
				link.Text = v["text"].(string)
			}
			links = append(links, link)
		}
	}
	service.Links = &links

	if val, ok := d.GetOk("team"); ok {
		vL := val.(*schema.Set).List()
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
	} else if d.HasChange("team") {
		// All team blocks removed. The API only clears teams on an explicit empty
		// array; an omitted or null field leaves them untouched. HasChange keeps
		// this to services whose teams were actually dropped from the config:
		// without it every update on a config that never declared a team block
		// would clear the teams assigned to it elsewhere.
		service.Teams = []ilert.TeamShort{}
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
		r, err := client.GetService(&ilert.GetServiceInput{ServiceID: ilert.Int64(serviceID), Include: serviceInclude()})
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
	d.Set("alias", service.Alias)
	d.Set("status", service.Status)
	d.Set("description", service.Description)
	d.Set("one_open_incident_only", service.OneOpenIncidentOnly)
	d.Set("show_uptime_history", service.ShowUptimeHistory)
	d.Set("icon_url", service.IconUrl)
	d.Set("public_status", service.PublicStatus)

	if err := d.Set("labels", flattenLabels(service.Labels, d, "labels")); err != nil {
		return fmt.Errorf("[ERROR] Error setting labels: %s", err.Error())
	}

	if err := d.Set("link", flattenServiceLinkList(service.Links)); err != nil {
		return fmt.Errorf("[ERROR] Error setting links: %s", err.Error())
	}

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
		// Match the config by team id, not by position: the API returns teams
		// sorted by id, which need not match the order they were declared in.
		requestedTeamNames := make(map[int64]bool)
		for _, m := range val.(*schema.Set).List() {
			v, ok := m.(map[string]any)
			if !ok || v["name"] == nil {
				continue
			}
			if vName, ok := v["name"].(string); ok && vName != "" {
				requestedTeamNames[int64(v["id"].(int))] = true
			}
		}
		for _, item := range list {
			result := make(map[string]any)
			result["id"] = item.ID
			// Means: if server response has a name set, and the user typed in a name too,
			// only then team name is stored in the terraform state
			if item.Name != "" && requestedTeamNames[item.ID] {
				result["name"] = item.Name
			}
			results = append(results, result)
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

// links and the public status are only part of a service response when they are
// requested, the same way the heartbeat monitor handles its integration url. Every read
// asks for both, so a link added in the web app reaches the state and shows up in the
// plan of a configuration that declares link blocks.
func serviceInclude() []*string {
	return []*string{&ilert.ServiceInclude.Links, &ilert.ServiceInclude.PublicStatus}
}

func flattenServiceLinkList(list *[]ilert.ServiceLink) []any {
	results := make([]any, 0)
	if list == nil {
		return results
	}
	for _, item := range *list {
		result := make(map[string]any)
		result["href"] = item.Href
		result["text"] = item.Text
		results = append(results, result)
	}
	return results
}
