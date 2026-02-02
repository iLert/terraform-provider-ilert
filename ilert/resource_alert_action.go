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

func resourceAlertAction() *schema.Resource {
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "View available connector types at https://docs.ilert.com/developer-docs/rest-api/api-reference/alert-actions#post-alert-actions",
						},
					},
				},
			},
			"trigger_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				Default:      ilert.AlertActionTriggerModes.Automatic,
				ValidateFunc: validation.StringInSlice(ilert.AlertActionTriggerModesAll, false),
			},
			"trigger_types": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ilert.AlertActionTriggerTypesAll, false),
				},
			},
			"jira": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
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
						"body_template": {
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
			"email": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
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
			"autotask": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"dingtalk": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_at_all": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"at_mobiles": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"dingtalk_action": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secret": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"is_at_all": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"at_mobiles": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"automation_rule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alert_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ilert.AlertTypeAll, false),
						},
						"resolve_incident": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"service_status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ilert.ServiceStatusAll, false),
						},
						"template_id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"send_notification": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"service_ids": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
					},
				},
			},
			"telegram": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"microsoft_teams_bot": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
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
						"team_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "chat | meeting",
						},
					},
				},
			},
			"microsoft_teams_webhook": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
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
			"slack_webhook": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
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
			"alert_filter": {
				Type:       schema.TypeList,
				Optional:   true,
				MinItems:   1,
				MaxItems:   1,
				Deprecated: "This field is deprecated, use 'conditions' instead. If both are used this field is ignored.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operator": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ilert.AlertFilterOperatorAll, false),
						},
						"predicate": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(ilert.AlertFilterPredicateFieldsAll, false),
									},
									"criteria": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(ilert.AlertFilterPredicateCriteriaAll, false),
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
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
			"delay_sec": {
				Type:       schema.TypeInt,
				Optional:   true,
				Deprecated: "The field delay_sec is deprecated! Please use escalation_ended_delay_sec instead for trigger_type 'alert_escalation_ended' or not_resolved_delay_sec for trigger_type 'alert_not_resolved'.",
			},
			"escalation_ended_delay_sec": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"not_resolved_delay_sec": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"conditions": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		CreateContext: resourceAlertActionCreate,
		ReadContext:   resourceAlertActionRead,
		UpdateContext: resourceAlertActionUpdate,
		DeleteContext: resourceAlertActionDelete,
		Exists:        resourceAlertActionExists,
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

func buildAlertAction(d *schema.ResourceData) (*ilert.AlertAction, error) {
	name := d.Get("name").(string)

	alertAction := &ilert.AlertAction{
		Name: name,
	}

	if val, ok := d.GetOk("alert_source"); ok {
		vL := val.([]interface{})
		als := make([]ilert.AlertSource, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			asid, err := strconv.ParseInt(v["id"].(string), 10, 64)
			if err != nil {
				return nil, unconvertibleIDErr(v["id"].(string), err)
			}
			as := ilert.AlertSource{
				ID: asid,
			}
			als = append(als, as)
		}
		alertAction.AlertSources = &als
	}

	if val, ok := d.GetOk("connector"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.ConnectorID = v["id"].(string)
			alertAction.ConnectorType = v["type"].(string)
		}
	}

	if val, ok := d.GetOk("trigger_mode"); ok {
		triggerMode := val.(string)
		alertAction.TriggerMode = triggerMode
	}

	if val, ok := d.GetOk("trigger_types"); ok {
		vL := val.([]interface{})
		sL := make([]string, 0)
		for _, m := range vL {
			v := m.(string)
			sL = append(sL, v)
		}
		alertAction.TriggerTypes = sL
	}

	if val, ok := d.GetOk("jira"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsJira{
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
			alertAction.Params = &ilert.AlertActionParamsServiceNow{
				CallerID:     v["caller_id"].(string),
				Impact:       v["impact"].(string),
				Urgency:      v["urgency"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("slack"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsSlack{
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
			alertAction.Params = &ilert.AlertActionParamsWebhook{
				WebhookURL:   v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("zendesk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsZendesk{
				Priority: v["priority"].(string),
			}
		}
	}

	if val, ok := d.GetOk("github"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.AlertActionParamsGithub{
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
			alertAction.Params = params
		}
	}

	if val, ok := d.GetOk("topdesk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsTopdesk{
				Status: v["status"].(string),
			}
		}
	}

	if val, ok := d.GetOk("email"); ok {
		if vL, ok := val.([]interface{}); ok && len(vL) > 0 {
			if v, ok := vL[0].(map[string]interface{}); ok && len(v) > 0 {
				params := &ilert.AlertActionParamsEmail{}
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
				alertAction.Params = params
			}
		}
	}

	if val, ok := d.GetOk("autotask"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsAutotask{
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
			alertAction.Params = &ilert.AlertActionParamsZammad{
				Email: v["email"].(string),
			}
		}
	}

	if val, ok := d.GetOk("dingtalk"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.AlertActionParamsDingTalk{
				IsAtAll: v["is_at_all"].(bool),
			}
			vL := v["at_mobiles"].([]interface{})
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			params.AtMobiles = sL
			alertAction.Params = params
		}
	}

	if val, ok := d.GetOk("dingtalk_action"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.AlertActionParamsDingTalkAction{
				URL:     v["url"].(string),
				Secret:  v["secret"].(string),
				IsAtAll: v["is_at_all"].(bool),
			}
			vL := v["at_mobiles"].([]interface{})
			sL := make([]string, 0)
			for _, m := range vL {
				v := m.(string)
				sL = append(sL, v)
			}
			params.AtMobiles = sL
			alertAction.Params = params
		}
	}

	if val, ok := d.GetOk("automation_rule"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			params := &ilert.AlertActionParamsAutomationRule{
				AlertType:        v["alert_type"].(string),
				ResolveIncident:  v["resolve_incident"].(bool),
				ServiceStatus:    v["service_status"].(string),
				TemplateId:       int64(v["template_id"].(int)),
				SendNotification: v["send_notification"].(bool),
			}
			vL := v["service_ids"].([]interface{})
			sL := make([]int64, 0)
			for _, m := range vL {
				v := int64(m.(int))
				sL = append(sL, v)
			}
			params.ServiceIds = sL
			alertAction.Params = params
		}
	}

	if val, ok := d.GetOk("telegram"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsTelegram{
				ChannelID: v["channel_id"].(string),
			}
		}
	}

	if val, ok := d.GetOk("microsoft_teams_bot"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsMicrosoftTeamsBot{
				ChannelID:   v["channel_id"].(string),
				ChannelName: v["channel_name"].(string),
				TeamID:      v["team_id"].(string),
				TeamName:    v["team_name"].(string),
				Type:        v["type"].(string),
			}
		}
	}

	if val, ok := d.GetOk("microsoft_teams_webhook"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsMicrosoftTeamsWebhook{
				URL:          v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("slack_webhook"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertAction.Params = &ilert.AlertActionParamsSlackWebhook{
				URL: v["url"].(string),
			}
		}
	}

	if val, ok := d.GetOk("alert_filter"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			filter := &ilert.AlertFilter{
				Operator: v["operator"].(string),
			}
			vL := v["predicate"].([]interface{})
			pL := make([]ilert.AlertFilterPredicate, 0)
			for _, m := range vL {
				v := m.(map[string]interface{})
				p := ilert.AlertFilterPredicate{
					Field:    v["field"].(string),
					Criteria: v["criteria"].(string),
					Value:    v["value"].(string),
				}
				pL = append(pL, p)
			}
			filter.Predicates = pL
			alertAction.AlertFilter = filter
		}
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
		alertAction.Teams = &tms
	}

	if val, ok := d.GetOk("delay_sec"); ok {
		delaySec := val.(int)
		if delaySec != 0 && (delaySec < 30 || delaySec > 7200) {
			return nil, fmt.Errorf("[ERROR] Can't set 'delay_sec', value must be either 0 or between 30 and 7200")
		}
		alertAction.DelaySec = val.(int)
	}

	if val, ok := d.GetOk("escalation_ended_delay_sec"); ok {
		escalationEndedDelaySec := val.(int)
		if escalationEndedDelaySec != 0 && (escalationEndedDelaySec < 30 || escalationEndedDelaySec > 7200) {
			return nil, fmt.Errorf("[ERROR] Can't set 'escalation_ended_delay_sec', value must be either 0 or between 30 and 7200")
		}
		alertAction.EscalationEndedDelaySec = val.(int)
	}

	if val, ok := d.GetOk("not_resolved_delay_sec"); ok {
		notResolvedDelaySec := val.(int)
		if notResolvedDelaySec != 0 && (notResolvedDelaySec < 60 || notResolvedDelaySec > 7200) {
			return nil, fmt.Errorf("[ERROR] Can't set 'not_resolved_delay_sec', value must be either 0 or between 60 and 7200")
		}
		alertAction.NotResolvedDelaySec = val.(int)
	}

	if val, ok := d.GetOk("conditions"); ok {
		alertAction.Conditions = val.(string)
	}

	return alertAction, nil
}

func resourceAlertActionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertAction, err := buildAlertAction(d)
	if err != nil {
		log.Printf("[ERROR] Building alert action error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating alert action %s", alertAction.Name)

	result := &ilert.CreateAlertActionOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.CreateAlertAction(&ilert.CreateAlertActionInput{AlertAction: alertAction})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing alert action %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert alert action rule error %s, so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create an alert action with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Creating ilert alert action error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.AlertAction == nil {
		log.Printf("[ERROR] Creating ilert alert action error: empty response")
		return diag.FromErr(fmt.Errorf("alert action response is empty"))
	}

	d.SetId(result.AlertAction.ID)

	return resourceAlertActionRead(ctx, d, m)
}

func resourceAlertActionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertActionID := d.Id()
	log.Printf("[DEBUG] Reading alert action: %s", d.Id())

	result := &ilert.GetAlertActionOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		version := 2
		if val, ok := d.GetOk("alert_source"); ok && len(val.([]interface{})) == 1 {
			if val, ok := d.GetOk("team"); !ok || len(val.([]interface{})) == 0 {
				if val, ok := d.GetOk("conditions"); !ok || len(val.(string)) == 0 {
					version = 1
				}
			}
		}
		r, err := client.GetAlertAction(&ilert.GetAlertActionInput{AlertActionID: ilert.String(alertActionID), Version: ilert.Int(version)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing alert action %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an alert action with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert alert action error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.AlertAction == nil {
		log.Printf("[ERROR] Reading ilert alert action error: empty response")
		return diag.Errorf("alert action response is empty")
	}

	d.Set("name", result.AlertAction.Name)

	if _, ok := d.GetOk("alert_source"); ok && result.AlertAction.AlertSources != nil {
		alertSources, err := flattenAlertActionAlertSourcesList(*result.AlertAction.AlertSources)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("alert_source", alertSources); err != nil {
			return diag.Errorf("error setting alert sources: %s", err)
		}
	}

	connector := map[string]interface{}{}
	log.Printf("[DEBUG] Reading ilert alert action: %s , connector id: %s", d.Id(), result.AlertAction.ConnectorID)
	if result.AlertAction.ConnectorID != "" {
		connector["id"] = result.AlertAction.ConnectorID
	}
	connector["type"] = result.AlertAction.ConnectorType
	d.Set("connector", []interface{}{connector})
	d.Set("trigger_mode", result.AlertAction.TriggerMode)
	d.Set("trigger_types", result.AlertAction.TriggerTypes)
	d.Set("created_at", result.AlertAction.CreatedAt)
	d.Set("updated_at", result.AlertAction.UpdatedAt)

	switch result.AlertAction.ConnectorType {
	case ilert.ConnectorTypes.Jira:
		d.Set("jira", []interface{}{
			map[string]interface{}{
				"project":       result.AlertAction.Params.Project,
				"issue_type":    result.AlertAction.Params.IssueType,
				"body_template": result.AlertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.ServiceNow:
		d.Set("servicenow", []interface{}{
			map[string]interface{}{
				"caller_id":     result.AlertAction.Params.CallerID,
				"impact":        result.AlertAction.Params.Impact,
				"urgency":       result.AlertAction.Params.Urgency,
				"body_template": result.AlertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Slack:
		d.Set("slack", []interface{}{
			map[string]interface{}{
				"channel_id":   result.AlertAction.Params.ChannelID,
				"channel_name": result.AlertAction.Params.ChannelName,
				"team_id":      result.AlertAction.Params.TeamID,
				"team_domain":  result.AlertAction.Params.TeamDomain,
			},
		})
	case ilert.ConnectorTypes.Webhook:
		d.Set("webhook", []interface{}{
			map[string]interface{}{
				"url":           result.AlertAction.Params.WebhookURL,
				"body_template": result.AlertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Zendesk:
		d.Set("zendesk", []interface{}{
			map[string]interface{}{
				"priority": result.AlertAction.Params.Priority,
			},
		})
	case ilert.ConnectorTypes.Github:
		d.Set("github", []interface{}{
			map[string]interface{}{
				"owner":      result.AlertAction.Params.Owner,
				"repository": result.AlertAction.Params.Repository,
				"labels":     result.AlertAction.Params.Labels,
			},
		})
	case ilert.ConnectorTypes.Topdesk:
		d.Set("topdesk", []interface{}{
			map[string]interface{}{
				"status": result.AlertAction.Params.Status,
			},
		})
	case ilert.ConnectorTypes.Email:
		d.Set("email", []interface{}{
			map[string]interface{}{
				"recipients":    result.AlertAction.Params.Recipients,
				"subject":       result.AlertAction.Params.Subject,
				"body_template": result.AlertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Autotask:
		d.Set("autotask", []interface{}{
			map[string]interface{}{
				"company_id":      result.AlertAction.Params.CompanyID,
				"issue_type":      result.AlertAction.Params.IssueType,
				"queue_id":        result.AlertAction.Params.QueueID,
				"ticket_category": result.AlertAction.Params.TicketCategory,
				"ticket_type":     result.AlertAction.Params.TicketType,
			},
		})
	case ilert.ConnectorTypes.Zammad:
		d.Set("zammad", []interface{}{
			map[string]interface{}{
				"email": result.AlertAction.Params.Email,
			},
		})
	case ilert.ConnectorTypes.DingTalk:
		d.Set("dingtalk", []interface{}{
			map[string]interface{}{
				"is_at_all":  result.AlertAction.Params.IsAtAll,
				"at_mobiles": result.AlertAction.Params.AtMobiles,
			},
		})
	case ilert.ConnectorTypes.DingTalkAction:
		d.Set("dingtalk_action", []interface{}{
			map[string]interface{}{
				"url":        result.AlertAction.Params.URL,
				"secret":     result.AlertAction.Params.Secret,
				"is_at_all":  result.AlertAction.Params.IsAtAll,
				"at_mobiles": result.AlertAction.Params.AtMobiles,
			},
		})
	case ilert.ConnectorTypes.AutomationRule:
		d.Set("automation_rule", []interface{}{
			map[string]interface{}{
				"alert_type":        result.AlertAction.Params.AlertType,
				"resolve_incident":  result.AlertAction.Params.ResolveIncident,
				"service_status":    result.AlertAction.Params.ServiceStatus,
				"template_id":       result.AlertAction.Params.TemplateId,
				"send_notification": result.AlertAction.Params.SendNotification,
				"service_ids":       result.AlertAction.Params.ServiceIds,
			},
		})
	case ilert.ConnectorTypes.Telegram:
		d.Set("telegram", []interface{}{
			map[string]interface{}{
				"channel_id": result.AlertAction.Params.ChannelID,
			},
		})
	case ilert.ConnectorTypes.MicrosoftTeamsBot:
		d.Set("microsoft_teams_bot", []interface{}{
			map[string]interface{}{
				"channel_id":   result.AlertAction.Params.ChannelID,
				"channel_name": result.AlertAction.Params.ChannelName,
				"team_id":      result.AlertAction.Params.TeamID,
				"team_name":    result.AlertAction.Params.TeamName,
				"type":         result.AlertAction.Params.Type,
			},
		})
	case ilert.ConnectorTypes.MicrosoftTeamsWebhook:
		d.Set("microsoft_teams_webhook", []interface{}{
			map[string]interface{}{
				"url":          result.AlertAction.Params.URL,
				"bodyTemplate": result.AlertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.SlackWebhook:
		d.Set("slack_webhook", []interface{}{
			map[string]interface{}{
				"url": result.AlertAction.Params.URL,
			},
		})
	}

	alertFilter, err := flattenAlertActionAlertFilter(result.AlertAction.AlertFilter)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("alert_filter", alertFilter); err != nil {
		return diag.Errorf("error setting alert filter: %s", err)
	}

	if val, ok := d.GetOk("team"); ok {
		if val != nil && result.AlertAction.Teams != nil {
			vL := val.([]interface{})
			teams := make([]interface{}, 0)
			for i, item := range *result.AlertAction.Teams {
				team := make(map[string]interface{})
				if i >= len(vL) {
					break
				}
				v := vL[i].(map[string]interface{})
				team["id"] = item.ID

				// Means: if server response has a name set, and the user typed in a name too,
				// only then team name is stored in the terraform state
				if item.Name != "" && v["name"] != nil && v["name"].(string) != "" {
					team["name"] = item.Name
				}
				teams = append(teams, team)
			}

			if err := d.Set("team", teams); err != nil {
				return diag.Errorf("error setting teams: %s", err)
			}
		}
	}

	d.Set("delay_sec", result.AlertAction.DelaySec)
	d.Set("escalation_ended_delay_sec", result.AlertAction.EscalationEndedDelaySec)
	d.Set("not_resolved_delay_sec", result.AlertAction.NotResolvedDelaySec)

	if val, ok := d.GetOk("alert_source"); ok && len(val.([]interface{})) == 1 {
		if v, ok := d.GetOk("team"); !ok || len(v.([]interface{})) == 0 {
			sourceId := result.AlertAction.AlertSourceIDs[0]

			sources := make([]interface{}, 0)
			source := make(map[string]interface{})
			source["id"] = strconv.FormatInt(sourceId, 10)
			sources = append(sources, source)

			d.Set("alert_source", sources)
		}
	}

	d.Set("conditions", result.AlertAction.Conditions)

	return nil
}

func resourceAlertActionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertAction, err := buildAlertAction(d)
	if err != nil {
		log.Printf("[ERROR] Building alert action error %s", err.Error())
		return diag.FromErr(err)
	}

	alertActionID := d.Id()
	log.Printf("[DEBUG] Updating alert action: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateAlertAction(&ilert.UpdateAlertActionInput{AlertAction: alertAction, AlertActionID: ilert.String(alertActionID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an alert action with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert alert action error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceAlertActionRead(ctx, d, m)
}

func resourceAlertActionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertActionID := d.Id()
	log.Printf("[DEBUG] Deleting alert action: %s", d.Id())

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := client.DeleteAlertAction(&ilert.DeleteAlertActionInput{AlertActionID: ilert.String(alertActionID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an alert action with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert alert action error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceAlertActionExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	alertActionID := d.Id()
	log.Printf("[DEBUG] Reading alert action: %s", d.Id())
	ctx := context.Background()
	result := false
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetAlertAction(&ilert.GetAlertActionInput{AlertActionID: ilert.String(alertActionID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert alert action error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert action to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an alert action with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert alert action error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func flattenAlertActionAlertSourcesList(list []ilert.AlertSource) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, item := range list {
		result := make(map[string]interface{})
		result["id"] = strconv.FormatInt(item.ID, 10)
		results = append(results, result)
	}

	return results, nil
}

func flattenAlertActionAlertFilter(filter *ilert.AlertFilter) ([]interface{}, error) {
	if filter == nil {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	r := make(map[string]interface{})
	r["operator"] = filter.Operator
	prds := make([]interface{}, 0)
	for _, p := range filter.Predicates {
		prd := make(map[string]interface{})
		prd["field"] = p.Field
		prd["criteria"] = p.Criteria
		prd["value"] = p.Value
		prds = append(prds, prd)
	}
	r["predicate"] = prds
	results = append(results, r)

	return results, nil
}
