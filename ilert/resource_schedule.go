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
	"github.com/iLert/ilert-go/v2"
)

func resourceSchedule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"timezone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ilert.ScheduleTypeAll, false),
			},
			"schedule_layer": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"starts_on": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ends_on": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"user": {
							Type:     schema.TypeList,
							Required: true,
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
						"rotation": {
							Type:     schema.TypeString,
							Required: true,
						},
						"restriction_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"restriction": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"from": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"day_of_week": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(ilert.DayOfWeekAll, false),
												},
												"time": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"to": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"day_of_week": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(ilert.DayOfWeekAll, false),
												},
												"time": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"shift": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Required: true,
						},
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"show_gaps": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"default_shift_duration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"current_shift": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Required: true,
						},
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"next_shift": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Required: true,
						},
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
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
		},
		CreateContext: resourceScheduleCreate,
		ReadContext:   resourceScheduleRead,
		UpdateContext: resourceScheduleUpdate,
		DeleteContext: resourceScheduleDelete,
		Exists:        resourceScheduleExists,
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

func buildSchedule(d *schema.ResourceData) (*ilert.Schedule, error) {
	name := d.Get("name").(string)
	timezone := d.Get("timezone").(string)
	scheduleType := d.Get("type").(string)

	schedule := &ilert.Schedule{
		Name:     name,
		Timezone: timezone,
		Type:     scheduleType,
	}

	if val, ok := d.GetOk("schedule_layer"); ok && scheduleType == ilert.ScheduleType.Recurring {
		vL := val.([]interface{})
		sdl := make([]ilert.ScheduleLayer, 0)
		for _, m := range vL {
			v := m.(map[string]interface{})
			sd := ilert.ScheduleLayer{
				Name:     v["name"].(string),
				StartsOn: v["starts_on"].(string),
				Rotation: v["rotation"].(string),
			}

			usr := make([]ilert.User, 0)
			uL := v["user"].([]interface{})
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
			sd.Users = usr

			if v["restriction_type"] != nil && v["restriction_type"].(string) != "" {
				sd.RestrictionType = v["restriction_type"].(string)
			}

			if v["restriction"] != nil {
				rns := make([]ilert.LayerRestriction, 0)
				rL := v["restriction"].([]interface{})
				for _, r := range rL {
					v := r.(map[string]interface{})

					fL := v["from"].([]interface{})
					f := fL[0].(map[string]interface{})
					from := ilert.TimeOfWeek{
						DayOfWeek: f["day_of_week"].(string),
						Time:      f["time"].(string),
					}

					tL := v["to"].([]interface{})
					t := tL[0].(map[string]interface{})
					to := ilert.TimeOfWeek{
						DayOfWeek: t["day_of_week"].(string),
						Time:      t["time"].(string),
					}

					rn := ilert.LayerRestriction{
						From: &from,
						To:   &to,
					}

					rns = append(rns, rn)
				}
				sd.Restrictions = rns
			}
			sdl = append(sdl, sd)
		}
		schedule.ScheduleLayers = sdl
	}

	if val, ok := d.GetOk("shift"); ok && scheduleType == ilert.ScheduleType.Static {
		vL := val.([]interface{})
		shs := make([]ilert.Shift, 0)
		for _, s := range vL {
			v := s.(map[string]interface{})
			userID, err := strconv.ParseInt(v["user"].(string), 10, 64)
			if err != nil {
				log.Printf("[ERROR] Could not parse user id %s", err.Error())
				return nil, unconvertibleIDErr(v["user"].(string), err)
			}
			us := ilert.User{
				ID: userID,
			}
			sh := ilert.Shift{
				User:  us,
				Start: v["start"].(string),
				End:   v["end"].(string),
			}
			shs = append(shs, sh)
		}
		schedule.Shifts = shs
	}

	if val, ok := d.GetOk("show_gaps"); ok {
		schedule.ShowGaps = val.(bool)
	}

	if val, ok := d.GetOk("default_shift_duration"); ok {
		schedule.DefaultShiftDuration = val.(string)
	}

	if val, ok := d.GetOk("current_shift"); ok {
		if vL, ok := val.([]interface{}); ok && len(vL) > 0 && vL[0] != nil {
			tmp := &ilert.Shift{}
			if v, ok := vL[0].(map[string]interface{}); ok && len(v) > 0 {
				usr := ilert.User{
					ID: int64(v["user"].(int)),
				}
				tmp.User = usr
				tmp.Start = v["start"].(string)
				tmp.End = v["end"].(string)
				schedule.CurrentShift = tmp
			} else {
				schedule.CurrentShift = nil
			}
		}
	}

	if val, ok := d.GetOk("next_shift"); ok {
		if vL, ok := val.([]interface{}); ok && len(vL) > 0 && vL[0] != nil {
			tmp := &ilert.Shift{}
			if v, ok := vL[0].(map[string]interface{}); ok && len(v) > 0 {
				usr := ilert.User{
					ID: int64(v["user"].(int)),
				}
				tmp.User = usr
				tmp.Start = v["start"].(string)
				tmp.End = v["end"].(string)
				schedule.NextShift = tmp
			} else {
				schedule.NextShift = nil
			}
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
		schedule.Teams = tms
	}

	return schedule, nil
}

func resourceScheduleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	schedule, err := buildSchedule(d)
	if err != nil {
		log.Printf("[ERROR] Building schedule error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating schedule %s", schedule.Name)

	result := &ilert.CreateScheduleOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateSchedule(&ilert.CreateScheduleInput{Schedule: schedule})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating iLert schedule error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for schedule to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(err)
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating iLert schedule error %s", err.Error())
		return diag.FromErr(err)
	}
	if result == nil || result.Schedule == nil {
		log.Printf("[ERROR] Creating iLert schedule error: empty response ")
		return diag.Errorf("schedule response is empty")
	}

	d.SetId(strconv.FormatInt(result.Schedule.ID, 10))

	return resourceScheduleRead(ctx, d, m)
}

func resourceScheduleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	scheduleID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse schedule id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading schedule: %s", d.Id())
	result := &ilert.GetScheduleOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		includes := make([]*string, 0)
		if d.Get("type").(string) == "RECURRING" {
			includes = append(includes, ilert.String("scheduleLayers"), ilert.String("nextShift"), ilert.String("currentShift"))
		} else if d.Get("type").(string) == "STATIC" {
			includes = append(includes, ilert.String("shifts"))
		}
		r, err := client.GetSchedule(&ilert.GetScheduleInput{ScheduleID: ilert.Int64(scheduleID), Include: includes})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing schedule %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for schedule with id '%s' to be read", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an schedule with ID %s", d.Id()))
		}
		result = r
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if result == nil || result.Schedule == nil {
		log.Printf("[ERROR] Reading iLert schedule error: empty response ")
		return diag.Errorf("schedule response is empty")
	}

	d.Set("name", result.Schedule.Name)
	d.Set("timezone", result.Schedule.Timezone)
	d.Set("type", result.Schedule.Type)

	layers, err := flattenScheduleLayerList(result.Schedule.ScheduleLayers, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schedule_layer", layers); err != nil {
		return diag.Errorf("error setting schedule layers: %s", err)
	}

	shifts, err := flattenShiftList(result.Schedule.Shifts)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("shift", shifts); err != nil {
		return diag.Errorf("error setting shifts: %s", err)
	}

	d.Set("show_gaps", result.Schedule.ShowGaps)
	d.Set("default_shift_duration", result.Schedule.DefaultShiftDuration)

	if result.Schedule.CurrentShift != nil {
		d.Set("current_shift", []interface{}{
			map[string]interface{}{
				"user":  result.Schedule.CurrentShift.User,
				"start": result.Schedule.CurrentShift.Start,
				"end":   result.Schedule.CurrentShift.End,
			},
		})
	} else {
		d.Set("current_shift", []interface{}{})
	}

	if result.Schedule.NextShift != nil {
		d.Set("next_shift", []interface{}{
			map[string]interface{}{
				"user":  result.Schedule.NextShift.User,
				"start": result.Schedule.NextShift.Start,
				"end":   result.Schedule.NextShift.End,
			},
		})
	} else {
		d.Set("next_shift", []interface{}{})
	}

	if val, ok := d.GetOk("team"); ok {
		if val != nil {
			vL := val.([]interface{})
			teams := make([]interface{}, 0)
			for i, item := range result.Schedule.Teams {
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

	return nil
}

func resourceScheduleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	schedule, err := buildSchedule(d)
	if err != nil {
		log.Printf("[ERROR] Building schedule error %s", err.Error())
		return diag.FromErr(err)
	}

	scheduleID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse schedule id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating schedule: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateSchedule(&ilert.UpdateScheduleInput{Schedule: schedule, ScheduleID: ilert.Int64(scheduleID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for schedule with id '%s' to be updated", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an schedule with ID %s", d.Id()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating iLert schedule error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceScheduleRead(ctx, d, m)
}

func resourceScheduleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*ilert.Client)

	scheduleID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse schedule id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting schedule: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteSchedule(&ilert.DeleteScheduleInput{ScheduleID: ilert.Int64(scheduleID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for schedule with id '%s' to be deleted", d.Id()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an schedule with ID %s", d.Id()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting iLert schedule error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceScheduleExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ilert.Client)

	scheduleID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse schedule id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading schedule: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetSchedule(&ilert.GetScheduleInput{ScheduleID: ilert.Int64(scheduleID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading iLert schedule error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for schedule to be read, error: %s", err.Error()))
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

func flattenScheduleLayerList(list []ilert.ScheduleLayer, d *schema.ResourceData) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}

	scL := d.Get("schedule_layer").([]interface{})

	results := make([]interface{}, 0)
	for i, item := range list {
		result := make(map[string]interface{})
		result["name"] = item.Name
		result["starts_on"] = item.StartsOn

		user := scL[i].(map[string]interface{})["user"].([]interface{})

		users, err := flattenUserShortList(item.Users, user)
		if err != nil {
			return nil, err
		}
		result["user"] = users

		result["rotation"] = item.Rotation
		if item.RestrictionType != "" {
			result["restriction_type"] = item.RestrictionType
		}

		restr, err := flattenRestrictionList(item.Restrictions)
		if err != nil {
			return nil, err
		}
		result["restriction"] = restr

		results = append(results, result)
	}

	return results, nil
}

func flattenUserShortList(list []ilert.User, user []interface{}) ([]interface{}, error) {
	if list == nil || user == nil || len(user) <= 0 {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	for i, item := range list {
		result := make(map[string]interface{})
		result["id"] = strconv.Itoa(int(item.ID))
		var ufn, uln interface{}
		if len(user) > 0 && user[i] != nil && len(user[i].(map[string]interface{})) > 0 {
			ufn = user[i].(map[string]interface{})["first_name"]
			uln = user[i].(map[string]interface{})["last_name"]
		}
		if item.FirstName != "" && ufn != nil && ufn.(string) != "" {
			result["first_name"] = item.FirstName
		}
		if item.LastName != "" && uln != nil && uln.(string) != "" {
			result["last_name"] = item.LastName
		}
		results = append(results, result)
	}

	return results, nil
}

func flattenRestrictionList(list []ilert.LayerRestriction) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	for _, item := range list {
		result := make(map[string]interface{})

		fromL := make([]interface{}, 0)
		fromL = append(fromL, make(map[string]interface{}))
		from := fromL[0].(map[string]interface{})
		from["day_of_week"] = item.From.DayOfWeek
		from["time"] = item.From.Time

		toL := make([]interface{}, 0)
		toL = append(toL, make(map[string]interface{}))
		to := toL[0].(map[string]interface{})
		to["day_of_week"] = item.To.DayOfWeek
		to["time"] = item.To.Time

		result["from"] = fromL
		result["to"] = toL

		results = append(results, result)
	}

	return results, nil
}

func flattenShiftList(list []ilert.Shift) ([]interface{}, error) {
	if list == nil {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)
	for _, item := range list {
		result := make(map[string]interface{})
		result["user"] = strconv.Itoa(int(item.User.ID))
		result["start"] = item.Start
		result["end"] = item.End

		results = append(results, result)
	}

	return results, nil
}

// func flattenShift(shift *ilert.Shift) (map[string]interface{}, error) {
// 	if shift == nil {
// 		return make(map[string]interface{}), nil
// 	}

// 	result := make(map[string]interface{})
// 	user := make(map[string]interface{})
// 	user["id"] = strconv.Itoa(int(shift.User.ID))
// 	result["user"] = user

// 	result["start"] = shift.Start
// 	result["end"] = shift.End

// 	return result, nil
// }
