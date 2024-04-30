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

// Legacy API - please use alert-actions - for more information see https://docs.ilert.com/rest-api/api-version-history#renaming-connections-to-alert-actions
func resourceConnection() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "The resource connection is deprecated! Please use alert action instead.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"alert_source": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"connector": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ilert.ConnectorTypesAll, false),
						},
					},
				},
			},
			"trigger_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  ilert.ConnectionTriggerModes.Automatic,
				ValidateFunc: validation.StringInSlice([]string{
					ilert.ConnectionTriggerModes.Automatic,
					ilert.ConnectionTriggerModes.Manual,
				}, false),
			},
			"trigger_types": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ilert.ConnectionTriggerTypesAll, false),
				},
			},
			"datadog": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"site": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "EU",
							ValidateFunc: validation.StringInSlice([]string{
								"EU",
								"US",
							}, false),
						},
						"tags": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"jira": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:     schema.TypeString,
							Required: true,
						},
						"issue_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "Task",
							ValidateFunc: validation.StringInSlice([]string{
								"Bug",
								"Epic",
								"Subtask",
								"Story",
								"Task",
							}, false),
						},
						"body_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"servicenow": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"caller_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"impact": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"urgency": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"slack": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"channel_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"team_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"team_domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"webhook": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"body_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"zendesk": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"urgent",
								"high",
								"normal",
								"low",
							}, false),
						},
					},
				},
			},
			"github": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"owner": {
							Type:     schema.TypeString,
							Required: true,
						},
						"repository": {
							Type:     schema.TypeString,
							Required: true,
						},
						"labels": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"topdesk": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "firstLine",
							ValidateFunc: validation.StringInSlice([]string{
								"firstLine",
								"secondLine",
								"partial",
							}, false),
						},
					},
				},
			},
			"aws_lambda": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"body_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"azure_faas": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"body_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"google_faas": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"body_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"email": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recipients": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subject": {
							Type:     schema.TypeString,
							Required: true,
						},
						"body_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"sysdig": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"zapier",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tags": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"event_filter": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"zapier": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"autotask",
					"zammad",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"autotask": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"company_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"issue_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ticket_category": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ticket_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"zammad": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"status_page_io",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"status_page_io": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"datadog",
					"jira",
					"servicenow",
					"slack",
					"webhook",
					"zendesk",
					"github",
					"topdesk",
					"aws_lambda",
					"azure_faas",
					"google_faas",
					"email",
					"sysdig",
					"zapier",
					"autotask",
					"zammad",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"page_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
		CreateContext: resourceConnectionCreate,
		ReadContext:   resourceConnectionRead,
		UpdateContext: resourceConnectionUpdate,
		DeleteContext: resourceConnectionDelete,
		Exists:        resourceConnectionExists,
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

func buildConnection(d *schema.ResourceData) (*ilert.Connection, error) {
	name := d.Get("name").(string)

	connection := &ilert.Connection{
		Name: name,
	}

	if val, ok := d.GetOk("alert_source"); ok {
		vL := val.([]interface{})
		aids := make([]int64, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			aid, err := strconv.ParseInt(v["id"].(string), 10, 64)
			if err != nil {
				return nil, unconvertibleIDErr(v["id"].(string), err)
			}
			aids = append(aids, aid)
		}
		connection.AlertSourceIDs = aids
	}

	if val, ok := d.GetOk("connector"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.ConnectorID = v["id"].(string)
			connection.ConnectorType = v["type"].(string)
		}
	}

	if val, ok := d.GetOk("trigger_mode"); ok {
		triggerMode := val.(string)
		connection.TriggerMode = triggerMode
	}

	if val, ok := d.GetOk("trigger_types"); ok {
		vL := val.([]interface{})
		sL := make([]string, 0)
		for _, m := range vL {
			v := m.(string)
			sL = append(sL, v)
		}
		connection.TriggerTypes = sL
	}

	if val, ok := d.GetOk("datadog"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.ConnectionParamsDatadog{
				Site:     v["site"].(string),
				Priority: v["priority"].(string),
			}
			vL := v["tags"].([]interface{})
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			params.Tags = sL
			connection.Params = params
		}
	}

	if val, ok := d.GetOk("jira"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsJira{
				Project:      v["project"].(string),
				IssueType:    v["issue_type"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("servicenow"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsServiceNow{
				CallerID: v["caller_id"].(string),
				Impact:   v["impact"].(string),
				Urgency:  v["urgency"].(string),
			}
		}
	}

	if val, ok := d.GetOk("slack"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsSlack{
				ChannelID:   v["channel_id"].(string),
				ChannelName: v["channel_name"].(string),
				TeamID:      v["team_id"].(string),
				TeamDomain:  v["team_domain"].(string),
			}
		}
	}

	if val, ok := d.GetOk("webhook"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsWebhook{
				WebhookURL:   v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("zendesk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsZendesk{
				Priority: v["priority"].(string),
			}
		}
	}

	if val, ok := d.GetOk("github"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.ConnectionParamsGithub{
				Owner:      v["owner"].(string),
				Repository: v["repository"].(string),
			}
			vL := v["labels"].([]interface{})
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			params.Labels = sL
			connection.Params = params
		}
	}

	if val, ok := d.GetOk("topdesk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsTopdesk{
				Status: v["status"].(string),
			}
		}
	}

	if val, ok := d.GetOk("aws_lambda"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsAWSLambda{
				WebhookURL:   v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("azure_faas"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsAzureFunction{
				WebhookURL:   v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("google_faas"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsGoogleFunction{
				WebhookURL:   v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("email"); ok {
		if vL, ok := val.([]interface{}); ok && len(vL) > 0 {
			if v, ok := vL[0].(map[string]interface{}); ok && len(v) > 0 {
				params := &ilert.ConnectionParamsEmail{}
				if p, ok := v["subject"].(string); ok && p != "" {
					params.Subject = p
				}
				if p, ok := v["body_template"].(string); ok && p != "" {
					params.BodyTemplate = p
				}
				if vL, ok := v["recipients"].([]interface{}); ok && len(vL) > 0 {
					sL := make([]string, 0)
					for _, m := range vL {
						if v, ok := m.(string); ok && v != "" {
							sL = append(sL, v)
						}
					}
					params.Recipients = sL
				}
				connection.Params = params
			}
		}
	}

	if val, ok := d.GetOk("sysdig"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.ConnectionParamsSysdig{
				EventFilter: v["event_filter"].(string),
			}
			vL := v["tags"].([]interface{})
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			params.Tags = sL
			connection.Params = params
		}
	}

	if val, ok := d.GetOk("zapier"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsZapier{
				WebhookURL: v["url"].(string),
			}
		}
	}

	if val, ok := d.GetOk("autotask"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsAutotask{
				CompanyID:      v["company_id"].(string),
				IssueType:      v["issue_type"].(string),
				QueueID:        int64(v["queue_id"].(int)),
				TicketCategory: v["ticket_category"].(string),
				TicketType:     v["ticket_type"].(string),
			}
		}
	}

	if val, ok := d.GetOk("zammad"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsZammad{
				Email: v["email"].(string),
			}
		}
	}

	if val, ok := d.GetOk("status_page_io"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			connection.Params = &ilert.ConnectionParamsStatusPageIO{
				PageID: v["page_id"].(string),
			}
		}
	}

	return connection, nil
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connection, err := buildConnection(d)
	if err != nil {
		log.Printf("[ERROR] Building connection error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating connection %s", connection.Name)

	result := &ilert.CreateConnectionOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.CreateConnection(&ilert.CreateConnectionInput{Connection: connection})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing connection %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert connection error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connection to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an connection with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.Connection == nil {
		log.Printf("[ERROR] Creating ilert connection error: empty response ")
		return diag.FromErr(fmt.Errorf("connection response is empty"))
	}

	d.SetId(result.Connection.ID)

	return resourceConnectionRead(ctx, d, m)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connectionID := d.Id()
	log.Printf("[DEBUG] Reading connection: %s", d.Id())

	result := &ilert.GetConnectionOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetConnection(&ilert.GetConnectionInput{ConnectionID: ilert.String(connectionID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing connection %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connection with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an connection with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.Connection == nil {
		log.Printf("[ERROR] Reading ilert connection error: empty response ")
		return diag.Errorf("connection response is empty")
	}

	d.Set("name", result.Connection.Name)

	alertSources, err := flattenConnectionAlertSourceIDList(result.Connection.AlertSourceIDs)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("alert_source", alertSources); err != nil {
		return diag.Errorf("error setting alert sources: %s", err)
	}

	connector := map[string]interface{}{}
	log.Printf("[DEBUG] Reading ilert connection: %s , connector id: %s", d.Id(), result.Connection.ConnectorID)
	if result.Connection.ConnectorID != "" {
		connector["id"] = result.Connection.ConnectorID
		connector["type"] = result.Connection.ConnectorType
	}
	d.Set("connector", []interface{}{connector})
	d.Set("trigger_mode", result.Connection.TriggerMode)
	d.Set("trigger_types", result.Connection.TriggerTypes)
	d.Set("created_at", result.Connection.CreatedAt)
	d.Set("updated_at", result.Connection.UpdatedAt)

	switch result.Connection.ConnectorType {
	case ilert.ConnectorTypes.Jira:
		d.Set("jira", []interface{}{
			map[string]interface{}{
				"project":       result.Connection.Params.Project,
				"issue_type":    result.Connection.Params.IssueType,
				"body_template": result.Connection.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.ServiceNow:
		d.Set("servicenow", []interface{}{
			map[string]interface{}{
				"caller_id": result.Connection.Params.CallerID,
				"impact":    result.Connection.Params.Impact,
				"urgency":   result.Connection.Params.Urgency,
			},
		})
	case ilert.ConnectorTypes.Slack:
		d.Set("slack", []interface{}{
			map[string]interface{}{
				"channel_id":   result.Connection.Params.ChannelID,
				"channel_name": result.Connection.Params.ChannelName,
				"team_id":      result.Connection.Params.TeamID,
				"team_domain":  result.Connection.Params.TeamDomain,
			},
		})
	case ilert.ConnectorTypes.Webhook:
		d.Set("webhook", []interface{}{
			map[string]interface{}{
				"url":           result.Connection.Params.WebhookURL,
				"body_template": result.Connection.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Zendesk:
		d.Set("zendesk", []interface{}{
			map[string]interface{}{
				"priority": result.Connection.Params.Priority,
			},
		})
	case ilert.ConnectorTypes.Github:
		d.Set("github", []interface{}{
			map[string]interface{}{
				"owner":      result.Connection.Params.Owner,
				"repository": result.Connection.Params.Repository,
				"labels":     result.Connection.Params.Labels,
			},
		})
	case ilert.ConnectorTypes.Topdesk:
		d.Set("topdesk", []interface{}{
			map[string]interface{}{
				"status": result.Connection.Params.Status,
			},
		})
	case ilert.ConnectorTypes.Email:
		d.Set("email", []interface{}{
			map[string]interface{}{
				"recipients":    result.Connection.Params.Recipients,
				"subject":       result.Connection.Params.Subject,
				"body_template": result.Connection.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Autotask:
		d.Set("autotask", []interface{}{
			map[string]interface{}{
				"company_id":      result.Connection.Params.CompanyID,
				"issue_type":      result.Connection.Params.IssueType,
				"queue_id":        result.Connection.Params.QueueID,
				"ticket_category": result.Connection.Params.TicketCategory,
				"ticket_type":     result.Connection.Params.TicketType,
			},
		})
	case ilert.ConnectorTypes.Zammad:
		d.Set("zammad", []interface{}{
			map[string]interface{}{
				"email": result.Connection.Params.Email,
			},
		})
	}

	return nil
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connection, err := buildConnection(d)
	if err != nil {
		log.Printf("[ERROR] Building connection error %s", err.Error())
		return diag.FromErr(err)
	}

	connectionID := d.Id()
	log.Printf("[DEBUG] Updating connection: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateConnection(&ilert.UpdateConnectionInput{Connection: connection, ConnectionID: ilert.String(connectionID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connection with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an connection with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert connection error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceConnectionRead(ctx, d, m)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	connectionID := d.Id()
	log.Printf("[DEBUG] Deleting connection: %s", d.Id())

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := client.DeleteConnection(&ilert.DeleteConnectionInput{ConnectionID: ilert.String(connectionID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connection with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an connection with ID %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert connection error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceConnectionExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	connectionID := d.Id()
	log.Printf("[DEBUG] Reading connection: %s", d.Id())
	ctx := context.Background()
	result := false
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetConnection(&ilert.GetConnectionInput{ConnectionID: ilert.String(connectionID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert connection error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for connection to be read, error: %s", err.Error()))
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

func flattenConnectionAlertSourceIDList(list []int64) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, item := range list {
		result := make(map[string]interface{})
		result["id"] = strconv.FormatInt(item, 10)
		results = append(results, result)
	}

	return results, nil
}
