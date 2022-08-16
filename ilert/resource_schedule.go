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
				Type:     schema.TypeString,
				Required: true,
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
									"name": {
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
													Type:     schema.TypeString,
													Required: true,
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
													Type:     schema.TypeString,
													Required: true,
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

	schedule := &ilert.Schedule{
		Name: name,
	}

	if val, ok := d.GetOk("status"); ok {
		schedule.Status = val.(string)
	}

	if val, ok := d.GetOk("description"); ok {
		schedule.Description = val.(string)
	}

	if val, ok := d.GetOk("one_open_incident_only"); ok {
		schedule.OneOpenIncidentOnly = val.(bool)
	}

	if val, ok := d.GetOk("show_uptime_history"); ok {
		schedule.ShowUptimeHistory = val.(bool)
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
		r, err := client.GetSchedule(&ilert.GetScheduleInput{ScheduleID: ilert.Int64(scheduleID)})
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
	d.Set("status", result.Schedule.Status)
	d.Set("description", result.Schedule.Description)
	d.Set("one_open_incident_only", result.Schedule.OneOpenIncidentOnly)
	d.Set("show_uptime_history", result.Schedule.ShowUptimeHistory)

	teams, err := flattenTeamShortList(result.Schedule.Teams)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
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
