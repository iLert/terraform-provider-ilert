package ilert

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/iLert/ilert-go"
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
				ForceNew:     false,
				ValidateFunc: validateName,
			},
			"integration_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validateValueFunc([]string{
					"NAGIOS",
					"ICINGA",
					"EMAIL",
					"SMS",
					"API",
					"CRN",
					"HEARTBEAT",
					"PRTG",
					"PINGDOM",
					"CLOUDWATCH",
					"AWSPHD",
					"STACKDRIVER",
					"INSTANA",
					"ZABBIX",
					"SOLARWINDS",
					"PROMETHEUS",
					"NEWRELIC",
					"GRAFANA",
					"GITHUB",
					"DATADOG",
					"UPTIMEROBOT",
					"APPDYNAMICS",
					"DYNATRACE",
					"TOPDESK",
					"STATUSCAKE",
					"MONITOR",
					"TOOL",
					"CHECKMK",
					"AUTOTASK",
					"AWSBUDGET",
					"KENTIXAM",
				}),
			},
			"escalation_policy": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    false,
				Description: "The escalation policy specifies who will be notified when an incident is created by this alert source",
			},
			"incident_creation": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ONE_INCIDENT_PER_EMAIL",
				ValidateFunc: validateValueFunc([]string{
					"ONE_INCIDENT_PER_EMAIL",
					"ONE_INCIDENT_PER_EMAIL_SUBJECT",
					"ONE_PENDING_INCIDENT_ALLOWED",
					"ONE_OPEN_INCIDENT_ALLOWED",
					"OPEN_RESOLVE_ON_EXTRACTION",
				}),
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"incident_priority_rule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "HIGH",
				ValidateFunc: validateValueFunc([]string{
					"HIGH",
					"LOW",
					"HIGH_DURING_SUPPORT_HOURS",
					"LOW_DURING_SUPPORT_HOURS",
				}),
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
				ValidateFunc: validateValueFunc([]string{
					"AND",
					"OR",
				}),
			},
			"resolve_filter_operator": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "AND",
				ValidateFunc: validateValueFunc([]string{
					"AND",
					"OR",
				}),
			},
			"heartbeat": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				MinItems:    1,
				ForceNew:    true,
				Description: "A heartbeat alert source will automatically create an incident if it does not receive a heartbeat signal from your app at regular intervals.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"summary": {
							Type:        schema.TypeString,
							Description: "This text will be used as the incident summary, when incidents are created by this alert source",
							Optional:    true,
						},
						"interval_sec": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     900,
							Description: "The interval after which the heartbeat alert source will create an incident if it does not receive a ping",
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
				MinItems: 1,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timezone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"support_days": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"MONDAY": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"TUESDAY": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"WEDNESDAY": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"THURSDAY": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"FRIDAY": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"SATURDAY": {
										Type:     schema.TypeList,
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Elem:     getSupportDaySchemaResource(),
									},
									"SUNDAY": {
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
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
			"resolve_key_extractor": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"EMAIL_SUBJECT",
								"EMAIL_BODY",
							}),
						},
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"ALL_TEXT_BEFORE",
								"MATCHES_REGEX",
								"ALL_TEXT_AFTER",
							}),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"email_predicates": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"EMAIL_FROM",
								"EMAIL_SUBJECT",
								"EMAIL_BODY",
							}),
						},
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"CONTAINS_ANY_WORDS",
								"CONTAINS_NOT_WORDS",
								"CONTAINS_STRING",
								"CONTAINS_NOT_STRING",
								"IS_STRING",
								"IS_NOT_STRING",
								"MATCHES_REGEX",
								"MATCHES_NOT_REGEX",
							}),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"email_resolve_predicates": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"EMAIL_FROM",
								"EMAIL_SUBJECT",
								"EMAIL_BODY",
							}),
						},
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"CONTAINS_ANY_WORDS",
								"CONTAINS_NOT_WORDS",
								"CONTAINS_STRING",
								"CONTAINS_NOT_STRING",
								"IS_STRING",
								"IS_NOT_STRING",
								"MATCHES_REGEX",
								"MATCHES_NOT_REGEX",
							}),
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"icon_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"light_icon_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dark_icon_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Create: resourceAlertSourceCreate,
		Read:   resourceAlertSourceRead,
		Update: resourceAlertSourceUpdate,
		Delete: resourceAlertSourceDelete,
		Exists: resourceAlertSourceExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceAlertSourceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	alertSource := &ilert.AlertSource{
		Name:            d.Get("name").(string),
		IntegrationType: d.Get("integration_type").(string),
		EscalationPolicy: &ilert.EscalationPolicy{
			ID: d.Get("escalation_policy").(int64),
		},
	}

	log.Printf("[DEBUG] Creating iLert alert source %s", alertSource.Name)

	result, err := client.CreateAlertSource(&ilert.CreateAlertSourceInput{
		AlertSource: alertSource,
	})
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(result.AlertSource.ID, 10))
	return resourceAlertSourceRead(d, m)
}

func resourceAlertSourceRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	alertSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading alert source: %s", d.Id())
	result, err := client.GetAlertSource(&ilert.GetAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] Removing alert source %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Could not read an alert source with ID %s", d.Id())
	}

	d.Set("name", result.AlertSource.Name)
	d.Set("integration_type", result.AlertSource.IntegrationType)
	d.Set("escalation_policy", result.AlertSource.EscalationPolicy.ID)
	d.Set("incident_creation", result.AlertSource.IncidentCreation)
	d.Set("active", result.AlertSource.Active)
	d.Set("incident_priority_rule", result.AlertSource.IncidentPriorityRule)
	d.Set("email_filtered", result.AlertSource.EmailFiltered)
	d.Set("email_resolve_filtered", result.AlertSource.EmailResolveFiltered)
	d.Set("filter_operator", result.AlertSource.FilterOperator)
	d.Set("resolve_filter_operator", result.AlertSource.ResolveFilterOperator)
	d.Set("status", result.AlertSource.Status)
	d.Set("integration_key", result.AlertSource.IntegrationKey)
	d.Set("icon_url", result.AlertSource.IconURL)
	d.Set("light_icon_url", result.AlertSource.LightIconURL)
	d.Set("dark_icon_url", result.AlertSource.DarkIconURL)

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

	emailPredicates, err := flattenEmailPredicateList(result.AlertSource.EmailPredicates)
	if err != nil {
		return err
	}
	if err := d.Set("email_predicates", emailPredicates); err != nil {
		return fmt.Errorf("error setting email predicates: %s", err)
	}

	emailResolvePredicates, err := flattenEmailPredicateList(result.AlertSource.EmailResolvePredicates)
	if err != nil {
		return err
	}
	if err := d.Set("email_resolve_predicates", emailResolvePredicates); err != nil {
		return fmt.Errorf("error setting email resolve predicates: %s", err)
	}

	supportHours, err := flattenSupportHours(result.AlertSource.SupportHours)
	if err != nil {
		return err
	}
	if err := d.Set("support_hours", supportHours); err != nil {
		return fmt.Errorf("error setting support hours: %s", err)
	}

	return nil
}

func resourceAlertSourceUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	alertSource := &ilert.AlertSource{
		Name:            d.Get("name").(string),
		IntegrationType: d.Get("integration_type").(string),
		EscalationPolicy: &ilert.EscalationPolicy{
			ID: int64(d.Get("escalation_policy").(int)),
		},
	}
	if val, ok := d.GetOk("incident_creation"); ok {
		incidentCreation := val.(string)
		alertSource.IncidentCreation = incidentCreation
	}
	if val, ok := d.GetOk("active"); ok {
		active := val.(bool)
		alertSource.Active = active
	}
	if val, ok := d.GetOk("incident_priority_rule"); ok {
		incidentPriorityRule := val.(string)
		alertSource.IncidentPriorityRule = incidentPriorityRule
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
	if val, ok := d.GetOk("support_hours"); ok {
		vL := val.(*schema.Set).List()
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			supportHours := &ilert.SupportHours{
				Timezone: v["timezone"].(string),
			}
			sds := v["support_days"].(map[string]interface{})
			for d, sd := range sds {
				s := sd.(*schema.Set).List()
				if len(s) > 0 {
					v := s[0].(map[string]interface{})
					if d == "MONDAY" {
						supportHours.SupportDays.MONDAY = &ilert.SupportDay{
							Start: v["start"].(string),
							End:   v["end"].(string),
						}
					}
					if d == "TUESDAY" {
						supportHours.SupportDays.TUESDAY = &ilert.SupportDay{
							Start: v["start"].(string),
							End:   v["end"].(string),
						}
					}
					if d == "WEDNESDAY" {
						supportHours.SupportDays.WEDNESDAY = &ilert.SupportDay{
							Start: v["start"].(string),
							End:   v["end"].(string),
						}
					}
					if d == "THURSDAY" {
						supportHours.SupportDays.THURSDAY = &ilert.SupportDay{
							Start: v["start"].(string),
							End:   v["end"].(string),
						}
					}
					if d == "FRIDAY" {
						supportHours.SupportDays.FRIDAY = &ilert.SupportDay{
							Start: v["start"].(string),
							End:   v["end"].(string),
						}
					}
					if d == "SATURDAY" {
						supportHours.SupportDays.SATURDAY = &ilert.SupportDay{
							Start: v["start"].(string),
							End:   v["end"].(string),
						}
					}
					if d == "SUNDAY" {
						supportHours.SupportDays.SUNDAY = &ilert.SupportDay{
							Start: v["start"].(string),
							End:   v["end"].(string),
						}
					}
				}
			}
			alertSource.SupportHours = supportHours
		}
	}
	if val, ok := d.GetOk("heartbeat"); ok {
		vL := val.(*schema.Set).List()
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.Heartbeat = &ilert.Heartbeat{
				Summary:     v["summary"].(string),
				IntervalSec: v["interval_sec"].(int),
			}
		}
	}
	if val, ok := d.GetOk("autotask_metadata"); ok {
		vL := val.(*schema.Set).List()
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
		vL := val.(*schema.Set).List()
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			alertSource.ResolveKeyExtractor = &ilert.EmailPredicate{
				Field:    v["field"].(string),
				Criteria: v["criteria"].(string),
				Value:    v["value"].(string),
			}
		}
	}
	if val, ok := d.GetOk("email_predicates"); ok {
		vL := val.(*schema.Set).List()
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
	if val, ok := d.GetOk("email_resolve_predicates"); ok {
		vL := val.(*schema.Set).List()
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

	alertSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Updating alert source: %s", d.Id())
	_, err = client.UpdateAlertSource(&ilert.UpdateAlertSourceInput{AlertSource: alertSource, AlertSourceID: ilert.Int64(alertSourceID)})
	if err != nil {
		return err
	}
	return resourceAlertSourceRead(d, m)
}

func resourceAlertSourceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	alertSourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Deleting alert source: %s", d.Id())
	_, err = client.DeleteAlertSource(&ilert.DeleteAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID)})
	if err != nil {
		return err
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
	_, err = client.GetAlertSource(&ilert.GetAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

func flattenSupportHours(supportHours *ilert.SupportHours) ([]interface{}, error) {
	if supportHours == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)

	result := make(map[string]interface{})
	result["timezone"] = supportHours.Timezone

	supportDays := make(map[string]interface{})
	if supportHours.SupportDays.MONDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.MONDAY.Start
		supportDay["end"] = supportHours.SupportDays.MONDAY.End
		supportDays["MONDAY"] = supportDay
	}
	if supportHours.SupportDays.TUESDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.TUESDAY.Start
		supportDay["end"] = supportHours.SupportDays.TUESDAY.End
		supportDays["TUESDAY"] = supportDay
	}
	if supportHours.SupportDays.WEDNESDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.WEDNESDAY.Start
		supportDay["end"] = supportHours.SupportDays.WEDNESDAY.End
		supportDays["WEDNESDAY"] = supportDay
	}
	if supportHours.SupportDays.THURSDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.THURSDAY.Start
		supportDay["end"] = supportHours.SupportDays.THURSDAY.End
		supportDays["THURSDAY"] = supportDay
	}
	if supportHours.SupportDays.FRIDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.FRIDAY.Start
		supportDay["end"] = supportHours.SupportDays.FRIDAY.End
		supportDays["FRIDAY"] = supportDay
	}
	if supportHours.SupportDays.SATURDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.SATURDAY.Start
		supportDay["end"] = supportHours.SupportDays.SATURDAY.End
		supportDays["SATURDAY"] = supportDay
	}
	if supportHours.SupportDays.SUNDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.SUNDAY.Start
		supportDay["end"] = supportHours.SupportDays.SUNDAY.End
		supportDays["SUNDAY"] = supportDay
	}
	result["support_days"] = supportDays

	results = append(results, result)

	return results, nil
}
