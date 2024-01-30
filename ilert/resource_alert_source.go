package ilert

import (
	"context"
	"errors"
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

func getSupportDaySchemaResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"start": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "08:00",
			},
			"end": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "17:00",
			},
		},
	}
}

func resourceAlertSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the resource, also acts as it's unique ID",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"integration_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ilert.AlertSourceIntegrationTypesAll, false),
			},
			"escalation_policy": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The escalation policy specifies who will be notified when an alert is created by this alert source",
			},
			"incident_creation": { // @deprecated
				Deprecated: "The field incident_creation is deprecated! Please use alert_creation instead.",
				Type:       schema.TypeString,
				Optional:   true,
				ValidateFunc: validation.StringInSlice([]string{
					"ONE_INCIDENT_PER_EMAIL",
					"ONE_INCIDENT_PER_EMAIL_SUBJECT",
					"ONE_PENDING_INCIDENT_ALLOWED",
					"ONE_OPEN_INCIDENT_ALLOWED",
					"OPEN_RESOLVE_ON_EXTRACTION",
				}, false),
			},
			"alert_creation": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ONE_ALERT_PER_EMAIL",
				ValidateFunc: validation.StringInSlice(ilert.AlertSourceAlertCreationsAll, false),
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"incident_priority_rule": { // @deprecated
				Deprecated: "The field incident_priority_rule is deprecated! Please use alert_priority_rule instead.",
				Type:       schema.TypeString,
				Optional:   true,
				ValidateFunc: validation.StringInSlice([]string{
					"HIGH",
					"LOW",
					"HIGH_DURING_SUPPORT_HOURS",
					"LOW_DURING_SUPPORT_HOURS",
				}, false),
			},
			"alert_priority_rule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "HIGH",
				ValidateFunc: validation.StringInSlice([]string{
					"HIGH",
					"LOW",
					"HIGH_DURING_SUPPORT_HOURS",
					"LOW_DURING_SUPPORT_HOURS",
				}, false),
			},
			"auto_resolution_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"PT10M",
					"PT20M",
					"PT30M",
					"PT40M",
					"PT50M",
					"PT60M",
					"PT90M",
					"PT2H",
					"PT3H",
					"PT4H",
					"PT5H",
					"PT6H",
					"PT12H",
					"PT24H",
				}, false),
			},
			"email_filtered": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"email_resolve_filtered": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"filter_operator": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "AND",
				ValidateFunc: validation.StringInSlice([]string{
					"AND",
					"OR",
				}, false),
			},
			"resolve_filter_operator": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "AND",
				ValidateFunc: validation.StringInSlice([]string{
					"AND",
					"OR",
				}, false),
			},
			"teams": {
				Type:       schema.TypeList,
				Optional:   true,
				Deprecated: "The field teams is deprecated! Please use team instead.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
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
			"heartbeat": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				MinItems:    1,
				ForceNew:    true,
				Description: "A heartbeat alert source will automatically create an alert if it does not receive a heartbeat signal from your app at regular intervals.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"summary": {
							Type:        schema.TypeString,
							Description: "This text will be used as the alert summary, when alerts are created by this alert source",
							Optional:    true,
						},
						"interval_sec": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  900,
							ValidateFunc: validation.IntInSlice([]int{
								1 * 60,
								5 * 60,
								10 * 60,
								15 * 60,
								30 * 60,
								60 * 60,
								24 * 60 * 60,
								7 * 24 * 60 * 60,
								30 * 24 * 60 * 60,
							}),
							Description: "The interval after which the heartbeat alert source will create an alert if it does not receive a ping",
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"support_hours": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:          schema.TypeInt,
							Optional:      true,
							ConflictsWith: []string{"support_hours.0.timezone", "support_hours.0.auto_raise_incidents", "support_hours.0.auto_raise_alerts", "support_hours.0.support_days"},
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"timezone": {
							Type:         schema.TypeString,
							Deprecated:   "The field `timezone` is deprecated! Please use the support hour resource instead and reference it via field `id`.",
							Optional:     true,
							RequiredWith: []string{"support_hours.0.support_days"},
						},
						"auto_raise_incidents": { // @deprecated
							Deprecated: "The field `auto_raise_incidents` is deprecated! Please use auto_raise_alerts instead.",
							Type:       schema.TypeBool,
							Optional:   true,
						},
						"auto_raise_alerts": {
							Type:       schema.TypeBool,
							Optional:   true,
							Deprecated: "The field `auto_raise_alerts` is deprecated! Please use the support hour resource instead and reference it via field `id`."},
						"support_days": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							MinItems:     1,
							ForceNew:     false,
							RequiredWith: []string{"support_hours.0.timezone"},
							Deprecated:   "The field `support_days` is deprecated! Please use the support hour resource instead and reference it via field `id`.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"monday": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"tuesday": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"wednesday": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"thursday": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"friday": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"saturday": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"sunday": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
								},
							},
						},
					},
				},
			},
			"autotask_metadata": {
				Type:       schema.TypeList,
				Deprecated: "The field autotask_metadata is deprecated! Please use the web UI to configure autotask metadata.",
				Optional:   true,
				MaxItems:   1,
				MinItems:   1,
				ForceNew:   true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secret": {
							Type:     schema.TypeString,
							Required: true,
						},
						"web_server": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "https://webservices2.autotask.net",
						},
					},
				},
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resolve_key_extractor": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"EMAIL_SUBJECT",
								"EMAIL_BODY",
							}, false),
						},
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"ALL_TEXT_BEFORE",
								"ALL_TEXT_AFTER",
								"MATCHES_REGEX",
							}, false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"email_predicate": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"EMAIL_FROM",
								"EMAIL_SUBJECT",
								"EMAIL_BODY",
							}, false),
						},
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"CONTAINS_ANY_WORDS",
								"CONTAINS_NOT_WORDS",
								"CONTAINS_STRING",
								"CONTAINS_NOT_STRING",
								"IS_STRING",
								"IS_NOT_STRING",
								"MATCHES_REGEX",
								"MATCHES_NOT_REGEX",
							}, false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"email_resolve_predicate": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"EMAIL_FROM",
								"EMAIL_SUBJECT",
								"EMAIL_BODY",
							}, false),
						},
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"CONTAINS_ANY_WORDS",
								"CONTAINS_NOT_WORDS",
								"CONTAINS_STRING",
								"CONTAINS_NOT_STRING",
								"IS_STRING",
								"IS_NOT_STRING",
								"MATCHES_REGEX",
								"MATCHES_NOT_REGEX",
							}, false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"integration_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"summary_template": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"text_template": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"details_template": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"text_template": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"routing_template": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"text_template": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"link_template": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"text": {
							Type:     schema.TypeString,
							Required: true,
						},
						"href_template": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text_template": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"priority_template": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value_template": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text_template": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"mapping": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
									"priority": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(ilert.AlertPrioritiesAll, false),
									},
								},
							},
						},
					},
				},
			},
			"alert_grouping_window": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ilert.AlertSourceAlertGroupingWindowsAll, false),
			},
		},
		CreateContext: resourceAlertSourceCreate,
		ReadContext:   resourceAlertSourceRead,
		UpdateContext: resourceAlertSourceUpdate,
		DeleteContext: resourceAlertSourceDelete,
		Exists:        resourceAlertSourceExists,
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

func buildAlertSource(d *schema.ResourceData) (*ilert.AlertSource, error) {
	escalationPolicyID, err := strconv.ParseInt(d.Get("escalation_policy").(string), 10, 64)
	if err != nil {
		return nil, unconvertibleIDErr(d.Id(), err)
	}
	name := d.Get("name").(string)
	integrationType := d.Get("integration_type").(string)

	alertSource := &ilert.AlertSource{
		Name:            name,
		IntegrationType: integrationType,
		EscalationPolicy: &ilert.EscalationPolicy{
			ID: escalationPolicyID,
		},
	}

	if integrationType == "EMAIL" {
		if val, ok := d.GetOk("email"); ok {
			email := val.(string)
			alertSource.IntegrationKey = email
		} else {
			return nil, errors.New("email is required")
		}
	}
	if val, ok := d.GetOk("incident_creation"); ok {
		incidentCreation := val.(string)
		alertSource.IncidentCreation = incidentCreation
	}
	if val, ok := d.GetOk("alert_creation"); ok {
		alertCreation := val.(string)
		if _, ok := d.GetOk("alert_grouping_window"); !ok && alertCreation == ilert.AlertSourceAlertCreations.OneAlertGroupedPerWindow {
			return nil, fmt.Errorf("[ERROR] Can't set alert creation type 'ONE_ALERT_GROUPED_PER_WINDOW' when alert grouping window is not set")
		}
		alertSource.AlertCreation = alertCreation
	}
	if val, ok := d.GetOk("integration_key"); ok {
		integrationKey := val.(string)
		alertSource.IntegrationKey = integrationKey
	}
	if val, ok := d.GetOk("integration_url"); ok {
		integrationURL := val.(string)
		alertSource.IntegrationURL = integrationURL
	}
	if val, ok := d.GetOk("active"); ok {
		active := val.(bool)
		alertSource.Active = active
	}
	if val, ok := d.GetOk("incident_priority_rule"); ok {
		incidentPriorityRule := val.(string)
		alertSource.IncidentPriorityRule = incidentPriorityRule
	}
	if val, ok := d.GetOk("alert_priority_rule"); ok {
		alertPriorityRule := val.(string)
		alertSource.AlertPriorityRule = alertPriorityRule
	}
	if val, ok := d.GetOk("auto_resolution_timeout"); ok {
		autoResolutionTimeout := val.(string)
		alertSource.AutoResolutionTimeout = autoResolutionTimeout
	}
	if val, ok := d.GetOk("email_filtered"); ok {
		emailFiltered := val.(bool)
		alertSource.EmailFiltered = emailFiltered
	}
	if val, ok := d.GetOk("email_resolve_filtered"); ok {
		emailResolveFiltered := val.(bool)
		alertSource.EmailResolveFiltered = emailResolveFiltered
	}
	if val, ok := d.GetOk("filter_operator"); ok {
		filterOperator := val.(string)
		alertSource.FilterOperator = filterOperator
	}
	if val, ok := d.GetOk("resolve_filter_operator"); ok {
		resolveFilterOperator := val.(string)
		alertSource.ResolveFilterOperator = resolveFilterOperator
	}
	if val, ok := d.GetOk("teams"); ok {
		vL := val.([]interface{})
		tms := make([]ilert.TeamShort, 0)

		for _, m := range vL {
			v := int64(m.(int))
			tms = append(tms, ilert.TeamShort{ID: v})
		}
		alertSource.Teams = tms
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
		alertSource.Teams = tms
	}
	if val, ok := d.GetOk("support_hours"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 && vL[0] != nil {
			v := vL[0].(map[string]interface{})
			if id := int64(v["id"].(int)); id > 0 {
				if err != nil {
					log.Printf("[ERROR] Could not parse support hours id %s", err.Error())
					return nil, err
				}
				alertSource.SupportHours = &ilert.SupportHour{
					ID: id,
				}
			} else {
				// legacy
				supportHours := &ilert.SupportHours{
					Timezone:           v["timezone"].(string),
					AutoRaiseIncidents: v["auto_raise_incidents"].(bool),
					AutoRaiseAlerts:    v["auto_raise_alerts"].(bool),
				}
				sdA := v["support_days"].([]interface{})
				if len(sdA) > 0 {
					if sdA[0] == nil {
						return nil, fmt.Errorf("[ERROR] Can't set support hours, support days needs at least one day to be defined")
					}
					sds := sdA[0].(map[string]interface{})
					for d, sd := range sds {
						s := sd.([]interface{})
						if len(s) > 0 && s[0] != nil {
							v := s[0].(map[string]interface{})
							if d == "monday" {
								supportHours.SupportDays.MONDAY = &ilert.SupportDay{
									Start: v["start"].(string),
									End:   v["end"].(string),
								}
							}
							if d == "tuesday" {
								supportHours.SupportDays.TUESDAY = &ilert.SupportDay{
									Start: v["start"].(string),
									End:   v["end"].(string),
								}
							}
							if d == "wednesday" {
								supportHours.SupportDays.WEDNESDAY = &ilert.SupportDay{
									Start: v["start"].(string),
									End:   v["end"].(string),
								}
							}
							if d == "thursday" {
								supportHours.SupportDays.THURSDAY = &ilert.SupportDay{
									Start: v["start"].(string),
									End:   v["end"].(string),
								}
							}
							if d == "friday" {
								supportHours.SupportDays.FRIDAY = &ilert.SupportDay{
									Start: v["start"].(string),
									End:   v["end"].(string),
								}
							}
							if d == "saturday" {
								supportHours.SupportDays.SATURDAY = &ilert.SupportDay{
									Start: v["start"].(string),
									End:   v["end"].(string),
								}
							}
							if d == "sunday" {
								supportHours.SupportDays.SUNDAY = &ilert.SupportDay{
									Start: v["start"].(string),
									End:   v["end"].(string),
								}
							}
						}
					}
				}
				alertSource.SupportHours = supportHours
			}
		}
	}
	if val, ok := d.GetOk("heartbeat"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.Heartbeat = &ilert.Heartbeat{
				Summary:     v["summary"].(string),
				IntervalSec: v["interval_sec"].(int),
			}
		}
	}
	if val, ok := d.GetOk("autotask_metadata"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.AutotaskMetadata = &ilert.AutotaskMetadata{
				Username:  v["username"].(string),
				Secret:    v["secret"].(string),
				WebServer: v["web_server"].(string),
			}
		}
	}
	if val, ok := d.GetOk("resolve_key_extractor"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.ResolveKeyExtractor = &ilert.EmailPredicate{
				Field:    v["field"].(string),
				Criteria: v["criteria"].(string),
				Value:    v["value"].(string),
			}
		}
	}
	if val, ok := d.GetOk("email_predicate"); ok {
		vL := val.([]interface{})
		eps := make([]ilert.EmailPredicate, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			ep := ilert.EmailPredicate{
				Field:    v["field"].(string),
				Criteria: v["criteria"].(string),
				Value:    v["value"].(string),
			}
			eps = append(eps, ep)
		}
		alertSource.EmailPredicates = eps
	}
	if val, ok := d.GetOk("email_resolve_predicate"); ok {
		vL := val.([]interface{})
		eps := make([]ilert.EmailPredicate, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			ep := ilert.EmailPredicate{
				Field:    v["field"].(string),
				Criteria: v["criteria"].(string),
				Value:    v["value"].(string),
			}
			eps = append(eps, ep)
		}
		alertSource.EmailResolvePredicates = eps
	}
	if val, ok := d.GetOk("summary_template"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.SummaryTemplate = &ilert.Template{
				TextTemplate: v["text_template"].(string),
			}
		}
	}
	if val, ok := d.GetOk("details_template"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.DetailsTemplate = &ilert.Template{
				TextTemplate: v["text_template"].(string),
			}
		}
	}
	if val, ok := d.GetOk("routing_template"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.RoutingTemplate = &ilert.Template{
				TextTemplate: v["text_template"].(string),
			}
		}
	}
	if val, ok := d.GetOk("link_template"); ok {
		vL := val.([]interface{})
		ltmps := make([]ilert.LinkTemplate, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			ltmp := ilert.LinkTemplate{}
			if v["text"] != nil && v["text"].(string) != "" {
				ltmp.Text = v["text"].(string)
			}
			htmp := ilert.Template{}
			if v["href_template"] != nil && len(v["href_template"].([]interface{})) > 0 {
				htL := v["href_template"].([]interface{})
				h := htL[0].(map[string]interface{})
				htmp.TextTemplate = h["text_template"].(string)

			}
			ltmp.HrefTemplate = &htmp
			ltmps = append(ltmps, ltmp)
		}
		alertSource.LinkTemplates = ltmps
	}
	if val, ok := d.GetOk("priority_template"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			ptmp := ilert.PriorityTemplate{}

			vtmp := ilert.Template{}
			if v["value_template"] != nil && len(v["value_template"].([]interface{})) > 0 {
				htL := v["value_template"].([]interface{})
				h := htL[0].(map[string]interface{})
				vtmp.TextTemplate = h["text_template"].(string)
			}
			ptmp.ValueTemplate = &vtmp

			if v["mapping"] != nil {
				mL := v["mapping"].([]interface{})
				mpgs := make([]ilert.Mapping, 0)

				for _, m := range mL {
					mpg := ilert.Mapping{}
					mp := m.(map[string]interface{})
					if mp["value"] != nil && mp["value"].(string) != "" {
						mpg.Value = mp["value"].(string)
					}
					if mp["priority"] != nil && mp["priority"].(string) != "" {
						mpg.Priority = mp["priority"].(string)
					}
					mpgs = append(mpgs, mpg)
				}
				ptmp.Mappings = mpgs
			}
			alertSource.PriorityTemplate = &ptmp
		}
	}
	if val, ok := d.GetOk("alert_grouping_window"); ok {
		if alert_creation, ok := d.GetOk("alert_creation"); !ok || alert_creation.(string) != ilert.AlertSourceAlertCreations.OneAlertGroupedPerWindow {
			return nil, fmt.Errorf("[ERROR] Can't set alert grouping window when alert creation is not set or not of type 'ONE_ALERT_GROUPED_PER_WINDOW'")
		}
		alertGroupingWindow := val.(string)
		alertSource.AlertGroupingWindow = alertGroupingWindow
	}

	return alertSource, nil
}

func resourceAlertSourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertSource, err := buildAlertSource(d)
	if err != nil {
		log.Printf("[ERROR] Building alert source error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Creating ilert alert source %s\n", alertSource.Name)
	result := &ilert.CreateAlertSourceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateAlertSource(&ilert.CreateAlertSourceInput{AlertSource: alertSource})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert alert source error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source, %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create an alert source with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert alert source error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.AlertSource == nil {
		log.Printf("[ERROR] Creating ilert alert source error: empty response")
		return diag.Errorf("alert source response is empty")
	}
	d.SetId(strconv.FormatInt(result.AlertSource.ID, 10))
	return resourceAlertSourceRead(ctx, d, m)
}

func resourceAlertSourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading alert source: %s", d.Id())
	result := &ilert.GetAlertSourceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		includes := make([]*string, 0)
		includes = append(includes, ilert.String("summaryTemplate"), ilert.String("detailsTemplate"), ilert.String("routingTemplate"), ilert.String("textTemplate"), ilert.String("linkTemplates"), ilert.String("priorityTemplate"))
		r, err := client.GetAlertSource(&ilert.GetAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID), Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing alert source %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an alert source with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert alert source error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.AlertSource == nil {
		log.Printf("[ERROR] Reading ilert alert source error: empty response")
		return diag.Errorf("alert source response is empty")
	}

	d.Set("name", result.AlertSource.Name)
	d.Set("integration_type", result.AlertSource.IntegrationType)
	d.Set("escalation_policy", strconv.FormatInt(result.AlertSource.EscalationPolicy.ID, 10))
	d.Set("incident_creation", result.AlertSource.IncidentCreation)
	d.Set("alert_creation", result.AlertSource.AlertCreation)
	d.Set("active", result.AlertSource.Active)
	d.Set("incident_priority_rule", result.AlertSource.IncidentPriorityRule)
	d.Set("alert_priority_rule", result.AlertSource.AlertPriorityRule)
	d.Set("auto_resolution_timeout", result.AlertSource.AutoResolutionTimeout)
	d.Set("email_filtered", result.AlertSource.EmailFiltered)
	d.Set("email_resolve_filtered", result.AlertSource.EmailResolveFiltered)
	d.Set("filter_operator", result.AlertSource.FilterOperator)
	d.Set("resolve_filter_operator", result.AlertSource.ResolveFilterOperator)
	d.Set("status", result.AlertSource.Status)
	d.Set("integration_key", result.AlertSource.IntegrationKey)
	d.Set("integration_url", result.AlertSource.IntegrationURL)
	if result.AlertSource.IntegrationType == "EMAIL" {
		d.Set("email", result.AlertSource.IntegrationKey)
	}
	d.Set("alert_grouping_window", result.AlertSource.AlertGroupingWindow)

	if result.AlertSource.Heartbeat != nil {
		d.Set("heartbeat", []interface{}{
			map[string]interface{}{
				"summary":      result.AlertSource.Heartbeat.Summary,
				"interval_sec": result.AlertSource.Heartbeat.IntervalSec,
				"status":       result.AlertSource.Heartbeat.Status,
			},
		})
	} else {
		d.Set("heartbeat", []interface{}{})
	}

	if result.AlertSource.AutotaskMetadata != nil {
		d.Set("autotask_metadata", []interface{}{
			map[string]interface{}{
				"username":   result.AlertSource.AutotaskMetadata.Username,
				"secret":     result.AlertSource.AutotaskMetadata.Secret,
				"web_server": result.AlertSource.AutotaskMetadata.WebServer,
			},
		})
	} else {
		d.Set("autotask_metadata", []interface{}{})
	}

	if result.AlertSource.ResolveKeyExtractor != nil {
		d.Set("resolve_key_extractor", []interface{}{
			map[string]interface{}{
				"field":    result.AlertSource.ResolveKeyExtractor.Field,
				"criteria": result.AlertSource.ResolveKeyExtractor.Criteria,
				"value":    result.AlertSource.ResolveKeyExtractor.Value,
			},
		})
	} else {
		d.Set("resolve_key_extractor", []interface{}{})
	}

	if result.AlertSource.SummaryTemplate != nil {
		d.Set("summary_template", []interface{}{
			map[string]interface{}{
				"text_template": result.AlertSource.SummaryTemplate.TextTemplate,
			},
		})
	} else {
		d.Set("summary_template", []interface{}{})
	}

	if result.AlertSource.DetailsTemplate != nil {
		d.Set("details_template", []interface{}{
			map[string]interface{}{
				"text_template": result.AlertSource.DetailsTemplate.TextTemplate,
			},
		})
	} else {
		d.Set("details_template", []interface{}{})
	}

	if result.AlertSource.RoutingTemplate != nil {
		d.Set("routing_template", []interface{}{
			map[string]interface{}{
				"text_template": result.AlertSource.RoutingTemplate.TextTemplate,
			},
		})
	} else {
		d.Set("routing_template", []interface{}{})
	}

	linkTemplates, err := flattenLinkTemplatesList(result.AlertSource.LinkTemplates)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("link_template", linkTemplates); err != nil {
		return diag.FromErr(fmt.Errorf("error setting link templates: %s", err))
	}

	priorityTemplate, err := flattenPriorityTemplate(result.AlertSource.PriorityTemplate)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("priority_template", priorityTemplate); err != nil {
		return diag.FromErr(fmt.Errorf("error setting priority template: %s", err))
	}

	if val, ok := d.GetOk("team"); ok {
		if val != nil {
			vL := val.([]interface{})
			teams := make([]interface{}, 0)
			for i, item := range result.AlertSource.Teams {
				team := make(map[string]interface{})
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

	if val, ok := d.GetOk("teams"); ok {
		if val != nil {
			teams := make([]interface{}, 0)
			for _, item := range result.AlertSource.Teams {
				team := make(map[string]interface{})
				team["id"] = item.ID
				teams = append(teams, team)
			}
			if err := d.Set("team", teams); err != nil {
				return diag.Errorf("error setting teams: %s", err)
			}

			d.Set("teams", nil)
		}
	}

	emailPredicates, err := flattenEmailPredicateList(result.AlertSource.EmailPredicates)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email_predicate", emailPredicates); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email predicates: %s", err))
	}

	emailResolvePredicates, err := flattenEmailPredicateList(result.AlertSource.EmailResolvePredicates)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email_resolve_predicate", emailResolvePredicates); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email resolve predicates: %s", err))
	}

	// never set support hours when user doesn't define them, even if server returns some
	if _, ok := d.GetOk("support_hours"); ok {
		supportHours, err := flattenSupportHoursInterface(result.AlertSource.SupportHours)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("support_hours", supportHours); err != nil {
			return diag.FromErr(fmt.Errorf("error setting support hours: %s", err))
		}
	}

	return nil
}

func resourceAlertSourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertSource, err := buildAlertSource(d)
	if err != nil {
		log.Printf("[ERROR] Building alert source error %s", err.Error())
		return diag.FromErr(err)
	}

	alertSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating alert source: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateAlertSource(&ilert.UpdateAlertSourceInput{AlertSource: alertSource, AlertSourceID: ilert.Int64(alertSourceID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an alert source with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert alert source error %s", err.Error())
		return diag.FromErr(err)
	}
	return resourceAlertSourceRead(ctx, d, m)
}

func resourceAlertSourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting alert source: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteAlertSource(&ilert.DeleteAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an alert source with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert alert source error %s", err.Error())
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceAlertSourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	alertSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading alert source: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		includes := make([]*string, 0)
		includes = append(includes, ilert.String("summaryTemplate"), ilert.String("detailsTemplate"), ilert.String("routingTemplate"), ilert.String("textTemplate"), ilert.String("linkTemplates"), ilert.String("priorityTemplate"))
		_, err := client.GetAlertSource(&ilert.GetAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID), Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert alert source error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an alert source with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert alert source error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func flattenEmailPredicateList(predicateList []ilert.EmailPredicate) ([]interface{}, error) {
	if predicateList == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, predicate := range predicateList {
		result := make(map[string]interface{})
		result["criteria"] = predicate.Criteria
		result["field"] = predicate.Field
		result["value"] = predicate.Value
		results = append(results, result)
	}

	return results, nil
}

func flattenSupportHoursInterface(supportHoursInterface interface{}) ([]interface{}, error) {
	supportHoursMap, ok := supportHoursInterface.(map[string]interface{})
	if !ok {
		return make([]interface{}, 0), nil
	}

	if supportHoursMap["id"] != nil {
		supportHours, err := flattenSupportHours(supportHoursMap)
		if err != nil {
			return nil, err
		}
		return supportHours, nil
	}

	supportHours, err := flattenSupportHoursLegacy(supportHoursMap)
	if err != nil {
		return nil, err
	}
	return supportHours, nil
}

func flattenSupportHours(supportHours map[string]interface{}) ([]interface{}, error) {
	if supportHours == nil {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	result := make(map[string]interface{})

	result["id"] = supportHours["id"]

	results = append(results, result)

	return results, nil
}

func flattenSupportHoursLegacy(supportHours map[string]interface{}) ([]interface{}, error) {
	if supportHours == nil {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	result := make(map[string]interface{})

	result["timezone"] = supportHours["timezone"]
	result["auto_raise_incidents"] = supportHours["autoRaiseIncidents"]
	result["auto_raise_alerts"] = supportHours["autoRaiseAlerts"]

	supportDaysMap, ok := supportHours["supportDays"].(map[string]interface{})
	if !ok {
		return make([]interface{}, 0), nil
	}
	supportDays, err := flattenSupportDaysMap(supportDaysMap)
	if err != nil {
		return nil, err
	}
	result["support_days"] = supportDays

	results = append(results, result)

	return results, nil
}

func flattenSupportDaysMap(supportDays map[string]interface{}) ([]interface{}, error) {
	results := make([]interface{}, 0)
	result := make(map[string]interface{})

	if supportDays["MONDAY"] != nil {
		supportDay := make(map[string]interface{})
		day := supportDays["MONDAY"].(map[string]interface{})
		supportDay["start"] = day["start"]
		supportDay["end"] = day["end"]
		result["monday"] = []interface{}{supportDay}
	}
	if supportDays["TUESDAY"] != nil {
		supportDay := make(map[string]interface{})
		day := supportDays["TUESDAY"].(map[string]interface{})
		supportDay["start"] = day["start"]
		supportDay["end"] = day["end"]
		result["tuesday"] = []interface{}{supportDay}
	}
	if supportDays["WEDNESDAY"] != nil {
		supportDay := make(map[string]interface{})
		day := supportDays["WEDNESDAY"].(map[string]interface{})
		supportDay["start"] = day["start"]
		supportDay["end"] = day["end"]
		result["wednesday"] = []interface{}{supportDay}
	}
	if supportDays["THURSDAY"] != nil {
		supportDay := make(map[string]interface{})
		day := supportDays["THURSDAY"].(map[string]interface{})
		supportDay["start"] = day["start"]
		supportDay["end"] = day["end"]
		result["thursday"] = []interface{}{supportDay}
	}
	if supportDays["FRIDAY"] != nil {
		supportDay := make(map[string]interface{})
		day := supportDays["FRIDAY"].(map[string]interface{})
		supportDay["start"] = day["start"]
		supportDay["end"] = day["end"]
		result["friday"] = []interface{}{supportDay}
	}
	if supportDays["SATURDAY"] != nil {
		supportDay := make(map[string]interface{})
		day := supportDays["SATURDAY"].(map[string]interface{})
		supportDay["start"] = day["start"]
		supportDay["end"] = day["end"]
		result["saturday"] = []interface{}{supportDay}
	}
	if supportDays["SUNDAY"] != nil {
		supportDay := make(map[string]interface{})
		day := supportDays["SUNDAY"].(map[string]interface{})
		supportDay["start"] = day["start"]
		supportDay["end"] = day["end"]
		result["sunday"] = []interface{}{supportDay}
	}
	results = append(results, result)
	return results, nil
}

func flattenLinkTemplatesList(linkTemplatesList []ilert.LinkTemplate) ([]interface{}, error) {
	if linkTemplatesList == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, linkTemplate := range linkTemplatesList {
		result := make(map[string]interface{})
		result["text"] = linkTemplate.Text

		hrefTemplates := make([]interface{}, 0)
		hrefTemplate := make(map[string]interface{})
		hrefTemplate["text_template"] = linkTemplate.HrefTemplate.TextTemplate

		hrefTemplates = append(hrefTemplates, hrefTemplate)
		result["href_template"] = hrefTemplates

		results = append(results, result)
	}

	return results, nil
}

func flattenPriorityTemplate(priorityTemplate *ilert.PriorityTemplate) ([]interface{}, error) {
	if priorityTemplate == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)

	result := make(map[string]interface{})

	valueTemplates := make([]interface{}, 0)
	valueTemplate := make(map[string]interface{})
	valueTemplate["text_template"] = priorityTemplate.ValueTemplate.TextTemplate

	valueTemplates = append(valueTemplates, valueTemplate)
	result["value_template"] = valueTemplates

	mappings := make([]interface{}, 0)
	for _, priorityMapping := range priorityTemplate.Mappings {
		mapping := make(map[string]interface{})

		mapping["value"] = priorityMapping.Value
		mapping["priority"] = priorityMapping.Priority

		mappings = append(mappings, mapping)
	}

	result["mapping"] = mappings

	results = append(results, result)

	return results, nil
}
