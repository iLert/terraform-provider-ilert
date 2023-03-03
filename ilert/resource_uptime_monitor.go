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

func resourceUptimeMonitor() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "EU",
				ValidateFunc: validation.StringInSlice([]string{
					"EU",
					"US",
				}, false),
			},
			"check_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"http",
					"ping",
					"tcp",
					"udp",
					"ssl",
				}, false),
			},
			"check_params": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"check_params.url"},
						},
						"port": {
							Type:          schema.TypeInt,
							Optional:      true,
							ConflictsWith: []string{"check_params.url"},
						},
						"url": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"check_params.host"},
						},
						"response_keywords": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"alert_before_sec": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"alert_on_fingerprint_change": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"interval_sec": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
				ValidateFunc: validation.IntInSlice([]int{
					1 * 60,
					5 * 60,
					10 * 60,
					15 * 60,
					30 * 60,
					60 * 60,
				}),
			},
			"timeout_ms": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30000,
				ValidateFunc: validation.IntBetween(1000, 60000),
			},
			"create_incident_after_failed_checks": { // @deprecated
				Deprecated: "The field create_incident_after_failed_checks is deprecated! Please use create_alert_after_failed_checks instead.",
				Type:       schema.TypeInt,
				Optional:   true,
				Default:    0,
			},
			"create_alert_after_failed_checks": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 12),
			},
			"escalation_policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"paused": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"embed_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CreateContext: resourceUptimeMonitorCreate,
		ReadContext:   resourceUptimeMonitorRead,
		UpdateContext: resourceUptimeMonitorUpdate,
		DeleteContext: resourceUptimeMonitorDelete,
		Exists:        resourceUptimeMonitorExists,
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

func buildUptimeMonitor(d *schema.ResourceData) (*ilert.UptimeMonitor, error) {
	name := d.Get("name").(string)
	region := d.Get("region").(string)
	checkType := d.Get("check_type").(string)
	escalationPolicyID, err := strconv.ParseInt(d.Get("escalation_policy").(string), 10, 64)
	if err != nil {
		return nil, unconvertibleIDErr(d.Id(), err)
	}

	uptimeMonitor := &ilert.UptimeMonitor{
		Name:      name,
		Region:    region,
		CheckType: checkType,
		EscalationPolicy: &ilert.EscalationPolicy{
			ID: escalationPolicyID,
		},
	}

	if val, ok := d.GetOk("check_params"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			checkParams := ilert.UptimeMonitorCheckParams{}
			if v["url"].(string) != "" {
				checkParams.URL = v["url"].(string)
			} else if v["host"].(string) != "" {
				checkParams.Host = v["host"].(string)
				if v["port"].(int) > 0 {
					checkParams.Port = v["port"].(int)
				}
			}
			if v["response_keywords"].([]interface{}) != nil {
				for _, keyword := range v["response_keywords"].([]interface{}) {
					checkParams.ResponseKeywords = append(checkParams.ResponseKeywords, keyword.(string))
				}
			}
			if v["alert_before_sec"].(int) > 0 {
				checkParams.AlertBeforeSec = v["alert_before_sec"].(int)
			}
			if v["alert_on_fingerprint_change"].(bool) {
				checkParams.AlertOnFingerprintChange = v["alert_on_fingerprint_change"].(bool)
			}
			uptimeMonitor.CheckParams = checkParams
		}
	}

	if val, ok := d.GetOk("interval_sec"); ok {
		intervalSec := val.(int)
		uptimeMonitor.IntervalSec = intervalSec
	}

	if val, ok := d.GetOk("timeout_ms"); ok {
		timeoutMs := val.(int)
		uptimeMonitor.TimeoutMs = timeoutMs
	}

	if val, ok := d.GetOk("create_incident_after_failed_checks"); ok {
		createIncidentAfterFailedChecks := val.(int)
		uptimeMonitor.CreateIncidentAfterFailedChecks = createIncidentAfterFailedChecks
	}

	if val, ok := d.GetOk("create_alert_after_failed_checks"); ok {
		createAlertAfterFailedChecks := val.(int)
		uptimeMonitor.CreateAlertAfterFailedChecks = createAlertAfterFailedChecks
	}

	if val, ok := d.GetOk("paused"); ok {
		paused := val.(bool)
		uptimeMonitor.Paused = paused
	}

	return uptimeMonitor, nil
}

func resourceUptimeMonitorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	uptimeMonitor, err := buildUptimeMonitor(d)
	if err != nil {
		log.Printf("[ERROR] Building uptime monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating uptime monitor %s", uptimeMonitor.Name)

	result := &ilert.CreateUptimeMonitorOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateUptimeMonitor(&ilert.CreateUptimeMonitorInput{UptimeMonitor: uptimeMonitor})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert uptime monitor error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for uptime monitor to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert uptime monitor error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.UptimeMonitor == nil {
		log.Printf("[ERROR] Creating ilert uptime monitor error: empty response ")
		return diag.Errorf("alert source response is empty")
	}

	d.SetId(strconv.FormatInt(result.UptimeMonitor.ID, 10))

	return resourceUptimeMonitorRead(ctx, d, m)
}

func resourceUptimeMonitorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	uptimeMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse uptime monitor id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading uptime monitor: %s", d.Id())

	result := &ilert.GetUptimeMonitorOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetUptimeMonitor(&ilert.GetUptimeMonitorInput{UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing uptime monitor %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for uptime monitor with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an uptime monitor with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.UptimeMonitor == nil {
		log.Printf("[ERROR] Reading ilert uptime monitor error: empty response ")
		return diag.Errorf("uptime monitor response is empty")
	}

	d.Set("name", result.UptimeMonitor.Name)
	d.Set("region", result.UptimeMonitor.Region)
	d.Set("check_type", result.UptimeMonitor.CheckType)

	checkParams := map[string]interface{}{}
	if result.UptimeMonitor.CheckParams.URL != "" {
		checkParams["url"] = result.UptimeMonitor.CheckParams.URL
	} else if result.UptimeMonitor.CheckParams.Host != "" {
		checkParams["host"] = result.UptimeMonitor.CheckParams.Host
		if result.UptimeMonitor.CheckParams.Port > 0 {
			checkParams["port"] = result.UptimeMonitor.CheckParams.Port
		}
	}
	if result.UptimeMonitor.CheckParams.ResponseKeywords != nil && len(result.UptimeMonitor.CheckParams.ResponseKeywords) > 0 {
		checkParams["response_keywords"] = result.UptimeMonitor.CheckParams.ResponseKeywords
	}
	if result.UptimeMonitor.CheckParams.AlertBeforeSec > 0 {
		checkParams["alert_before_sec"] = result.UptimeMonitor.CheckParams.AlertBeforeSec
	}
	if result.UptimeMonitor.CheckParams.AlertOnFingerprintChange {
		checkParams["alert_on_fingerprint_change"] = result.UptimeMonitor.CheckParams.AlertOnFingerprintChange
	}
	d.Set("check_params", []interface{}{checkParams})

	d.Set("interval_sec", result.UptimeMonitor.IntervalSec)
	d.Set("timeout_ms", result.UptimeMonitor.TimeoutMs)

	if d.Get("create_incident_after_failed_checks") != nil {
		d.Set("create_incident_after_failed_checks", result.UptimeMonitor.CreateIncidentAfterFailedChecks)
	}

	if d.Get("create_alert_after_failed_checks") != nil {
		d.Set("create_alert_after_failed_checks", result.UptimeMonitor.CreateAlertAfterFailedChecks)
	}
	d.Set("escalation_policy", strconv.FormatInt(result.UptimeMonitor.EscalationPolicy.ID, 10))
	d.Set("paused", result.UptimeMonitor.Paused)
	d.Set("status", result.UptimeMonitor.Status)
	d.Set("embed_url", result.UptimeMonitor.EmbedURL)
	d.Set("share_url", result.UptimeMonitor.ShareURL)

	return nil
}

func resourceUptimeMonitorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	uptimeMonitor, err := buildUptimeMonitor(d)
	if err != nil {
		log.Printf("[ERROR] Building uptime monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	uptimeMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse uptime monitor id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating uptime monitor: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateUptimeMonitor(&ilert.UpdateUptimeMonitorInput{UptimeMonitor: uptimeMonitor, UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for uptime monitor with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an uptime monitor with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert uptime monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceUptimeMonitorRead(ctx, d, m)
}

func resourceUptimeMonitorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	uptimeMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse uptime monitor id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting uptime monitor: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteUptimeMonitor(&ilert.DeleteUptimeMonitorInput{UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for uptime monitor with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an uptime monitor with ID %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert uptime monitor error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceUptimeMonitorExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	uptimeMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse uptime monitor id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading uptime monitor: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetUptimeMonitor(&ilert.GetUptimeMonitorInput{UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert uptime monitor error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for uptime monitor to be read, error: %s", err.Error()))
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
