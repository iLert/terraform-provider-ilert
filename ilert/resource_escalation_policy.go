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
	"github.com/iLert/ilert-go/v2"
)

func resourceEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"frequency": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 9),
			},
			"repeating": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"escalation_rule": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"escalation_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
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
	frequency := d.Get("frequency").(int)
	repeating := d.Get("repeating").(bool)

	escalationPolicy := &ilert.EscalationPolicy{
		Name:      name,
		Frequency: frequency,
		Repeating: repeating,
	}

	if val, ok := d.GetOk("escalation_rule"); ok {
		vL := val.([]interface{})
		nps := make([]ilert.EscalationRule, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
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
				if v["users"] != nil && len(v["users"].([]interface{})) > 0 {
					usr := make([]ilert.User, 0)
					uL := v["users"].([]interface{})
					for _, u := range uL {
						v := u.(map[string]interface{})
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
				if v["schedules"] != nil && len(v["schedules"].([]interface{})) > 0 {
					sdl := make([]ilert.Schedule, 0)
					sL := v["schedules"].([]interface{})
					for _, u := range sL {
						v := u.(map[string]interface{})
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
		vL := val.([]interface{})
		tms := make([]ilert.TeamShort, 0)

		for _, m := range vL {
			v := int64(m.(int))
			tms = append(tms, ilert.TeamShort{ID: v})
		}
		escalationPolicy.Teams = tms
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
		escalationPolicy.Teams = tms
	}

	return escalationPolicy, nil
}

func resourceEscalationPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		log.Printf("[ERROR] Creating ilert escalation policy error: empty response ")
		return diag.Errorf("escalation policy response is empty")
	}

	d.SetId(strconv.FormatInt(result.EscalationPolicy.ID, 10))

	return resourceEscalationPolicyRead(ctx, d, m)
}

func resourceEscalationPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an escalation policy with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.EscalationPolicy == nil {
		log.Printf("[ERROR] Reading ilert escalation policy error: empty response ")
		return diag.Errorf("escalation policy response is empty")
	}

	d.Set("name", result.EscalationPolicy.Name)
	d.Set("frequency", result.EscalationPolicy.Frequency)
	d.Set("repeating", result.EscalationPolicy.Repeating)

	escalationRules, err := flattenEscalationRulesList(result.EscalationPolicy.EscalationRules, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("escalation_rule", escalationRules); err != nil {
		return diag.Errorf("error setting escalation rules: %s", err)
	}

	if val, ok := d.GetOk("team"); ok {
		if val != nil {
			vL := val.([]interface{})
			teams := make([]interface{}, 0)
			for i, item := range result.EscalationPolicy.Teams {
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
			for _, item := range result.EscalationPolicy.Teams {
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

	return nil
}

func resourceEscalationPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an escalation policy with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert escalation policy error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceEscalationPolicyRead(ctx, d, m)
}

func resourceEscalationPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
				return resource.RetryableError(fmt.Errorf("waiting for escalation policy with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an escalation policy with ID %s", d.Id()))
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

func resourceEscalationPolicyExists(d *schema.ResourceData, m interface{}) (bool, error) {
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

func flattenEscalationRulesList(list []ilert.EscalationRule, d *schema.ResourceData) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}

	usL := d.Get("escalation_rule").([]interface{})

	results := make([]interface{}, 0)
	for i, item := range list {
		if usL != nil && i < len(usL) && usL[i] != nil {
			result := make(map[string]interface{})
			result["escalation_timeout"] = item.EscalationTimeout
			v := usL[i].(map[string]interface{})
			if item.User != nil && v["user"] != nil && v["user"].(string) != "" {
				result["user"] = strconv.FormatInt(item.User.ID, 10)
			}
			if item.Schedule != nil && v["schedule"] != nil && v["schedule"].(string) != "" {
				result["schedule"] = strconv.FormatInt(item.Schedule.ID, 10)
			}

			user := v["users"].([]interface{})
			users, err := flattenUserShortList(item.Users, user)
			if err != nil {
				return nil, err
			}
			result["users"] = users

			schedule := v["schedules"].([]interface{})
			schedules, err := flattenScheduleShortList(item.Schedules, schedule)
			if err != nil {
				return nil, err
			}
			result["schedules"] = schedules

			results = append(results, result)
		}
	}

	return results, nil
}

func flattenScheduleShortList(list []ilert.Schedule, schedule []interface{}) ([]interface{}, error) {
	if list == nil || schedule == nil || len(schedule) <= 0 {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	for i, item := range list {
		if schedule != nil {
			result := make(map[string]interface{})
			result["id"] = strconv.FormatInt(item.ID, 10)
			var sdn interface{}
			if len(schedule) > 0 && i < len(schedule) && schedule[i] != nil && len(schedule[i].(map[string]interface{})) > 0 {
				sdn = schedule[i].(map[string]interface{})["name"]
			}

			if item.Name != "" && sdn != nil && sdn.(string) != "" {
				result["name"] = item.Name
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func checkEscalationRuleSchema(rule map[string]interface{}) error {
	if rule["user"] != nil && rule["user"].(string) != "" {
		if (rule["schedule"] != nil && rule["schedule"].(string) != "") || (rule["users"] != nil && len(rule["users"].([]interface{})) > 0) || (rule["schedules"] != nil && len(rule["schedules"].([]interface{})) > 0) {
			err := errors.New("fields 'schedule', 'users', or 'schedules' are not allowed when setting 'user'")
			return err
		}

	}
	if rule["schedule"] != nil && rule["schedule"].(string) != "" {
		if (rule["user"] != nil && rule["user"].(string) != "") || (rule["users"] != nil && len(rule["users"].([]interface{})) > 0) || (rule["schedules"] != nil && len(rule["schedules"].([]interface{})) > 0) {
			err := errors.New("fields 'user', 'users', or 'schedules' are not allowed when setting 'schedule'")
			return err
		}

	}
	if rule["users"] != nil && len(rule["users"].([]interface{})) > 0 {
		if (rule["user"] != nil && rule["user"].(string) != "") || (rule["schedule"] != nil && rule["schedule"].(string) != "") {
			err := errors.New("fields 'user' or 'schedule' are not allowed when setting 'users'")
			return err
		}

	}
	if rule["schedules"] != nil && len(rule["schedules"].([]interface{})) > 0 {
		if (rule["user"] != nil && rule["user"].(string) != "") || (rule["schedule"] != nil && rule["schedule"].(string) != "") {
			err := errors.New("fields 'user' or 'schedule' are not allowed when setting 'schedules'")
			return err
		}

	}
	return nil
}
