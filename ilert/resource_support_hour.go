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

func resourceSupportHour() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
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
			"timezone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"support_days": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
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
		CreateContext: resourceSupportHourCreate,
		ReadContext:   resourceSupportHourRead,
		UpdateContext: resourceSupportHourUpdate,
		DeleteContext: resourceSupportHourDelete,
		Exists:        resourceSupportHourExists,
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

func buildSupportHour(d *schema.ResourceData) (*ilert.SupportHour, error) {
	name := d.Get("name").(string)

	supportHour := &ilert.SupportHour{
		Name: name,
	}

	if val, ok := d.GetOk("timezone"); ok {
		supportHour.Timezone = val.(string)
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
		supportHour.Teams = tms
	}

	if val, ok := d.GetOk("support_days"); ok {
		vL := val.([]any)
		days := ilert.SupportDays{}
		if len(vL) > 0 && vL[0] != nil {
			v := vL[0].(map[string]any)
			for d, sd := range v {
				s := sd.([]any)
				if len(s) > 0 && s[0] != nil {
					ds := s[0].(map[string]any)
					if d == "monday" {
						day := ilert.SupportDay{
							Start: ds["start"].(string),
							End:   ds["end"].(string),
						}
						days.MONDAY = &day
					}
					if d == "tuesday" {
						day := ilert.SupportDay{
							Start: ds["start"].(string),
							End:   ds["end"].(string),
						}
						days.TUESDAY = &day
					}
					if d == "wednesday" {
						day := ilert.SupportDay{
							Start: ds["start"].(string),
							End:   ds["end"].(string),
						}
						days.WEDNESDAY = &day
					}
					if d == "thursday" {
						day := ilert.SupportDay{
							Start: ds["start"].(string),
							End:   ds["end"].(string),
						}
						days.THURSDAY = &day
					}
					if d == "friday" {
						day := ilert.SupportDay{
							Start: ds["start"].(string),
							End:   ds["end"].(string),
						}
						days.FRIDAY = &day
					}
					if d == "saturday" {
						day := ilert.SupportDay{
							Start: ds["start"].(string),
							End:   ds["end"].(string),
						}
						days.SATURDAY = &day
					}
					if d == "sunday" {
						day := ilert.SupportDay{
							Start: ds["start"].(string),
							End:   ds["end"].(string),
						}
						days.SUNDAY = &day
					}
				}
			}
		}
		supportHour.SupportDays = &days
	}

	// if val, ok := d.GetOk("support_days"); ok {
	// 	vL := val.([]any)
	// 	if len(vL) > 0 && vL[0] != nil {
	// 		dL := vL[0].(map[string]any)
	// 		for day, times := range dL {
	// 			if

	// 		}

	// 	}
	// }

	return supportHour, nil
}

func resourceSupportHourCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	supportHour, err := buildSupportHour(d)
	if err != nil {
		log.Printf("[ERROR] Building support hour error %s", err.Error())
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating support hour %s", supportHour.Name)

	result := &ilert.CreateSupportHourOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		r, err := client.CreateSupportHour(&ilert.CreateSupportHourInput{SupportHour: supportHour})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Creating ilert support hour error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for support hour to be created, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not create a support hour with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Creating ilert support hour error %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.SupportHour == nil {
		log.Printf("[ERROR] Creating ilert support hour error: empty response")
		return diag.Errorf("support hour response is empty")
	}

	d.SetId(strconv.FormatInt(result.SupportHour.ID, 10))

	return resourceSupportHourRead(ctx, d, m)
}

func resourceSupportHourRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	supportHourID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse support hour id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Reading support hour: %s", d.Id())
	result := &ilert.GetSupportHourOutput{}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		r, err := client.GetSupportHour(&ilert.GetSupportHourInput{SupportHourID: ilert.Int64(supportHourID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				log.Printf("[WARN] Removing support hour %s from state because it no longer exist", d.Id())
				d.SetId("")
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for support hour with id '%s' to be read, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read an support hour with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = r
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert support hour error: %s", err.Error())
		return diag.FromErr(err)
	}

	if result == nil || result.SupportHour == nil {
		log.Printf("[ERROR] Reading ilert support hour error: empty response")
		return diag.Errorf("support hour response is empty")
	}

	d.Set("name", result.SupportHour.Name)
	d.Set("timezone", result.SupportHour.Timezone)

	teams, err := flattenTeamShortList(result.SupportHour.Teams, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team", teams); err != nil {
		return diag.Errorf("error setting teams: %s", err)
	}

	supportDays, err := flattenSupportDays(result.SupportHour.SupportDays)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("support_days", supportDays); err != nil {
		return diag.Errorf("error setting support days: %s", err)
	}

	return nil
}

func resourceSupportHourUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	supportHour, err := buildSupportHour(d)
	if err != nil {
		log.Printf("[ERROR] Building support hour error %s", err.Error())
		return diag.FromErr(err)
	}

	supportHourID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse support hour id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Updating support hour: %s", d.Id())

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err = client.UpdateSupportHour(&ilert.UpdateSupportHourInput{SupportHour: supportHour, SupportHourID: ilert.Int64(supportHourID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for support hour with id '%s' to be updated, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not update an support hour with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Updating ilert support hour error %s", err.Error())
		return diag.FromErr(err)
	}

	return resourceSupportHourRead(ctx, d, m)
}

func resourceSupportHourDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*ilert.Client)

	supportHourID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse support hour id %s", err.Error())
		return diag.FromErr(unconvertibleIDErr(d.Id(), err))
	}
	log.Printf("[DEBUG] Deleting support hour: %s", d.Id())
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = client.DeleteSupportHour(&ilert.DeleteSupportHourInput{SupportHourID: ilert.Int64(supportHourID)})
		if err != nil {
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for support hour with id '%s' to be deleted, error: %s", d.Id(), err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not delete an support hour with ID %s, error: %s", d.Id(), err.Error()))
		}
		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Deleting ilert support hour error %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceSupportHourExists(d *schema.ResourceData, m any) (bool, error) {
	client := m.(*ilert.Client)

	supportHourID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		log.Printf("[ERROR] Could not parse support hour id %s", err.Error())
		return false, unconvertibleIDErr(d.Id(), err)
	}
	log.Printf("[DEBUG] Reading support hour: %s", d.Id())
	ctx := context.Background()
	result := false
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := client.GetSupportHour(&ilert.GetSupportHourInput{SupportHourID: ilert.Int64(supportHourID)})
		if err != nil {
			if _, ok := err.(*ilert.NotFoundAPIError); ok {
				result = false
				return nil
			}
			if _, ok := err.(*ilert.RetryableAPIError); ok {
				log.Printf("[ERROR] Reading ilert support hour error '%s', so retry again", err.Error())
				time.Sleep(2 * time.Second)
				return resource.RetryableError(fmt.Errorf("waiting for support hour to be read, error: %s", err.Error()))
			}
			return resource.NonRetryableError(fmt.Errorf("could not read a support hour with ID %s, error: %s", d.Id(), err.Error()))
		}
		result = true
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] Reading ilert support hour error: %s", err.Error())
		return false, err
	}
	return result, nil
}

func flattenSupportDays(supportDays *ilert.SupportDays) ([]any, error) {
	if supportDays == nil {
		return make([]any, 0), nil
	}

	results := make([]any, 0)
	result := make(map[string]any)

	if supportDays.MONDAY != nil {
		supportDay := make(map[string]any)
		supportDay["start"] = supportDays.MONDAY.Start
		supportDay["end"] = supportDays.MONDAY.End
		result["monday"] = []any{supportDay}
	}
	if supportDays.TUESDAY != nil {
		supportDay := make(map[string]any)
		supportDay["start"] = supportDays.TUESDAY.Start
		supportDay["end"] = supportDays.TUESDAY.End
		result["tuesday"] = []any{supportDay}
	}
	if supportDays.WEDNESDAY != nil {
		supportDay := make(map[string]any)
		supportDay["start"] = supportDays.WEDNESDAY.Start
		supportDay["end"] = supportDays.WEDNESDAY.End
		result["wednesday"] = []any{supportDay}
	}
	if supportDays.THURSDAY != nil {
		supportDay := make(map[string]any)
		supportDay["start"] = supportDays.THURSDAY.Start
		supportDay["end"] = supportDays.THURSDAY.End
		result["thursday"] = []any{supportDay}
	}
	if supportDays.FRIDAY != nil {
		supportDay := make(map[string]any)
		supportDay["start"] = supportDays.FRIDAY.Start
		supportDay["end"] = supportDays.FRIDAY.End
		result["friday"] = []any{supportDay}
	}
	if supportDays.SATURDAY != nil {
		supportDay := make(map[string]any)
		supportDay["start"] = supportDays.SATURDAY.Start
		supportDay["end"] = supportDays.SATURDAY.End
		result["saturday"] = []any{supportDay}
	}
	if supportDays.SUNDAY != nil {
		supportDay := make(map[string]any)
		supportDay["start"] = supportDays.SUNDAY.Start
		supportDay["end"] = supportDays.SUNDAY.End
		result["sunday"] = []any{supportDay}
	}
	results = append(results, result)

	return results, nil
}
