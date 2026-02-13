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

func resourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"escalation_rule": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"escalation_timeout": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 525600),
						},
						"user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"schedule": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"users": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"first_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"last_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"schedules": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
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
			"repeating": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"frequency": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 9),
			},
			"delay_min": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 15),
			},
			"routing_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		CreateContext: resourceEscalationPolicyCreate,
		ReadContext:   resourceEscalationPolicyRead,
		UpdateContext: resourceEscalationPolicyUpdate,
		DeleteContext: resourceEscalationPolicyDelete,
		Exists:        resourceEscalationPolicyExists,
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

func buildEscalationPolicy(d *schema.ResourceData) (*ilert.EscalationPolicy, error) {
	name := d.Get("name").(string)

	escalationPolicy := &ilert.EscalationPolicy{
		Name: name,
	}

	if val, ok := d.GetOk("escalation_rule"); ok {
		vL := val.([]any)
		nps := make([]ilert.EscalationRule, 0)
		for _, m := range vL {
			v := m.(map[string]any)
			ep := ilert.EscalationRule{
				EscalationTimeout: v["escalation_timeout"].(int),
			}
			err := checkEscalationRuleSchema(v)
			if err != nil {
				log.Printf("[ERROR] Could not validate escalation rule: %s", err.Error())
				return nil, err
			}
			if v["user"] != nil && v["user"].(string) != "" {
				userID, err := strconv.ParseInt(v["user"].(string), 10, 64)
				if err != nil {
					log.Printf("[ERROR] Could not parse user id %s", err.Error())
					return nil, unconvertibleIDErr(v["user"].(string), err)
				}
				ep.User = &ilert.User{
					ID: userID,
				}
			} else if v["schedule"] != nil && v["schedule"].(string) != "" {
				scheduleID, err := strconv.ParseInt(v["schedule"].(string), 10, 64)
				if err != nil {
					log.Printf("[ERROR] Could not parse schedule id %s", err.Error())
					return nil, unconvertibleIDErr(v["schedule"].(string), err)
				}
				ep.Schedule = &ilert.Schedule{
					ID: scheduleID,
				}
			} else {
				if v["users"] != nil && len(v["users"].([]any)) > 0 {
					usr := make([]ilert.User, 0)
					uL := v["users"].([]any)
					for _, u := range uL {
						v := u.(map[string]any)
						uid, err := strconv.ParseInt(v["id"].(string), 10, 64)
						if err != nil {
							log.Printf("[ERROR] Could not parse user id %s", err.Error())
							return nil, unconvertibleIDErr(v["id"].(string), err)
						}
						us := ilert.User{
							ID: uid,
						}
						if v["first_name"] != nil && v["first_name"].(string) != "" {
							us.FirstName = v["first_name"].(string)
						}
						if v["last_name"] != nil && v["last_name"].(string) != "" {
							us.LastName = v["last_name"].(string)
						}
						usr = append(usr, us)
					}
					ep.Users = usr
				}
				if v["schedules"] != nil && len(v["schedules"].([]any)) > 0 {
					sdl := make([]ilert.Schedule, 0)
					sL := v["schedules"].([]any)
					for _, u := range sL {
						v := u.(map[string]any)
						sid, err := strconv.ParseInt(v["id"].(string), 10, 64)
						if err != nil {
							log.Printf("[ERROR] Could not parse user id %s", err.Error())
							return nil, unconvertibleIDErr(v["id"].(string), err)
						}
						sd := ilert.Schedule{
							ID: sid,
						}
						if v["name"] != nil && v["name"].(string) != "" {
							sd.Name = v["name"].(string)
						}
						sdl = append(sdl, sd)
					}
					ep.Schedules = sdl
				}
			}
			nps = append(nps, ep)
		}
		escalationPolicy.EscalationRules = nps
	}

	if val, ok := d.GetOk("teams"); ok {
		vL := val.([]any)
		tms := make([]ilert.TeamShort, 0)

		for _, m := range vL {
			v := int64(m.(int))
			tms = append(tms, ilert.TeamShort{ID: v})
		}
		escalationPolicy.Teams = tms
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
		escalationPolicy.Teams = tms
	}

	if val, ok := d.GetOk("repeating"); ok {
		escalationPolicy.Repeating = val.(bool)
	}

	if val, ok := d.GetOk("frequency"); ok {
		escalationPolicy.Frequency = val.(int)
	}

	if val, ok := d.GetOk("delay_min"); ok {
		escalationPolicy.DelayMin = val.(int)
	}

	if val, ok := d.GetOk("routing_key"); ok {
		escalationPolicy.RoutingKey = val.(string)
	}

	return escalationPolicy, nil
}

func resourceEscalationPolicyCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	escalationPolicy, err := buildEscalationPolicy(d)
	if err != nil {
		log.Printf("[ERROR] Building escalation policy error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating escalation policy %s", escalationPolicy.Name)

	result := &ilert.CreateEscalationPolicyOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateEscalationPolicy(&ilert.CreateEscalationPolicyInput{EscalationPolicy: escalationPolicy})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert escalation policy error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert escalation policy error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.EscalationPolicy == nil {
		log.Printf("[ERROR] Creating ilert escalation policy error: empty response")
		return diag.Errorf("escalation policy response is empty")
	}

	d.SetId(strconv.FormatInt(result.EscalationPolicy.ID, 10))

	return resourceEscalationPolicyRead(ctx, d, m)
}

func resourceEscalationPolicyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	escalationPolicyID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse escalation policy id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading escalation policy: %s", d.Id())

	result := &ilert.GetEscalationPolicyOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetEscalationPolicy(&ilert.GetEscalationPolicyInput{EscalationPolicyID: ilert.Int64(escalationPolicyID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing escalation policy %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an escalation policy with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert escalation policy error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.EscalationPolicy == nil {
		log.Printf("[ERROR] Reading ilert escalation policy error: empty response")
		return diag.Errorf("escalation policy response is empty")
	}

	err = transformEscalationPolicyResource(result.EscalationPolicy, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceEscalationPolicyUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	escalationPolicy, err := buildEscalationPolicy(d)
	if err != nil {
		log.Printf("[ERROR] Building escalation policy error %s", err.Error())
		return diag.FromErr(err)
	}

	escalationPolicyID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse escalation policy id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating escalation policy: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateEscalationPolicy(&ilert.UpdateEscalationPolicyInput{EscalationPolicy: escalationPolicy, EscalationPolicyID: ilert.Int64(escalationPolicyID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an escalation policy with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert escalation policy error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceEscalationPolicyRead(ctx, d, m)
}

func resourceEscalationPolicyDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	escalationPolicyID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse escalation policy id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting escalation policy: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteEscalationPolicy(&ilert.DeleteEscalationPolicyInput{EscalationPolicyID: ilert.Int64(escalationPolicyID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an escalation policy with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert escalation policy error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceEscalationPolicyExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	escalationPolicyID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse escalation policy id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading escalation policy: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetEscalationPolicy(&ilert.GetEscalationPolicyInput{EscalationPolicyID: ilert.Int64(escalationPolicyID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert escalation policy error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an escalation policy with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert escalation policy error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func transformEscalationPolicyResource(escalationPolicy *ilert.EscalationPolicy, d *schema.ResourceData) error {
	d.Set("name", escalationPolicy.Name)

	escalationRules, err := flattenEscalationRulesList(escalationPolicy.EscalationRules, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening escalation rules: %s", err.Error())
	}
	if err := d.Set("escalation_rule", escalationRules); err != nil {
		return fmt.Errorf("[ERROR] Error setting escalation rules: %s", err.Error())
	}

	if val, ok := d.GetOk("team"); ok {
		if val != nil {
			vL := val.([]any)
			teams := make([]any, 0)
			for i, item := range escalationPolicy.Teams {
				team := make(map[string]any)
				v := vL[i].(map[string]any)
				team["id"] = item.ID

				// Means: if server response has a name set, and the user typed in a name too,
				// only then team name is stored in the terraform state
				if item.Name != "" && v["name"] != nil && v["name"].(string) != "" {
					team["name"] = item.Name
				}
				teams = append(teams, team)
			}

			if err := d.Set("team", teams); err != nil {
				return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
			}
		}
	} else if val, ok := d.GetOk("teams"); ok {
		if val != nil {
			teams := make([]any, 0)
			for _, item := range escalationPolicy.Teams {
				team := make(map[string]any)
				team["id"] = item.ID
				teams = append(teams, team)
			}
			if err := d.Set("team", teams); err != nil {
				return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
			}

			d.Set("teams", nil)
		}
	} else if d.Id() == "" && len(escalationPolicy.Teams) > 0 {
		teams, err := flattenTeamShortList(escalationPolicy.Teams, d)
		if err != nil {
			return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
		}
		if err := d.Set("team", teams); err != nil {
			return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
		}
	}

	d.Set("repeating", escalationPolicy.Repeating)
	d.Set("frequency", escalationPolicy.Frequency)
	d.Set("delay_min", escalationPolicy.DelayMin)
	d.Set("routing_key", escalationPolicy.RoutingKey)

	return nil
}

func flattenEscalationRulesList(list []ilert.EscalationRule, d *schema.ResourceData) ([]any, error) {
	if list == nil {
		return make([]any, 0), nil
	}

	results := make([]any, 0)
	if val, ok := d.GetOk("escalation_rule"); ok && len(val.([]any)) > 0 {
		vL := val.([]any)
		for i, item := range list {
			if vL != nil && i < len(vL) && vL[i] != nil {
				result := make(map[string]any)
				result["escalation_timeout"] = item.EscalationTimeout
				v := vL[i].(map[string]any)
				if item.User != nil && v["user"] != nil && v["user"].(string) != "" {
					result["user"] = strconv.FormatInt(item.User.ID, 10)
				}
				if item.Schedule != nil && v["schedule"] != nil && v["schedule"].(string) != "" {
					result["schedule"] = strconv.FormatInt(item.Schedule.ID, 10)
				}

				user := v["users"].([]any)
				users, err := flattenResponderUserShortList(item.Users, user)
				if err != nil {
					return nil, err
				}
				result["users"] = users

				schedule := v["schedules"].([]any)
				schedules, err := flattenResponderScheduleList(item.Schedules, schedule)
				if err != nil {
					return nil, err
				}
				result["schedules"] = schedules

				results = append(results, result)
			}
		}
	} else if d.Id() == "" {
		for _, item := range list {
			result := make(map[string]any)
			result["escalation_timeout"] = item.EscalationTimeout

			users, err := flattenUserShortList(item.Users)
			if err != nil {
				return nil, err
			}
			result["users"] = users

			schedules, err := flattenScheduleList(item.Schedules)
			if err != nil {
				return nil, err
			}
			result["schedules"] = schedules

			results = append(results, result)
		}
	}

	return results, nil
}

func flattenScheduleList(list []ilert.Schedule) ([]any, error) {
	if list == nil {
		return make([]any, 0), nil
	}

	results := make([]any, 0)
	for _, item := range list {
		result := make(map[string]any)
		result["id"] = strconv.FormatInt(item.ID, 10)
		if item.Name != "" {
			result["name"] = item.Name
		}
		results = append(results, result)
	}

	return results, nil
}

func flattenResponderScheduleList(list []ilert.Schedule, schedule []any) ([]any, error) {
	if list == nil || schedule == nil || len(schedule) <= 0 {
		return make([]any, 0), nil
	}

	results := make([]any, 0)
	for i, item := range list {
		result := make(map[string]any)
		result["id"] = strconv.FormatInt(item.ID, 10)
		var sdn any
		if len(schedule) > 0 && i < len(schedule) && schedule[i] != nil && len(schedule[i].(map[string]any)) > 0 {
			sdn = schedule[i].(map[string]any)["name"]
		}

		if item.Name != "" && sdn != nil && sdn.(string) != "" {
			result["name"] = item.Name
		}
		results = append(results, result)
	}

	return results, nil
}

func checkEscalationRuleSchema(rule map[string]any) error {
	if rule["user"] != nil && rule["user"].(string) != "" {
		if (rule["schedule"] != nil && rule["schedule"].(string) != "") || (rule["users"] != nil && len(rule["users"].([]any)) > 0) || (rule["schedules"] != nil && len(rule["schedules"].([]any)) > 0) {
			err := errors.New("fields 'schedule', 'users', or 'schedules' are not allowed when setting 'user'")
			return err
		}

	}
	if rule["schedule"] != nil && rule["schedule"].(string) != "" {
		if (rule["user"] != nil && rule["user"].(string) != "") || (rule["users"] != nil && len(rule["users"].([]any)) > 0) || (rule["schedules"] != nil && len(rule["schedules"].([]any)) > 0) {
			err := errors.New("fields 'user', 'users', or 'schedules' are not allowed when setting 'schedule'")
			return err
		}

	}
	if rule["users"] != nil && len(rule["users"].([]any)) > 0 {
		if (rule["user"] != nil && rule["user"].(string) != "") || (rule["schedule"] != nil && rule["schedule"].(string) != "") {
			err := errors.New("fields 'user' or 'schedule' are not allowed when setting 'users'")
			return err
		}

	}
	if rule["schedules"] != nil && len(rule["schedules"].([]any)) > 0 {
		if (rule["user"] != nil && rule["user"].(string) != "") || (rule["schedule"] != nil && rule["schedule"].(string) != "") {
			err := errors.New("fields 'user' or 'schedule' are not allowed when setting 'schedules'")
			return err
		}

	}
	return nil
}
