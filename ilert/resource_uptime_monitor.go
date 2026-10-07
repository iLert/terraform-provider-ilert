package ilert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// uptimeMonitorDiscontinued is what the uptime monitor resource and data source answer with
const uptimeMonitorDiscontinued = "ilert discontinued uptime monitoring after 30.06.2024 and no longer serves uptime monitors, remove the %s from your configuration. It will be removed in the next major version of the provider"

// ilert discontinued uptime monitoring after 30.06.2024 and the API answers every uptime monitor request
// with a 404. The resource keeps its schema so existing configurations still parse and plan as before,
// but it no longer calls the API: creating one fails with an error that says so, and one found in the
// state is dropped from it, which is what the 404 caused until now.
func resourceUptimeMonitor() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: fmt.Sprintf(uptimeMonitorDiscontinued, "ilert_uptime_monitor resource"),
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

func resourceUptimeMonitorCreate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Errorf(uptimeMonitorDiscontinued, "ilert_uptime_monitor resource")
}

func resourceUptimeMonitorRead(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	log.Printf("[WARN] Removing uptime monitor %s from state because ilert discontinued uptime monitoring", d.Id())
	d.SetId("")
	return nil
}

func resourceUptimeMonitorUpdate(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Errorf(uptimeMonitorDiscontinued, "ilert_uptime_monitor resource")
}

func resourceUptimeMonitorDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")
	return nil
}
