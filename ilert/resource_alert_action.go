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
				Optional: true,
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
		vL := val.([]any)
		als := make([]ilert.AlertSource, 0)
		for _, m := range vL {
			v := m.(map[string]any)
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.ConnectorID = v["id"].(string)
			alertAction.ConnectorType = v["type"].(string)
		}
	}

	if val, ok := d.GetOk("trigger_mode"); ok {
		triggerMode := val.(string)
		alertAction.TriggerMode = triggerMode
	}

	if val, ok := d.GetOk("trigger_types"); ok {
		vL := val.([]any)
		sL := make([]string, 0)
		for _, m := range vL {
			v := m.(string)
			sL = append(sL, v)
		}
		alertAction.TriggerTypes = sL
	}

	if val, ok := d.GetOk("jira"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsJira{
				Project:      v["project"].(string),
				IssueType:    v["issue_type"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("servicenow"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsServiceNow{
				CallerID:     v["caller_id"].(string),
				Impact:       v["impact"].(string),
				Urgency:      v["urgency"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("slack"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsSlack{
				ChannelID:   v["channel_id"].(string),
				ChannelName: v["channel_name"].(string),
				TeamID:      v["team_id"].(string),
				TeamDomain:  v["team_domain"].(string),
			}
		}
	}

	if val, ok := d.GetOk("webhook"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsWebhook{
				WebhookURL:   v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("zendesk"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsZendesk{
				Priority: v["priority"].(string),
			}
		}
	}

	if val, ok := d.GetOk("github"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			params := &ilert.AlertActionParamsGithub{
				Owner:      v["owner"].(string),
				Repository: v["repository"].(string),
			}
			vL := v["labels"].([]any)
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsTopdesk{
				Status: v["status"].(string),
			}
		}
	}

	if val, ok := d.GetOk("email"); ok {
		if vL, ok := val.([]any); ok && len(vL) > 0 {
			if v, ok := vL[0].(map[string]any); ok && len(v) > 0 {
				params := &ilert.AlertActionParamsEmail{}
				if p, ok := v["subject"].(string); ok && p != "" {
					params.Subject = p
				}
				if p, ok := v["body_template"].(string); ok && p != "" {
					params.BodyTemplate = p
				}
				if vL, ok := v["recipients"].([]any); ok && len(vL) > 0 {
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsZammad{
				Email: v["email"].(string),
			}
		}
	}

	if val, ok := d.GetOk("dingtalk"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			params := &ilert.AlertActionParamsDingTalk{
				IsAtAll: v["is_at_all"].(bool),
			}
			vL := v["at_mobiles"].([]any)
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			params := &ilert.AlertActionParamsDingTalkAction{
				URL:     v["url"].(string),
				Secret:  v["secret"].(string),
				IsAtAll: v["is_at_all"].(bool),
			}
			vL := v["at_mobiles"].([]any)
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			params := &ilert.AlertActionParamsAutomationRule{
				AlertType:        v["alert_type"].(string),
				ResolveIncident:  v["resolve_incident"].(bool),
				ServiceStatus:    v["service_status"].(string),
				TemplateId:       int64(v["template_id"].(int)),
				SendNotification: v["send_notification"].(bool),
			}
			vL := v["service_ids"].([]any)
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsTelegram{
				ChannelID: v["channel_id"].(string),
			}
		}
	}

	if val, ok := d.GetOk("microsoft_teams_bot"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
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
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsMicrosoftTeamsWebhook{
				URL:          v["url"].(string),
				BodyTemplate: v["body_template"].(string),
			}
		}
	}

	if val, ok := d.GetOk("slack_webhook"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			alertAction.Params = &ilert.AlertActionParamsSlackWebhook{
				URL: v["url"].(string),
			}
		}
	}

	if val, ok := d.GetOk("alert_filter"); ok {
		vL := val.([]any)
		if len(vL) > 0 {
			v := vL[0].(map[string]any)
			filter := &ilert.AlertFilter{
				Operator: v["operator"].(string),
			}
			vL := v["predicate"].([]any)
			pL := make([]ilert.AlertFilterPredicate, 0)
			for _, m := range vL {
				v := m.(map[string]any)
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

func resourceAlertActionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func resourceAlertActionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertActionID := d.Id()
	log.Printf("[DEBUG] Reading alert action: %s", d.Id())

	result := &ilert.GetAlertActionOutput{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		version := 2
		if val, ok := d.GetOk("alert_source"); ok && len(val.([]any)) == 1 {
			if val, ok := d.GetOk("team"); !ok || len(val.([]any)) == 0 {
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

	err = transformAlertActionResource(result.AlertAction, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAlertActionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func resourceAlertActionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func resourceAlertActionExists(d *schema.ResourceData, m any) (bool, error) {
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

func transformAlertActionResource(alertAction *ilert.AlertActionOutput, d *schema.ResourceData) error {
	d.Set("name", alertAction.Name)

	if _, ok := d.GetOk("alert_source"); ok || d.Id() == "" {
		if alertAction.AlertSources != nil {
			alertSources, err := flattenAlertActionAlertSourcesList(*alertAction.AlertSources)
			if err != nil {
				return fmt.Errorf("[ERROR] Error flattening alert sources: %s", err.Error())
			}
			if err := d.Set("alert_source", alertSources); err != nil {
				return fmt.Errorf("[ERROR] Error setting alert sources: %s", err.Error())
			}
		}
	}

	connector := map[string]any{}
	if alertAction.ConnectorID != "" {
		connector["id"] = alertAction.ConnectorID
	}
	connector["type"] = alertAction.ConnectorType
	d.Set("connector", []any{connector})
	d.Set("trigger_mode", alertAction.TriggerMode)
	d.Set("trigger_types", alertAction.TriggerTypes)
	d.Set("created_at", alertAction.CreatedAt)
	d.Set("updated_at", alertAction.UpdatedAt)

	switch alertAction.ConnectorType {
	case ilert.ConnectorTypes.Jira:
		d.Set("jira", []any{
			map[string]any{
				"project":       alertAction.Params.Project,
				"issue_type":    alertAction.Params.IssueType,
				"body_template": alertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.ServiceNow:
		d.Set("servicenow", []any{
			map[string]any{
				"caller_id":     alertAction.Params.CallerID,
				"impact":        alertAction.Params.Impact,
				"urgency":       alertAction.Params.Urgency,
				"body_template": alertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Slack:
		d.Set("slack", []any{
			map[string]any{
				"channel_id":   alertAction.Params.ChannelID,
				"channel_name": alertAction.Params.ChannelName,
				"team_id":      alertAction.Params.TeamID,
				"team_domain":  alertAction.Params.TeamDomain,
			},
		})
	case ilert.ConnectorTypes.Webhook:
		d.Set("webhook", []any{
			map[string]any{
				"url":           alertAction.Params.WebhookURL,
				"body_template": alertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Zendesk:
		d.Set("zendesk", []any{
			map[string]any{
				"priority": alertAction.Params.Priority,
			},
		})
	case ilert.ConnectorTypes.Github:
		d.Set("github", []any{
			map[string]any{
				"owner":      alertAction.Params.Owner,
				"repository": alertAction.Params.Repository,
				"labels":     alertAction.Params.Labels,
			},
		})
	case ilert.ConnectorTypes.Topdesk:
		d.Set("topdesk", []any{
			map[string]any{
				"status": alertAction.Params.Status,
			},
		})
	case ilert.ConnectorTypes.Email:
		d.Set("email", []any{
			map[string]any{
				"recipients":    alertAction.Params.Recipients,
				"subject":       alertAction.Params.Subject,
				"body_template": alertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.Autotask:
		d.Set("autotask", []any{
			map[string]any{
				"company_id":      alertAction.Params.CompanyID,
				"issue_type":      alertAction.Params.IssueType,
				"queue_id":        alertAction.Params.QueueID,
				"ticket_category": alertAction.Params.TicketCategory,
				"ticket_type":     alertAction.Params.TicketType,
			},
		})
	case ilert.ConnectorTypes.Zammad:
		d.Set("zammad", []any{
			map[string]any{
				"email": alertAction.Params.Email,
			},
		})
	case ilert.ConnectorTypes.DingTalk:
		d.Set("dingtalk", []any{
			map[string]any{
				"is_at_all":  alertAction.Params.IsAtAll,
				"at_mobiles": alertAction.Params.AtMobiles,
			},
		})
	case ilert.ConnectorTypes.DingTalkAction:
		d.Set("dingtalk_action", []any{
			map[string]any{
				"url":        alertAction.Params.URL,
				"secret":     alertAction.Params.Secret,
				"is_at_all":  alertAction.Params.IsAtAll,
				"at_mobiles": alertAction.Params.AtMobiles,
			},
		})
	case ilert.ConnectorTypes.AutomationRule:
		d.Set("automation_rule", []any{
			map[string]any{
				"alert_type":        alertAction.Params.AlertType,
				"resolve_incident":  alertAction.Params.ResolveIncident,
				"service_status":    alertAction.Params.ServiceStatus,
				"template_id":       alertAction.Params.TemplateId,
				"send_notification": alertAction.Params.SendNotification,
				"service_ids":       alertAction.Params.ServiceIds,
			},
		})
	case ilert.ConnectorTypes.Telegram:
		d.Set("telegram", []any{
			map[string]any{
				"channel_id": alertAction.Params.ChannelID,
			},
		})
	case ilert.ConnectorTypes.MicrosoftTeamsBot:
		d.Set("microsoft_teams_bot", []any{
			map[string]any{
				"channel_id":   alertAction.Params.ChannelID,
				"channel_name": alertAction.Params.ChannelName,
				"team_id":      alertAction.Params.TeamID,
				"team_name":    alertAction.Params.TeamName,
				"type":         alertAction.Params.Type,
			},
		})
	case ilert.ConnectorTypes.MicrosoftTeamsWebhook:
		d.Set("microsoft_teams_webhook", []any{
			map[string]any{
				"url":          alertAction.Params.URL,
				"bodyTemplate": alertAction.Params.BodyTemplate,
			},
		})
	case ilert.ConnectorTypes.SlackWebhook:
		d.Set("slack_webhook", []any{
			map[string]any{
				"url": alertAction.Params.URL,
			},
		})
	}

	alertFilter, err := flattenAlertActionAlertFilter(alertAction.AlertFilter)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening alert filter: %s", err.Error())
	}
	if err := d.Set("alert_filter", alertFilter); err != nil {
		return fmt.Errorf("[ERROR] Error setting alert filter: %s", err.Error())
	}

	teams, err := flattenTeamShortList(*alertAction.Teams, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	d.Set("delay_sec", alertAction.DelaySec)
	d.Set("escalation_ended_delay_sec", alertAction.EscalationEndedDelaySec)
	d.Set("not_resolved_delay_sec", alertAction.NotResolvedDelaySec)

	if val, ok := d.GetOk("alert_source"); ok && len(val.([]any)) == 1 && d.Id() != "" {
		if v, ok := d.GetOk("team"); !ok || len(v.([]any)) == 0 {
			sourceId := alertAction.AlertSourceIDs[0]

			sources := make([]any, 0)
			source := make(map[string]any)
			source["id"] = strconv.FormatInt(sourceId, 10)
			sources = append(sources, source)

			d.Set("alert_source", sources)
		}
	}

	d.Set("conditions", alertAction.Conditions)

	return nil
}

func flattenAlertActionAlertSourcesList(list []ilert.AlertSource) ([]any, error) {
	if list == nil {
		return make([]any, 0), nil
	}
	results := make([]any, 0)
	for _, item := range list {
		result := make(map[string]any)
		result["id"] = strconv.FormatInt(item.ID, 10)
		results = append(results, result)
	}

	return results, nil
}

func flattenAlertActionAlertFilter(filter *ilert.AlertFilter) ([]any, error) {
	if filter == nil {
		return make([]any, 0), nil
	}

	results := make([]any, 0)
	r := make(map[string]any)
	r["operator"] = filter.Operator
	prds := make([]any, 0)
	for _, p := range filter.Predicates {
		prd := make(map[string]any)
		prd["field"] = p.Field
		prd["criteria"] = p.Criteria
		prd["value"] = p.Value
		prds = append(prds, prd)
	}
	r["predicate"] = prds
	results = append(results, r)

	return results, nil
}
