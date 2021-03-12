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
					},
				},
			},
			"teams": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
		Create: resourceEscalationPolicyCreate,
		Read:   resourceEscalationPolicyRead,
		Update: resourceEscalationPolicyUpdate,
		Delete: resourceEscalationPolicyDelete,
		Exists: resourceEscalationPolicyExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

	return escalationPolicy, nil
}

func resourceEscalationPolicyCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	escalationPolicy, err := buildEscalationPolicy(d)
	if err != nil {
		log.Printf("[ERROR] Building escalation policy error %s", err.Error())
		return err
	}

	log.Printf("[INFO] Creating escalation policy %s", escalationPolicy.Name)

	result, err := client.CreateEscalationPolicy(&ilert.CreateEscalationPolicyInput{EscalationPolicy: escalationPolicy})
	if err != nil {
		log.Printf("[ERROR] Creating iLert escalation policy error %s", err.Error())
		return err
	}

	d.SetId(strconv.FormatInt(result.EscalationPolicy.ID, 10))

	return resourceEscalationPolicyRead(d, m)
}

func resourceEscalationPolicyRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	escalationPolicyID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse escalation policy id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading escalation policy: %s", d.Id())
	result, err := client.GetEscalationPolicy(&ilert.GetEscalationPolicyInput{EscalationPolicyID: ilert.Int64(escalationPolicyID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			log.Printf("[WARN] Removing escalation policy %s from state because it no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Could not read an escalation policy with ID %s", d.Id())
	}

	d.Set("name", result.EscalationPolicy.Name)
	d.Set("frequency", result.EscalationPolicy.Frequency)
	d.Set("repeating", result.EscalationPolicy.Repeating)

	escalationRules, err := flattenEscalationRulesList(result.EscalationPolicy.EscalationRules)
	if err != nil {
		return err
	}
	if err := d.Set("escalation_rule", escalationRules); err != nil {
		return fmt.Errorf("error setting escalation rules: %s", err)
	}

	teams, err := flattenTeamsList(result.EscalationPolicy.Teams)
	if err != nil {
		return err
	}
	if err := d.Set("teams", teams); err != nil {
		return fmt.Errorf("error setting teams: %s", err)
	}

	return nil
}

func resourceEscalationPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	escalationPolicy, err := buildEscalationPolicy(d)
	if err != nil {
		log.Printf("[ERROR] Building escalation policy error %s", err.Error())
		return err
	}

	escalationPolicyID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse escalation policy id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Updating escalation policy: %s", d.Id())
	_, err = client.UpdateEscalationPolicy(&ilert.UpdateEscalationPolicyInput{EscalationPolicy: escalationPolicy, EscalationPolicyID: ilert.Int64(escalationPolicyID)})
	if err != nil {
		log.Printf("[ERROR] Updating iLert escalation policy error %s", err.Error())
		return err
	}
	return resourceEscalationPolicyRead(d, m)
}

func resourceEscalationPolicyDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ilert.Client)

	escalationPolicyID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse escalation policy id %s", err.Error())
		return unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Deleting escalation policy: %s", d.Id())
	_, err = client.DeleteEscalationPolicy(&ilert.DeleteEscalationPolicyInput{EscalationPolicyID: ilert.Int64(escalationPolicyID)})
	if err != nil {
		return err
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
	_, err = client.GetEscalationPolicy(&ilert.GetEscalationPolicyInput{EscalationPolicyID: ilert.Int64(escalationPolicyID)})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not find") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func flattenEscalationRulesList(list []ilert.EscalationRule) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, item := range list {
		result := make(map[string]interface{})
		result["escalation_timeout"] = item.EscalationTimeout
		if item.User != nil {
			result["user"] = strconv.FormatInt(item.User.ID, 10)
		}
		if item.Schedule != nil {
			result["schedule"] = strconv.FormatInt(item.Schedule.ID, 10)
		}
		results = append(results, result)
	}

	return results, nil
}
