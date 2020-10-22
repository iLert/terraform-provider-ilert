package ilert

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go"
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
				ValidateFunc: validateValueFunc([]string{
					"EU",
					"US",
				}),
			},
			"check_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validateValueFunc([]string{
					"http",
					"ping",
					"tcp",
					"udp",
				}),
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
					},
				},
			},
			"interval_sec": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
				ValidateFunc: validateIntValueFunc([]int{
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
			"create_incident_after_failed_checks": {
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
		Create: resourceUptimeMonitorCreate,
		Read:   resourceUptimeMonitorRead,
		Update: resourceUptimeMonitorUpdate,
		Delete: resourceUptimeMonitorDelete,
		Exists: resourceUptimeMonitorExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

	if val, ok := d.GetOk("paused"); ok {
		paused := val.(bool)
		uptimeMonitor.Paused = paused
	}

	return uptimeMonitor, nil
}

func resourceUptimeMonitorCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	uptimeMonitor, err := buildUptimeMonitor(d)
	if err != nil {
		log.Printf("[ERROR] Building uptime monitor error %s", err.Error())
		return err
	}

	log.Printf("[INFO] Creating uptime monitor %s", uptimeMonitor.Name)

	result, err := client.CreateUptimeMonitor(&ilert.CreateUptimeMonitorInput{UptimeMonitor: uptimeMonitor})
	if err != nil {
		log.Printf("[ERROR] Creating iLert uptime monitor error %s", err.Error())
		return err
	}

	d.SetId(strconv.FormatInt(result.UptimeMonitor.ID, 10))

	return resourceUptimeMonitorRead(d, m)
}

func resourceUptimeMonitorRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	uptimeMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse uptime monitor id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading uptime monitor: %s", d.Id())
	result, err := client.GetUptimeMonitor(&ilert.GetUptimeMonitorInput{UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			log.Printf("[WARN] Removing uptime monitor %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Could not read an uptime monitor with ID %s", d.Id())
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
	d.Set("check_params", []interface{}{checkParams})

	d.Set("interval_sec", result.UptimeMonitor.IntervalSec)
	d.Set("timeout_ms", result.UptimeMonitor.TimeoutMs)
	d.Set("create_incident_after_failed_checks", result.UptimeMonitor.CreateIncidentAfterFailedChecks)
	d.Set("escalation_policy", strconv.FormatInt(result.UptimeMonitor.EscalationPolicy.ID, 10))
	d.Set("paused", result.UptimeMonitor.Paused)
	d.Set("status", result.UptimeMonitor.Status)
	d.Set("embed_url", result.UptimeMonitor.EmbedURL)
	d.Set("share_url", result.UptimeMonitor.ShareURL)

	return nil
}

func resourceUptimeMonitorUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	uptimeMonitor, err := buildUptimeMonitor(d)
	if err != nil {
		log.Printf("[ERROR] Building uptime monitor error %s", err.Error())
		return err
	}

	uptimeMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse uptime monitor id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Updating uptime monitor: %s", d.Id())
	_, err = client.UpdateUptimeMonitor(&ilert.UpdateUptimeMonitorInput{UptimeMonitor: uptimeMonitor, UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
	if err != nil {
		log.Printf("[ERROR] Updating iLert uptime monitor error %s", err.Error())
		return err
	}
	return resourceUptimeMonitorRead(d, m)
}

func resourceUptimeMonitorDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	uptimeMonitorID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse uptime monitor id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Deleting uptime monitor: %s", d.Id())
	_, err = client.DeleteUptimeMonitor(&ilert.DeleteUptimeMonitorInput{UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
	if err != nil {
		return err
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
	_, err = client.GetUptimeMonitor(&ilert.GetUptimeMonitorInput{UptimeMonitorID: ilert.Int64(uptimeMonitorID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
