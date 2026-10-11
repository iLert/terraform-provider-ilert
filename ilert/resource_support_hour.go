package ilert

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/iLert/ilert-go/v4"
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
				// TypeSet, not TypeList: the API returns teams sorted by id, so an
				// ordered list produces a permanent diff whenever the config order
				// differs from that.
				Type:     schema.TypeSet,
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
				Optional: true,
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
				ExactlyOneOf: []string{"support_days", "support_windows"},
			},
			"support_windows": {
				// TypeList, not TypeSet, held to the order the API returns the windows in by
				// validateSupportWindows: merged, and ascending by from.
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: maxSupportWindows,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from": supportWindowTimeOfWeekSchema(),
						"to":   supportWindowTimeOfWeekSchema(),
					},
				},
				ExactlyOneOf: []string{"support_days", "support_windows"},
			},
			"exception": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
						"support_status": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ilert.SupportStatus.During,
							ValidateFunc: validation.StringInSlice(ilert.SupportStatusAll, false),
						},
					},
				},
			},
		},
		CustomizeDiff: validateSupportHourCoverage,
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
		vL := val.(*schema.Set).List()
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
	} else if d.HasChange("team") {
		// All team blocks removed. The API only clears teams on an explicit empty
		// array; an omitted or null field leaves them untouched. HasChange keeps
		// this to support hours whose teams were actually dropped from the config:
		// without it every update on a config that never declared a team block
		// would clear the teams assigned to it elsewhere.
		supportHour.Teams = []ilert.TeamShort{}
	}

	// The coverage is only ever sent as supportWindows, never as supportDays, so the configuration
	// is authoritative. The API keeps the stored coverage when it is sent back the per-day view it
	// returned, which would keep windows added in the web app that support_days cannot show, and it
	// rejects a per-day view that contradicts the stored coverage or the windows sent next to it.
	var windows []ilert.SupportWindow
	if val, ok := d.GetOk("support_windows"); ok {
		windows = expandSupportWindows(val.([]any))
	} else {
		var err error
		windows, err = supportDaysToWindows(d.Get("support_days").([]any))
		if err != nil {
			return nil, err
		}
	}
	supportHour.SupportWindows = &windows

	if val, ok := d.GetOk("exception"); ok {
		vL := val.([]any)
		exceptions := make([]ilert.SupportHourException, 0, len(vL))
		for _, exception := range vL {
			if exception == nil {
				continue
			}

			v := exception.(map[string]any)
			ex := ilert.SupportHourException{
				Start:         v["start"].(string),
				End:           v["end"].(string),
				SupportStatus: ilert.SupportStatus.During,
			}

			if v["name"] != nil && v["name"].(string) != "" {
				ex.Name = v["name"].(string)
			}

			if v["support_status"] != nil && v["support_status"].(string) != "" {
				ex.SupportStatus = v["support_status"].(string)
			}

			exceptions = append(exceptions, ex)
		}

		supportHour.Exceptions = exceptions
	}

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

	err = transformSupportHourResource(result.SupportHour, d)
	if err != nil {
		return diag.FromErr(err)
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

func transformSupportHourResource(supportHour *ilert.SupportHour, d *schema.ResourceData) error {
	d.Set("name", supportHour.Name)
	d.Set("timezone", supportHour.Timezone)

	teams, err := flattenTeamShortList(supportHour.Teams, d)
	if err != nil {
		return fmt.Errorf("[ERROR] Error flattening teams: %s", err.Error())
	}
	if err := d.Set("team", teams); err != nil {
		return fmt.Errorf("[ERROR] Error setting teams: %s", err.Error())
	}

	if err := setSupportHourCoverage(supportHour, d); err != nil {
		return err
	}

	exceptions := flattenSupportHourExceptions(supportHour.Exceptions)
	if err := d.Set("exception", exceptions); err != nil {
		return fmt.Errorf("[ERROR] Error setting exceptions: %s", err.Error())
	}

	return nil
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

// setSupportHourCoverage sets the weekly coverage as support_windows or support_days, leaving
// the other one empty.
//
// The API only returns supportWindows for coverage that supportDays cannot express, and its
// supportDays is lossy then: each day reports its first window. Such coverage is always set as
// support_windows, so a configuration written with support_days shows the windows it cannot hold
// as a diff instead of planning clean while they keep paging. Coverage that supportDays does
// express is set as whichever of the two the resource uses, the state of a refresh or the plan of
// an apply, since both describe it exactly: windows that fit one per day come back as supportDays,
// which would otherwise show as a diff on every plan of a configuration written with windows.
func setSupportHourCoverage(supportHour *ilert.SupportHour, d *schema.ResourceData) error {
	supportDays := make([]any, 0)
	var supportWindows []any

	switch {
	// the API returns null rather than [] for coverage that supportDays describes, but the IaC
	// export runs this on JSON the web app sends, where an empty list must not hide supportDays
	case supportHour.SupportWindows != nil && len(*supportHour.SupportWindows) > 0:
		supportWindows = flattenSupportWindows(*supportHour.SupportWindows)
	case len(d.Get("support_windows").([]any)) > 0:
		supportWindows = flattenSupportWindows(supportDaysAsWindows(supportHour.SupportDays))
	default:
		days, err := flattenSupportDays(supportHour.SupportDays)
		if err != nil {
			return fmt.Errorf("[ERROR] Error flattening support days: %s", err.Error())
		}
		supportDays = days
		supportWindows = make([]any, 0)
	}

	if err := d.Set("support_days", supportDays); err != nil {
		return fmt.Errorf("[ERROR] Error setting support days: %s", err.Error())
	}
	if err := d.Set("support_windows", supportWindows); err != nil {
		return fmt.Errorf("[ERROR] Error setting support windows: %s", err.Error())
	}

	return nil
}

// maxSupportWindows is the most windows the API takes in a single write.
const maxSupportWindows = 84

const (
	minutesPerDay  = 24 * 60
	minutesPerWeek = 7 * minutesPerDay
)

// supportWindowTimeRegexp matches the HH:mm the API returns window times in. It accepts a zero
// seconds part on input as well, but answers with HH:mm, so that would show as a diff on every plan.
var supportWindowTimeRegexp = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// supportDayTimeRegexp matches the support day times the API has always read: H:mm, HH:mm and
// HH:mm:ss.
var supportDayTimeRegexp = regexp.MustCompile(`^([0-9]{1,2}):([0-9]{2})(?::([0-9]{2}))?$`)

func supportWindowTimeOfWeekSchema() *schema.Schema {
	return &schema.Schema{
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
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(supportWindowTimeRegexp, "must be a time of day in HH:mm, e.g. 09:00"),
				},
			},
		},
	}
}

func expandSupportWindows(vL []any) []ilert.SupportWindow {
	windows := make([]ilert.SupportWindow, 0, len(vL))
	for _, item := range vL {
		v, ok := item.(map[string]any)
		if !ok {
			continue
		}
		windows = append(windows, ilert.SupportWindow{
			From: expandTimeOfWeek(v["from"]),
			To:   expandTimeOfWeek(v["to"]),
		})
	}

	return windows
}

func expandTimeOfWeek(val any) *ilert.TimeOfWeek {
	vL, _ := val.([]any)
	if len(vL) == 0 || vL[0] == nil {
		return nil
	}
	v := vL[0].(map[string]any)

	return &ilert.TimeOfWeek{
		DayOfWeek: v["day_of_week"].(string),
		Time:      v["time"].(string),
	}
}

func flattenSupportWindows(windows []ilert.SupportWindow) []any {
	results := make([]any, 0, len(windows))
	for _, window := range windows {
		if window.From == nil || window.To == nil {
			continue
		}
		results = append(results, map[string]any{
			"from": []any{map[string]any{"day_of_week": window.From.DayOfWeek, "time": window.From.Time}},
			"to":   []any{map[string]any{"day_of_week": window.To.DayOfWeek, "time": window.To.Time}},
		})
	}

	return results
}

// supportDaysToWindows turns the support_days of a configuration into one window per day, from
// start to end on that day. That is the window the API builds from a support day itself, so the
// coverage is the same as when support_days was sent as it is.
func supportDaysToWindows(vL []any) ([]ilert.SupportWindow, error) {
	windows := make([]ilert.SupportWindow, 0)
	if len(vL) == 0 || vL[0] == nil {
		return windows, nil
	}

	days := vL[0].(map[string]any)
	for _, dayOfWeek := range ilert.DayOfWeekAll {
		day, ok := supportDayFromConfig(days, dayOfWeek)
		if !ok {
			continue
		}
		start, end, err := normalizeSupportDay(dayOfWeek, day["start"].(string), day["end"].(string))
		if err != nil {
			return nil, err
		}
		windows = append(windows, ilert.SupportWindow{
			From: &ilert.TimeOfWeek{DayOfWeek: dayOfWeek, Time: start},
			To:   &ilert.TimeOfWeek{DayOfWeek: dayOfWeek, Time: end},
		})
	}

	return windows, nil
}

func supportDayFromConfig(days map[string]any, dayOfWeek string) (map[string]any, bool) {
	dL, _ := days[strings.ToLower(dayOfWeek)].([]any)
	if len(dL) == 0 || dL[0] == nil {
		return nil, false
	}
	day, ok := dL[0].(map[string]any)
	return day, ok
}

// normalizeSupportDay returns start and end of a support day in the HH:mm that window times are
// sent in. Support days were sent as they were written and the API read their times leniently, so
// 8:00 keeps working. A support day has to end after it starts: as a day it was rejected
// otherwise, while as a window 09:00 to 09:00 is the whole week, and 17:00 to 09:00 wraps around it.
func normalizeSupportDay(dayOfWeek, start, end string) (string, string, error) {
	field := "support_days.0." + strings.ToLower(dayOfWeek)

	startMinute, err := parseSupportDayTime(start)
	if err != nil {
		return "", "", fmt.Errorf("%s.0.start %q %s", field, start, err.Error())
	}
	endMinute, err := parseSupportDayTime(end)
	if err != nil {
		return "", "", fmt.Errorf("%s.0.end %q %s", field, end, err.Error())
	}
	if startMinute >= endMinute {
		return "", "", fmt.Errorf("%s: the start %s is not before the end %s. A support day covers the time from start to end on that day, use support_windows for coverage across midnight", field, start, end)
	}

	return formatMinuteOfDay(startMinute), formatMinuteOfDay(endMinute), nil
}

// parseSupportDayTime returns the minute of the day a support day time stands for.
func parseSupportDayTime(value string) (int, error) {
	match := supportDayTimeRegexp.FindStringSubmatch(value)
	if match == nil {
		return 0, fmt.Errorf("is not a time of day, use HH:mm, e.g. 09:00")
	}
	hour, _ := strconv.Atoi(match[1])
	minute, _ := strconv.Atoi(match[2])
	if hour > 23 || minute > 59 {
		return 0, fmt.Errorf("is not a time of day, use HH:mm, e.g. 09:00")
	}
	if match[3] != "" && match[3] != "00" {
		return 0, fmt.Errorf("is not a whole minute, use HH:mm, e.g. 09:00")
	}

	return hour*60 + minute, nil
}

func formatMinuteOfDay(minute int) string {
	return fmt.Sprintf("%02d:%02d", minute/60, minute%60)
}

// supportDaysAsWindows returns the per-day view the API returned as the windows it describes. Only
// exact while the API returns no supportWindows next to it.
func supportDaysAsWindows(supportDays *ilert.SupportDays) []ilert.SupportWindow {
	windows := make([]ilert.SupportWindow, 0)
	if supportDays == nil {
		return windows
	}

	days := []*ilert.SupportDay{
		supportDays.MONDAY,
		supportDays.TUESDAY,
		supportDays.WEDNESDAY,
		supportDays.THURSDAY,
		supportDays.FRIDAY,
		supportDays.SATURDAY,
		supportDays.SUNDAY,
	}
	for i, day := range days {
		if day == nil {
			continue
		}
		windows = append(windows, ilert.SupportWindow{
			From: &ilert.TimeOfWeek{DayOfWeek: ilert.DayOfWeekAll[i], Time: day.Start},
			To:   &ilert.TimeOfWeek{DayOfWeek: ilert.DayOfWeekAll[i], Time: day.End},
		})
	}

	return windows
}

// supportHourWindows returns the complete coverage of a support hour as windows, whichever of the
// two fields the API returned it in.
func supportHourWindows(supportHour *ilert.SupportHour) []ilert.SupportWindow {
	if supportHour.SupportWindows != nil && len(*supportHour.SupportWindows) > 0 {
		return *supportHour.SupportWindows
	}
	return supportDaysAsWindows(supportHour.SupportDays)
}

// validateSupportHourCoverage surfaces at plan time the coverage the API would misread or return
// differently. Values that are not known until apply are skipped, they get checked again when the
// plan is made with them.
func validateSupportHourCoverage(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if err := validateSupportDaysDiff(d); err != nil {
		return err
	}
	return validateSupportWindowsDiff(d)
}

// validateSupportDaysDiff rejects a support day that does not end after it starts at plan time.
// Sent as a window, 09:00 to 09:00 would be applied as the whole week instead of failing.
func validateSupportDaysDiff(d *schema.ResourceDiff) error {
	vL, _ := d.Get("support_days").([]any)
	if !d.NewValueKnown("support_days") || len(vL) == 0 || vL[0] == nil {
		return nil
	}

	days := vL[0].(map[string]any)
	for _, dayOfWeek := range ilert.DayOfWeekAll {
		day, ok := supportDayFromConfig(days, dayOfWeek)
		if !ok {
			continue
		}
		key := "support_days.0." + strings.ToLower(dayOfWeek) + ".0."
		if !d.NewValueKnown(key+"start") || !d.NewValueKnown(key+"end") {
			continue
		}
		if _, _, err := normalizeSupportDay(dayOfWeek, day["start"].(string), day["end"].(string)); err != nil {
			return err
		}
	}

	return nil
}

func validateSupportWindowsDiff(d *schema.ResourceDiff) error {
	vL, _ := d.Get("support_windows").([]any)
	if !d.NewValueKnown("support_windows") || len(vL) == 0 {
		return nil
	}
	for i := range vL {
		for _, key := range []string{"from.0.day_of_week", "from.0.time", "to.0.day_of_week", "to.0.time"} {
			if !d.NewValueKnown(fmt.Sprintf("support_windows.%d.%s", i, key)) {
				return nil
			}
		}
	}

	return validateSupportWindows(expandSupportWindows(vL))
}

// weekArc is a support window as minutes since Monday 00:00, covering [from, to) on the circular
// week: a to at or before from wraps past Sunday midnight, and from equal to to is the whole week.
type weekArc struct {
	from int
	to   int
}

// validateSupportWindows rejects windows that are not written the way the API returns them, since
// every plan would show the difference as a diff otherwise.
func validateSupportWindows(windows []ilert.SupportWindow) error {
	arcs := make([]weekArc, 0, len(windows))
	for _, window := range windows {
		from, ok := minuteOfWeek(window.From)
		if !ok {
			return nil // reported by the validation of the attribute itself
		}
		to, ok := minuteOfWeek(window.To)
		if !ok {
			return nil
		}
		arcs = append(arcs, weekArc{from: from, to: to})
	}

	canonical := canonicalWeekArcs(arcs)
	if slices.Equal(arcs, canonical) {
		return nil
	}

	var reason string
	switch {
	case slices.ContainsFunc(arcs, func(arc weekArc) bool { return arc.from == arc.to }):
		reason = "a window whose from equals its to covers the whole week, which the API returns as the single window MONDAY 00:00 to MONDAY 00:00"
	case len(canonical) < len(arcs):
		reason = "windows that overlap or touch are merged into one"
	default:
		reason = "windows are returned in ascending order of from, starting on Monday"
	}

	lines := make([]string, 0, len(canonical))
	for _, arc := range canonical {
		lines = append(lines, fmt.Sprintf("  %s to %s", formatMinuteOfWeek(arc.from), formatMinuteOfWeek(arc.to)))
	}

	return fmt.Errorf("support_windows must be written the way the API returns them, or every plan shows a diff: %s. The API stores these windows as:\n\n%s", reason, strings.Join(lines, "\n"))
}

func minuteOfWeek(timeOfWeek *ilert.TimeOfWeek) (int, bool) {
	if timeOfWeek == nil || !supportWindowTimeRegexp.MatchString(timeOfWeek.Time) {
		return 0, false
	}
	day := slices.Index(ilert.DayOfWeekAll, timeOfWeek.DayOfWeek)
	if day < 0 {
		return 0, false
	}
	hour, _ := strconv.Atoi(timeOfWeek.Time[:2])
	minute, _ := strconv.Atoi(timeOfWeek.Time[3:])

	return day*minutesPerDay + hour*60 + minute, true
}

func formatMinuteOfWeek(minute int) string {
	return ilert.DayOfWeekAll[minute/minutesPerDay] + " " + formatMinuteOfDay(minute%minutesPerDay)
}

// canonicalWeekArcs returns arcs the way the API stores and returns them: merged where they
// overlap or touch, in ascending order of from, and the whole week as the single arc from Monday
// 00:00 to Monday 00:00. An arc across Sunday midnight has the latest from, so it comes last.
func canonicalWeekArcs(arcs []weekArc) []weekArc {
	var covered [minutesPerWeek]bool
	for _, arc := range arcs {
		if arc.from == arc.to {
			return []weekArc{{from: 0, to: 0}}
		}
		for minute := arc.from; minute != arc.to; minute = (minute + 1) % minutesPerWeek {
			covered[minute] = true
		}
	}

	runs := make([]weekArc, 0)
	for minute := 0; minute < minutesPerWeek; {
		if !covered[minute] {
			minute++
			continue
		}
		start := minute
		for minute < minutesPerWeek && covered[minute] {
			minute++
		}
		// a run up to Sunday midnight ends at Monday 00:00, and one covering every minute is the
		// whole week
		runs = append(runs, weekArc{from: start, to: minute % minutesPerWeek})
	}

	// a run up to Sunday midnight and one from Monday 00:00 are a single arc across it
	if len(runs) > 1 && runs[0].from == 0 && runs[len(runs)-1].to == 0 {
		across := weekArc{from: runs[len(runs)-1].from, to: runs[0].to}
		runs = append(slices.Clone(runs[1:len(runs)-1]), across)
	}

	return runs
}

func flattenSupportHourExceptions(exceptionList []ilert.SupportHourException) []any {
	if exceptionList == nil {
		return make([]any, 0)
	}

	results := make([]any, 0)
	for _, exception := range exceptionList {
		result := map[string]any{
			"start": exception.Start,
			"end":   exception.End,
		}

		if exception.Name != "" {
			result["name"] = exception.Name
		}

		if exception.SupportStatus == "" {
			result["support_status"] = ilert.SupportStatus.During
		} else {
			result["support_status"] = exception.SupportStatus
		}

		results = append(results, result)
	}

	return results
}
