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
				Description: "The escalation policy specifies who will be notified when an incident is created by this alert source",
			},
			"incident_creation": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ONE_INCIDENT_PER_EMAIL",
				ValidateFunc: validation.StringInSlice([]string{
					"ONE_INCIDENT_PER_EMAIL",
					"ONE_INCIDENT_PER_EMAIL_SUBJECT",
					"ONE_PENDING_INCIDENT_ALLOWED",
					"ONE_OPEN_INCIDENT_ALLOWED",
					"OPEN_RESOLVE_ON_EXTRACTION",
				}, false),
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
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
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
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timezone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"auto_raise_incidents": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"support_days": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							MinItems: 1,
							ForceNew: false,
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
	if val, ok := d.GetOk("support_hours"); ok {
		vL := val.([]interface{})
		if len(vL) > 0 {
			v := vL[0].(map[string]interface{})
			supportHours := &ilert.SupportHours{
				Timezone:           v["timezone"].(string),
				AutoRaiseIncidents: v["auto_raise_incidents"].(bool),
			}
			sdA := v["support_days"].([]interface{})
			if len(vL) > 0 {
				sds := sdA[0].(map[string]interface{})
				for d, sd := range sds {
					s := sd.([]interface{})
					if len(s) > 0 {
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

	return alertSource, nil
}

func resourceAlertSourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	alertSource, err := buildAlertSource(d)
	if err != nil {
		log.Printf("[ERROR] Building alert source error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Creating iLert alert source %s\n", alertSource.Name)
	result := &ilert.CreateAlertSourceOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateAlertSource(&ilert.CreateAlertSourceInput{AlertSource: alertSource})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be created", d.Id()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating iLert alert source error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.AlertSource == nil {
		log.Printf("[ERROR] Creating iLert alert source error: empty response ")
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
		r, err := client.GetAlertSource(&ilert.GetAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing alert source %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an alert source with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.AlertSource == nil {
		log.Printf("[ERROR] Reading iLert alert source error: empty response ")
		return diag.Errorf("alert source response is empty")
	}

	d.Set("name", result.AlertSource.Name)
	d.Set("integration_type", result.AlertSource.IntegrationType)
	d.Set("escalation_policy", strconv.FormatInt(result.AlertSource.EscalationPolicy.ID, 10))
	d.Set("incident_creation", result.AlertSource.IncidentCreation)
	d.Set("active", result.AlertSource.Active)
	d.Set("incident_priority_rule", result.AlertSource.IncidentPriorityRule)
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

	teams, err := flattenTeamsList(result.AlertSource.Teams)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("teams", teams); err != nil {
		return diag.FromErr(fmt.Errorf("error setting teams: %s", err))
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

	supportHours, err := flattenSupportHours(result.AlertSource.SupportHours)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("support_hours", supportHours); err != nil {
		return diag.FromErr(fmt.Errorf("error setting support hours: %s", err))
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
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an alert source with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating iLert alert source error %s", err.Error())
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
				return resource.RetryableError(fmt.Errorf("waiting for alert source with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an alert source with ID %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting iLert alert source error %s", err.Error())
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
	_, err = client.GetAlertSource(&ilert.GetAlertSourceInput{AlertSourceID: ilert.Int64(alertSourceID)})
	if err != nil {
		if _, ok := err.(*ilert.NotFoundAPIError); ok {
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
	result["auto_raise_incidents"] = supportHours.AutoRaiseIncidents

	supportDays := make([]interface{}, 0)
	supportDaysItem := make(map[string]interface{})
	if supportHours.SupportDays.MONDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.MONDAY.Start
		supportDay["end"] = supportHours.SupportDays.MONDAY.End
		supportDaysItem["monday"] = []interface{}{supportDay}
	}
	if supportHours.SupportDays.TUESDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.TUESDAY.Start
		supportDay["end"] = supportHours.SupportDays.TUESDAY.End
		supportDaysItem["tuesday"] = []interface{}{supportDay}
	}
	if supportHours.SupportDays.WEDNESDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.WEDNESDAY.Start
		supportDay["end"] = supportHours.SupportDays.WEDNESDAY.End
		supportDaysItem["wednesday"] = []interface{}{supportDay}
	}
	if supportHours.SupportDays.THURSDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.THURSDAY.Start
		supportDay["end"] = supportHours.SupportDays.THURSDAY.End
		supportDaysItem["thursday"] = []interface{}{supportDay}
	}
	if supportHours.SupportDays.FRIDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.FRIDAY.Start
		supportDay["end"] = supportHours.SupportDays.FRIDAY.End
		supportDaysItem["friday"] = []interface{}{supportDay}
	}
	if supportHours.SupportDays.SATURDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.SATURDAY.Start
		supportDay["end"] = supportHours.SupportDays.SATURDAY.End
		supportDaysItem["saturday"] = []interface{}{supportDay}
	}
	if supportHours.SupportDays.SUNDAY != nil {
		supportDay := make(map[string]interface{})
		supportDay["start"] = supportHours.SupportDays.SUNDAY.Start
		supportDay["end"] = supportHours.SupportDays.SUNDAY.End
		supportDaysItem["sunday"] = []interface{}{supportDay}
	}
	supportDays = append(supportDays, supportDaysItem)
	result["support_days"] = supportDays

	results = append(results, result)

	return results, nil
}

func flattenTeamsList(list []ilert.TeamShort) ([]int64, error) {
	if list == nil {
		return make([]int64, 0), nil
	}
	results := make([]int64, 0)
	for _, item := range list {
		results = append(results, item.ID)
	}

	return results, nil
}
