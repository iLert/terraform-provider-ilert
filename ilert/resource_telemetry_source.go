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

func resourceTelemetrySource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"type": {
				// the telemetry protocol cannot be changed after the source was created
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.TelemetrySourceTypeAll, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"service_name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"service_name_suffix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"team": {
				// TypeSet, not TypeList: the API returns teams sorted by id, so an
				// ordered list produces a permanent diff whenever the config order
				// differs from that.
				Type:     schema.TypeSet,
				Optional: true,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_key": {
				// the credential the collector authenticates with. The API omits it for
				// users without update permission on the telemetry source.
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CreateContext: resourceTelemetrySourceCreate,
		ReadContext:   resourceTelemetrySourceRead,
		UpdateContext: resourceTelemetrySourceUpdate,
		DeleteContext: resourceTelemetrySourceDelete,
		Exists:        resourceTelemetrySourceExists,
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

// teams are only part of a telemetry source response when they are requested, the same
// way the heartbeat monitor handles its integration url. Every call asks for them, so a
// team assigned on the server always reaches the state and shows up in the plan.
func telemetrySourceInclude() []*string {
	return []*string{&ilert.TelemetrySourceInclude.Teams}
}

// The API stamps this label onto every telemetry source, derived from the source name, and
// documents it as read-only metadata that clients should omit when sending labels back. It
// is therefore kept out of both the payload and the state: leaving it in the state would
// make every plan propose deleting a label the server immediately writes back, and it
// changes on its own whenever the source is renamed.
const telemetrySourceDiscoveredByLabel = "ilert.com/discovered-by"

func buildTelemetrySource(d *schema.ResourceData) (*ilert.TelemetrySource, error) {
	name := d.Get("name").(string)
	sourceType := d.Get("type").(string)

	telemetrySource := &ilert.TelemetrySource{
		Name: name,
		Type: sourceType,
	}

	if val, ok := d.GetOk("description"); ok {
		telemetrySource.Description = val.(string)
	}

	if val, ok := d.GetOk("service_name_prefix"); ok {
		telemetrySource.ServiceNamePrefix = val.(string)
	}

	if val, ok := d.GetOk("service_name_suffix"); ok {
		telemetrySource.ServiceNameSuffix = val.(string)
	}

	if val, ok := d.GetOk("labels"); ok {
		labels := make(map[string]string)
		for k, v := range val.(map[string]any) {
			if k == telemetrySourceDiscoveredByLabel {
				continue
			}
			labels[k] = v.(string)
		}
		telemetrySource.Labels = &labels
	} else if d.HasChange("labels") {
		// Every label removed. An omitted field leaves the labels untouched, so clearing
		// them takes an explicit empty object. HasChange keeps this to sources whose
		// labels were actually dropped from the config, so an update on a configuration
		// that never declared labels does not clear labels set elsewhere.
		labels := make(map[string]string)
		telemetrySource.Labels = &labels
	}

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
		telemetrySource.Teams = &tms
	} else if d.HasChange("team") {
		// All team blocks removed, same reasoning as the labels above.
		tms := make([]ilert.TeamShort, 0)
		telemetrySource.Teams = &tms
	}

	return telemetrySource, nil
}

func resourceTelemetrySourceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	telemetrySource, err := buildTelemetrySource(d)
	if err != nil {
		log.Printf("[ERROR] Building telemetry source error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating telemetry source %s", telemetrySource.Name)

	result := &ilert.CreateTelemetrySourceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateTelemetrySource(&ilert.CreateTelemetrySourceInput{TelemetrySource: telemetrySource, Include: telemetrySourceInclude()})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert telemetry source error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for telemetry source to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert telemetry source error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.TelemetrySource == nil {
		log.Printf("[ERROR] Creating ilert telemetry source error: empty response")
		return diag.Errorf("telemetry source response is empty")
	}

	d.SetId(strconv.FormatInt(result.TelemetrySource.ID, 10))

	return resourceTelemetrySourceRead(ctx, d, m)
}

func resourceTelemetrySourceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	telemetrySourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse telemetry source id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading telemetry source: %s", d.Id())
	result := &ilert.GetTelemetrySourceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetTelemetrySource(&ilert.GetTelemetrySourceInput{TelemetrySourceID: ilert.Int64(telemetrySourceID), Include: telemetrySourceInclude()})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing telemetry source %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for telemetry source with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a telemetry source with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert telemetry source error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.TelemetrySource == nil {
		log.Printf("[ERROR] Reading ilert telemetry source error: empty response")
		return diag.Errorf("telemetry source response is empty")
	}

	err = transformTelemetrySourceResource(result.TelemetrySource, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTelemetrySourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	telemetrySource, err := buildTelemetrySource(d)
	if err != nil {
		log.Printf("[ERROR] Building telemetry source error %s", err.Error())
		return diag.FromErr(err)
	}

	telemetrySourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse telemetry source id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating telemetry source: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateTelemetrySource(&ilert.UpdateTelemetrySourceInput{TelemetrySource: telemetrySource, TelemetrySourceID: ilert.Int64(telemetrySourceID), Include: telemetrySourceInclude()})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for telemetry source with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update a telemetry source with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert telemetry source error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceTelemetrySourceRead(ctx, d, m)
}

func resourceTelemetrySourceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	telemetrySourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse telemetry source id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting telemetry source: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteTelemetrySource(&ilert.DeleteTelemetrySourceInput{TelemetrySourceID: ilert.Int64(telemetrySourceID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for telemetry source with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete a telemetry source with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert telemetry source error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceTelemetrySourceExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	telemetrySourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse telemetry source id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading telemetry source: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetTelemetrySource(&ilert.GetTelemetrySourceInput{TelemetrySourceID: ilert.Int64(telemetrySourceID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert telemetry source error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for telemetry source to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a telemetry source with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert telemetry source error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformTelemetrySourceResource(telemetrySource *ilert.TelemetrySource, d *schema.ResourceData) error {
	d.Set("name", telemetrySource.Name)
	d.Set("type", telemetrySource.Type)
	d.Set("description", telemetrySource.Description)
	d.Set("service_name_prefix", telemetrySource.ServiceNamePrefix)
	d.Set("service_name_suffix", telemetrySource.ServiceNameSuffix)
	d.Set("status", telemetrySource.Status)
	d.Set("created_at", telemetrySource.CreatedAt)
	d.Set("updated_at", telemetrySource.UpdatedAt)

	// The API omits the integration key for users without update permission on the
	// telemetry source. Overwriting the state with an empty string in that case would
	// drop a key the resource legitimately holds, so it is only set when returned.
	if telemetrySource.IntegrationKey != "" {
		d.Set("integration_key", telemetrySource.IntegrationKey)
	}

	labels := flattenLabels(telemetrySource.Labels, d, "labels")
	delete(labels, telemetrySourceDiscoveredByLabel)
	if err := d.Set("labels", labels); err != nil {
		return fmt.Errorf("[ERROR] Error setting labels: %s", err.Error())
	}

	teams, err := flattenTeamShortList(derefTeamShortList(telemetrySource.Teams), d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	return nil
}
