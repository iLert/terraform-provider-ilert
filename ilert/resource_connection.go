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

func resourceConnection() *schema.Resource {
	return &schema.Resource{
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
							ValidateFunc: validateStringValueFunc(ilert.ConnectorTypesAll),
						},
					},
				},
			},
			"trigger_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  ilert.ConnectionTriggerModes.Automatic,
				ValidateFunc: validateStringValueFunc([]string{
					ilert.ConnectionTriggerModes.Automatic,
					ilert.ConnectionTriggerModes.Manual,
				}),
			},
			"trigger_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateStringValueFunc(ilert.ConnectionTriggerTypesAll),
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
							ValidateFunc: validateStringValueFunc([]string{
								"EU",
								"US",
							}),
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
							ValidateFunc: validateStringValueFunc([]string{
								"Bug",
								"Epic",
								"Subtask",
								"Story",
								"Task",
							}),
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
							ValidateFunc: validateStringValueFunc([]string{
								"urgent",
								"high",
								"normal",
								"low",
							}),
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
							ValidateFunc: validateStringValueFunc([]string{
								"firstLine",
								"secondLine",
								"partial",
							}),
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
		Create: resourceConnectionCreate,
		Read:   resourceConnectionRead,
		Update: resourceConnectionUpdate,
		Delete: resourceConnectionDelete,
		Exists: resourceConnectionExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.ConnectionParamsEmail{
				Subject:      v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
			vL := v["recipients"].([]interface{})
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			params.Recipients = sL
			connection.Params = params
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

func resourceConnectionCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connection, err := buildConnection(d)
	if err != nil {
		log.Printf("[ERROR] Building connection error %s", err.Error())
		return err
	}

	log.Printf("[INFO] Creating connection %s", connection.Name)

	result, err := client.CreateConnection(&ilert.CreateConnectionInput{Connection: connection})
	if err != nil {
		log.Printf("[ERROR] Creating iLert connection error %s", err.Error())
		return err
	}

	d.SetId(result.Connection.ID)

	return resourceConnectionRead(d, m)
}

func resourceConnectionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connectionID := d.Id()
	log.Printf("[DEBUG] Reading connection: %s", d.Id())
	result, err := client.GetConnection(&ilert.GetConnectionInput{ConnectionID: ilert.String(connectionID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			log.Printf("[WARN] Removing connection %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Could not read an connection with ID %s", d.Id())
	}

	d.Set("name", result.Connection.Name)

	alertSources, err := flattenConnectionAlertSourceIDList(result.Connection.AlertSourceIDs)
	if err != nil {
		return err
	}
	if err := d.Set("alert_source", alertSources); err != nil {
		return fmt.Errorf("error setting alert sources: %s", err)
	}

	connector := map[string]interface{}{}
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
	case ilert.ConnectorTypes.Datadog:
		d.Set("datadog", []interface{}{
			map[string]interface{}{
				"priority": result.Connection.Params.Priority,
				"site":     result.Connection.Params.Site,
				"tags":     result.Connection.Params.Tags,
			},
		})
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
	case ilert.ConnectorTypes.AWSLambda,
		ilert.ConnectorTypes.AzureFAAS,
		ilert.ConnectorTypes.GoogleFAAS:
		d.Set("aws_lambda", []interface{}{
			map[string]interface{}{
				"url":           result.Connection.Params.WebhookURL,
				"body_template": result.Connection.Params.BodyTemplate,
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
	case ilert.ConnectorTypes.Sysdig:
		d.Set("sysdig", []interface{}{
			map[string]interface{}{
				"tags":         result.Connection.Params.Tags,
				"event_filter": result.Connection.Params.EventFilter,
			},
		})
	case ilert.ConnectorTypes.Zapier:
		d.Set("zapier", []interface{}{
			map[string]interface{}{
				"url": result.Connection.Params.WebhookURL,
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
	case ilert.ConnectorTypes.StatusPageIO:
		d.Set("zammad", []interface{}{
			map[string]interface{}{
				"page_id": result.Connection.Params.PageID,
			},
		})
	}

	return nil
}

func resourceConnectionUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connection, err := buildConnection(d)
	if err != nil {
		log.Printf("[ERROR] Building connection error %s", err.Error())
		return err
	}

	connectionID := d.Id()
	log.Printf("[DEBUG] Updating connection: %s", d.Id())
	_, err = client.UpdateConnection(&ilert.UpdateConnectionInput{Connection: connection, ConnectionID: ilert.String(connectionID)})
	if err != nil {
		log.Printf("[ERROR] Updating iLert connection error %s", err.Error())
		return err
	}
	return resourceConnectionRead(d, m)
}

func resourceConnectionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	connectionID := d.Id()
	log.Printf("[DEBUG] Deleting connection: %s", d.Id())
	_, err := client.DeleteConnection(&ilert.DeleteConnectionInput{ConnectionID: ilert.String(connectionID)})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceConnectionExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	connectionID := d.Id()
	log.Printf("[DEBUG] Reading connection: %s", d.Id())
	_, err := client.GetConnection(&ilert.GetConnectionInput{ConnectionID: ilert.String(connectionID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			return false, nil
		}
		return false, err
	}
	return true, nil
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
