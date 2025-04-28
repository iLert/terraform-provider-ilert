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

func resourceHeartbeatMonitor() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"interval_sec": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(25),
			},
			"alert_summary": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"alert_source": {
				Type:     schema.TypeList,
				Optional: true,
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
			"team": {
				Type:     schema.TypeList,
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
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CreateContext: resourceHeartbeatMonitorCreate,
		ReadContext:   resourceHeartbeatMonitorRead,
		UpdateContext: resourceHeartbeatMonitorUpdate,
		DeleteContext: resourceHeartbeatMonitorDelete,
		Exists:        resourceHeartbeatMonitorExists,
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

func buildHeartbeatMonitor(d *schema.ResourceData) (*ilert.HeartbeatMonitor, error) {
	name := d.Get("name").(string)
	intervalSec := d.Get("interval_sec").(int64)

	heartbeatMonitor := &ilert.HeartbeatMonitor{
		Name:        name,
		IntervalSec: intervalSec,
	}

	if val, ok := d.GetOk("alert_summary"); ok {
		heartbeatMonitor.AlertSummary = val.(string)
	}

	if val, ok := d.GetOk("alert_source"); ok {
		vL := val.([]interface{})
		v := vL[0].(map[string]interface{})
		as := ilert.AlertSource{
			ID: int64(v["id"].(int)),
		}
		if v["name"] != nil && v["name"].(string) != "" {
			as.Name = v["name"].(string)
		}
		heartbeatMonitor.AlertSource = &as
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
		heartbeatMonitor.Teams = tms
	}

	return heartbeatMonitor, nil
}

func resourceHeartbeatMonitorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	heartbeatMonitor, err := buildHeartbeatMonitor(d)
	if err != nil {
		log.Printf("[ERROR] Building heartbeat monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating heartbeat monitor %s", heartbeatMonitor.Name)

	result := &ilert.CreateHeartbeatMonitorOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		includes := []*string{ilert.String("integrationUrl")}
		r, err := client.CreateHeartbeatMonitor(&ilert.CreateHeartbeatMonitorInput{HeartbeatMonitor: heartbeatMonitor, Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert heartbeat monitor error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for heartbeat monitor to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a heartbeat monitor with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert heartbeat monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.HeartbeatMonitor == nil {
		log.Printf("[ERROR] Creating ilert heartbeat monitor error: empty response")
		return diag.Errorf("heartbeat monitor response is empty")
	}

	d.SetId(strconv.FormatInt(result.HeartbeatMonitor.ID, 10))

	return resourceHeartbeatMonitorRead(ctx, d, m)
}

func resourceHeartbeatMonitorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	heartbeatMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse heartbeat monitor id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading heartbeat monitor: %s", d.Id())
	result := &ilert.GetHeartbeatMonitorOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		includes := []*string{ilert.String("integrationUrl")}
		r, err := client.GetHeartbeatMonitor(&ilert.GetHeartbeatMonitorInput{HeartbeatMonitorID: ilert.Int64(heartbeatMonitorID), Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing heartbeat monitor %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for heartbeat monitor with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an heartbeat monitor with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert heartbeat monitor error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.HeartbeatMonitor == nil {
		log.Printf("[ERROR] Reading ilert heartbeat monitor error: empty response")
		return diag.Errorf("heartbeat monitor response is empty")
	}

	d.Set("name", result.HeartbeatMonitor.Name)
	d.Set("interval_sec", result.HeartbeatMonitor.IntervalSec)

	if _, ok := d.GetOk("alert_summary"); ok {
		d.Set("alert_summary", result.HeartbeatMonitor.AlertSummary)
	}

	if val, ok := d.GetOk("alert_source"); ok && val != nil {
		v := *result.HeartbeatMonitor.AlertSource
		as := val.([]interface{})[0].(map[string]interface{})
		alertSource := make(map[string]interface{})

		alertSource["id"] = v.ID
		if v.Name != "" && as["name"] != nil && as["name"].(string) != "" {
			alertSource["name"] = v.Name
		}

		d.Set("alert_source", alertSource)
	}

	teams, err := flattenTeamShortList(result.HeartbeatMonitor.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	d.Set("state", result.HeartbeatMonitor.State)
	d.Set("created_at", result.HeartbeatMonitor.CreatedAt)
	d.Set("updated_at", result.HeartbeatMonitor.UpdatedAt)
	d.Set("integration_key", result.HeartbeatMonitor.IntegrationKey)
	d.Set("integration_url", result.HeartbeatMonitor.IntegrationUrl)

	return nil
}

func resourceHeartbeatMonitorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	heartbeatMonitor, err := buildHeartbeatMonitor(d)
	if err != nil {
		log.Printf("[ERROR] Building heartbeat monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	// API expects integration key to be always set, even if not allowed to be set by user
	heartbeatMonitor.IntegrationKey = d.Get("integration_key").(string)

	heartbeatMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse heartbeat monitor id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating heartbeat monitor: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateHeartbeatMonitor(&ilert.UpdateHeartbeatMonitorInput{HeartbeatMonitor: heartbeatMonitor, HeartbeatMonitorID: ilert.Int64(heartbeatMonitorID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for heartbeat monitor with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an heartbeat monitor with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert heartbeat monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceHeartbeatMonitorRead(ctx, d, m)
}

func resourceHeartbeatMonitorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	heartbeatMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse heartbeat monitor id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting heartbeat monitor: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteHeartbeatMonitor(&ilert.DeleteHeartbeatMonitorInput{HeartbeatMonitorID: ilert.Int64(heartbeatMonitorID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for heartbeat monitor with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an heartbeat monitor with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert heartbeat monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceHeartbeatMonitorExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	heartbeatMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse heartbeat monitor id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading heartbeat monitor: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetHeartbeatMonitor(&ilert.GetHeartbeatMonitorInput{HeartbeatMonitorID: ilert.Int64(heartbeatMonitorID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert heartbeat monitor error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for heartbeat monitor to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a heartbeat monitor with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert heartbeat monitor error: %s", err.Error())
		return false, err
	}
	return result, nil
}
