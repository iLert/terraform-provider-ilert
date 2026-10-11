package ilert

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/iLert/ilert-go/v4"
)

// unknownConfigValue is how the SDK marks a value unknown at plan time in a raw configuration,
// hcl2shim.UnknownVariableValue, which is internal to it.
const unknownConfigValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

func supportDayConfig(start, end string) []any {
	return []any{map[string]any{"start": start, "end": end}}
}

func newSupportWindow(fromDay, fromTime, toDay, toTime string) ilert.SupportWindow {
	return ilert.SupportWindow{
		From: &ilert.TimeOfWeek{DayOfWeek: fromDay, Time: fromTime},
		To:   &ilert.TimeOfWeek{DayOfWeek: toDay, Time: toTime},
	}
}

func describeSupportWindows(windows []ilert.SupportWindow) []string {
	result := make([]string, 0, len(windows))
	for _, window := range windows {
		result = append(result, fmt.Sprintf("%s %s to %s %s", window.From.DayOfWeek, window.From.Time, window.To.DayOfWeek, window.To.Time))
	}
	return result
}

func supportHourConfig(coverage map[string]any) map[string]any {
	config := map[string]any{"name": "test-support-hour", "timezone": "Europe/Berlin"}
	for k, v := range coverage {
		config[k] = v
	}
	return config
}

func TestBuildSupportHour_SendsSupportDaysAsOneWindowPerDay(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, supportHourConfig(map[string]any{
		"support_days": []any{map[string]any{
			"friday": supportDayConfig("8:00", "12:00:00"),
			"monday": supportDayConfig("09:00", "17:00"),
		}},
	}))

	supportHour, err := buildSupportHour(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if supportHour.SupportDays != nil {
		t.Errorf("SupportDays = %+v, want nil: the coverage is only ever sent as windows", supportHour.SupportDays)
	}
	if supportHour.SupportWindows == nil {
		t.Fatal("SupportWindows = nil, want the support days as windows")
	}
	// in day order, with the times the API read leniently as a support day in the HH:mm it takes
	// window times in
	want := []string{"MONDAY 09:00 to MONDAY 17:00", "FRIDAY 08:00 to FRIDAY 12:00"}
	if got := describeSupportWindows(*supportHour.SupportWindows); !reflect.DeepEqual(got, want) {
		t.Errorf("SupportWindows = %v, want %v", got, want)
	}

	payload, err := json.Marshal(supportHour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(payload), `"supportDays":null`) {
		t.Errorf("payload = %s, want it to send no support days", payload)
	}
}

func TestBuildSupportHour_SendsSupportWindowsAsConfigured(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, supportHourConfig(map[string]any{
		"support_windows": []any{
			newRestrictionConfig("MONDAY", "09:00", "MONDAY", "12:00"),
			newRestrictionConfig("MONDAY", "13:00", "MONDAY", "17:00"),
			newRestrictionConfig("FRIDAY", "17:00", "MONDAY", "08:00"),
		},
	}))

	supportHour, err := buildSupportHour(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if supportHour.SupportDays != nil {
		t.Errorf("SupportDays = %+v, want nil", supportHour.SupportDays)
	}
	want := []string{"MONDAY 09:00 to MONDAY 12:00", "MONDAY 13:00 to MONDAY 17:00", "FRIDAY 17:00 to MONDAY 08:00"}
	if got := describeSupportWindows(*supportHour.SupportWindows); !reflect.DeepEqual(got, want) {
		t.Errorf("SupportWindows = %v, want %v", got, want)
	}
}

// TestBuildSupportHour_SendsEmptyWindowsWithoutCoverage pins that no coverage is sent as an
// explicit empty list: the API clears the coverage on [], and rejects a write naming neither field.
func TestBuildSupportHour_SendsEmptyWindowsWithoutCoverage(t *testing.T) {
	cases := map[string]map[string]any{
		"empty support_days block": {"support_days": []any{map[string]any{}}},
		"no coverage at all":       {},
	}

	for name, coverage := range cases {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, supportHourConfig(coverage))

			supportHour, err := buildSupportHour(d)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			payload, err := json.Marshal(supportHour)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(string(payload), `"supportWindows":[]`) {
				t.Errorf("payload = %s, want it to contain \"supportWindows\":[]", payload)
			}
		})
	}
}

// TestBuildSupportHour_RejectsSupportDayNotEndingAfterItStarts pins the check the API made of a
// support day: sent as a window, 09:00 to 09:00 would be the whole week, and 17:00 to 09:00 would
// wrap around it.
func TestBuildSupportHour_RejectsSupportDayNotEndingAfterItStarts(t *testing.T) {
	cases := map[string][]any{
		"start equals end": supportDayConfig("09:00", "09:00"),
		"start after end":  supportDayConfig("17:00", "09:00"),
		"not a time":       supportDayConfig("09:00", "24:00"),
		"seconds":          supportDayConfig("09:00:30", "17:00"),
	}

	for name, day := range cases {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, supportHourConfig(map[string]any{
				"support_days": []any{map[string]any{"tuesday": day}},
			}))

			_, err := buildSupportHour(d)
			if err == nil || !strings.Contains(err.Error(), "support_days.0.tuesday") {
				t.Errorf("error = %v, want one naming support_days.0.tuesday", err)
			}
		})
	}
}

// TestTransformSupportHour_ShowsWindowsSupportDaysCannotHold is the drift fix: a window added in
// the web app to a support hour configured with support_days reads back as support_windows, so
// the plan shows it instead of reading the first window of the day and planning clean.
func TestTransformSupportHour_ShowsWindowsSupportDaysCannotHold(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, supportHourConfig(map[string]any{
		"support_days": []any{map[string]any{"monday": supportDayConfig("09:00", "12:00")}},
	}))
	windows := []ilert.SupportWindow{
		newSupportWindow("MONDAY", "09:00", "MONDAY", "12:00"),
		newSupportWindow("MONDAY", "13:00", "MONDAY", "17:00"),
	}

	err := transformSupportHourResource(&ilert.SupportHour{
		Name:           "test-support-hour",
		Timezone:       "Europe/Berlin",
		SupportDays:    &ilert.SupportDays{MONDAY: &ilert.SupportDay{Start: "09:00", End: "12:00"}},
		SupportWindows: &windows,
	}, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := d.Get("support_days").([]any); len(got) != 0 {
		t.Errorf("support_days = %v, want it empty", got)
	}
	want := flattenSupportWindows(windows)
	if got := d.Get("support_windows").([]any); !reflect.DeepEqual(got, want) {
		t.Errorf("support_windows = %v, want %v", got, want)
	}
}

func TestTransformSupportHour_SetsSupportDaysThatDescribeTheCoverage(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, supportHourConfig(map[string]any{
		"support_days": []any{map[string]any{"monday": supportDayConfig("09:00", "17:00")}},
	}))

	err := transformSupportHourResource(&ilert.SupportHour{
		Name:        "test-support-hour",
		Timezone:    "Europe/Berlin",
		SupportDays: &ilert.SupportDays{MONDAY: &ilert.SupportDay{Start: "09:00", End: "17:00"}},
	}, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := d.Get("support_days.0.monday.0.end"); got != "17:00" {
		t.Errorf("support_days.0.monday.0.end = %v, want 17:00", got)
	}
	if got := d.Get("support_windows").([]any); len(got) != 0 {
		t.Errorf("support_windows = %v, want it empty", got)
	}
}

// TestTransformSupportHour_KeepsWindowsThatFitOneWindowPerDay pins the case the API answers with
// supportDays only: windows that fit one per day. Set as support_days, every plan of a
// configuration written with windows would show them as a diff.
func TestTransformSupportHour_KeepsWindowsThatFitOneWindowPerDay(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, supportHourConfig(map[string]any{
		"support_windows": []any{
			newRestrictionConfig("MONDAY", "09:00", "MONDAY", "17:00"),
			newRestrictionConfig("WEDNESDAY", "09:00", "WEDNESDAY", "23:59"),
		},
	}))

	err := transformSupportHourResource(&ilert.SupportHour{
		Name:     "test-support-hour",
		Timezone: "Europe/Berlin",
		SupportDays: &ilert.SupportDays{
			MONDAY:    &ilert.SupportDay{Start: "09:00", End: "17:00"},
			WEDNESDAY: &ilert.SupportDay{Start: "09:00", End: "23:59"},
		},
	}, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := flattenSupportWindows([]ilert.SupportWindow{
		newSupportWindow("MONDAY", "09:00", "MONDAY", "17:00"),
		newSupportWindow("WEDNESDAY", "09:00", "WEDNESDAY", "23:59"),
	})
	if got := d.Get("support_windows").([]any); !reflect.DeepEqual(got, want) {
		t.Errorf("support_windows = %v, want %v", got, want)
	}
	if got := d.Get("support_days").([]any); len(got) != 0 {
		t.Errorf("support_days = %v, want it empty", got)
	}
}

// TestTransformSupportHour_ImportsWindowsOnlyWhenNeeded covers an import and the IaC export, which
// read without a configuration: support_windows only where support_days cannot hold the coverage.
func TestTransformSupportHour_ImportsWindowsOnlyWhenNeeded(t *testing.T) {
	windows := []ilert.SupportWindow{newSupportWindow("FRIDAY", "17:00", "MONDAY", "09:00")}
	cases := map[string]struct {
		supportHour     *ilert.SupportHour
		wantDays        int
		wantWindowCount int
	}{
		"coverage across the weekend": {
			supportHour: &ilert.SupportHour{
				SupportDays: &ilert.SupportDays{
					FRIDAY:   &ilert.SupportDay{Start: "17:00", End: "23:59"},
					SATURDAY: &ilert.SupportDay{Start: "00:00", End: "23:59"},
					SUNDAY:   &ilert.SupportDay{Start: "00:00", End: "23:59"},
					MONDAY:   &ilert.SupportDay{Start: "00:00", End: "09:00"},
				},
				SupportWindows: &windows,
			},
			wantWindowCount: 1,
		},
		"coverage of one window per day": {
			supportHour: &ilert.SupportHour{
				SupportDays: &ilert.SupportDays{MONDAY: &ilert.SupportDay{Start: "09:00", End: "17:00"}},
			},
			wantDays: 1,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceSupportHour().Schema, nil)
			tc.supportHour.Name = "test-support-hour"
			tc.supportHour.Timezone = "Europe/Berlin"

			if err := transformSupportHourResource(tc.supportHour, d); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := len(d.Get("support_days").([]any)); got != tc.wantDays {
				t.Errorf("support_days has %d entries, want %d", got, tc.wantDays)
			}
			if got := len(d.Get("support_windows").([]any)); got != tc.wantWindowCount {
				t.Errorf("support_windows has %d entries, want %d", got, tc.wantWindowCount)
			}
		})
	}
}

func TestValidateSupportWindows(t *testing.T) {
	cases := []struct {
		name    string
		windows []ilert.SupportWindow
		// wantErr is empty for windows written the way the API returns them, else a part of the
		// error, and wantStored the windows the error says the API stores
		wantErr    string
		wantStored []string
	}{
		{
			name: "split shift and a weekend",
			windows: []ilert.SupportWindow{
				newSupportWindow("MONDAY", "09:00", "MONDAY", "12:00"),
				newSupportWindow("MONDAY", "13:00", "MONDAY", "17:00"),
				newSupportWindow("FRIDAY", "17:00", "MONDAY", "08:00"),
			},
		},
		{
			name:    "the whole week",
			windows: []ilert.SupportWindow{newSupportWindow("MONDAY", "00:00", "MONDAY", "00:00")},
		},
		{
			name:    "up to Sunday midnight",
			windows: []ilert.SupportWindow{newSupportWindow("SUNDAY", "20:00", "MONDAY", "00:00")},
		},
		{
			name:    "from Monday 00:00",
			windows: []ilert.SupportWindow{newSupportWindow("MONDAY", "00:00", "MONDAY", "09:00")},
		},
		{
			name: "unsorted",
			windows: []ilert.SupportWindow{
				newSupportWindow("MONDAY", "13:00", "MONDAY", "17:00"),
				newSupportWindow("MONDAY", "09:00", "MONDAY", "12:00"),
			},
			wantErr:    "ascending order of from",
			wantStored: []string{"MONDAY 09:00 to MONDAY 12:00", "MONDAY 13:00 to MONDAY 17:00"},
		},
		{
			name: "the window across Sunday midnight first",
			windows: []ilert.SupportWindow{
				newSupportWindow("FRIDAY", "17:00", "MONDAY", "08:00"),
				newSupportWindow("MONDAY", "09:00", "MONDAY", "12:00"),
			},
			wantErr:    "ascending order of from",
			wantStored: []string{"MONDAY 09:00 to MONDAY 12:00", "FRIDAY 17:00 to MONDAY 08:00"},
		},
		{
			name: "overlapping",
			windows: []ilert.SupportWindow{
				newSupportWindow("MONDAY", "09:00", "MONDAY", "13:00"),
				newSupportWindow("MONDAY", "12:00", "MONDAY", "17:00"),
			},
			wantErr:    "overlap or touch",
			wantStored: []string{"MONDAY 09:00 to MONDAY 17:00"},
		},
		{
			name: "touching",
			windows: []ilert.SupportWindow{
				newSupportWindow("MONDAY", "09:00", "MONDAY", "12:00"),
				newSupportWindow("MONDAY", "12:00", "MONDAY", "17:00"),
			},
			wantErr:    "overlap or touch",
			wantStored: []string{"MONDAY 09:00 to MONDAY 17:00"},
		},
		{
			name: "touching at Sunday midnight",
			windows: []ilert.SupportWindow{
				newSupportWindow("MONDAY", "00:00", "MONDAY", "09:00"),
				newSupportWindow("SUNDAY", "20:00", "MONDAY", "00:00"),
			},
			wantErr:    "overlap or touch",
			wantStored: []string{"SUNDAY 20:00 to MONDAY 09:00"},
		},
		{
			name:       "the whole week starting elsewhere",
			windows:    []ilert.SupportWindow{newSupportWindow("WEDNESDAY", "10:00", "WEDNESDAY", "10:00")},
			wantErr:    "covers the whole week",
			wantStored: []string{"MONDAY 00:00 to MONDAY 00:00"},
		},
		{
			name: "the whole week next to another window",
			windows: []ilert.SupportWindow{
				newSupportWindow("MONDAY", "00:00", "MONDAY", "00:00"),
				newSupportWindow("TUESDAY", "09:00", "TUESDAY", "17:00"),
			},
			wantErr:    "covers the whole week",
			wantStored: []string{"MONDAY 00:00 to MONDAY 00:00"},
		},
		{
			name: "the whole week in two halves",
			windows: []ilert.SupportWindow{
				newSupportWindow("MONDAY", "00:00", "THURSDAY", "00:00"),
				newSupportWindow("THURSDAY", "00:00", "MONDAY", "00:00"),
			},
			wantErr:    "overlap or touch",
			wantStored: []string{"MONDAY 00:00 to MONDAY 00:00"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSupportWindows(tc.windows)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("error = nil, want one containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantErr)
			}
			stored := err.Error()[strings.LastIndex(err.Error(), "as:\n\n")+len("as:\n\n"):]
			if got := strings.Split(stored, "\n"); !reflect.DeepEqual(got, prefixAll("  ", tc.wantStored)) {
				t.Errorf("stored windows = %q, want %q", got, tc.wantStored)
			}
		})
	}
}

func prefixAll(prefix string, values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, prefix+value)
	}
	return result
}

// TestResourceSupportHour_PlanChecksCoverage runs the checks through a plan of the resource, the
// way Terraform does, including values that are not known until apply.
func TestResourceSupportHour_PlanChecksCoverage(t *testing.T) {
	cases := map[string]struct {
		coverage map[string]any
		wantErr  string
	}{
		"windows in the order the API returns them": {
			coverage: map[string]any{"support_windows": []any{
				newRestrictionConfig("MONDAY", "09:00", "MONDAY", "12:00"),
				newRestrictionConfig("FRIDAY", "17:00", "MONDAY", "08:00"),
			}},
		},
		"unsorted windows": {
			coverage: map[string]any{"support_windows": []any{
				newRestrictionConfig("FRIDAY", "17:00", "MONDAY", "08:00"),
				newRestrictionConfig("MONDAY", "09:00", "MONDAY", "12:00"),
			}},
			wantErr: "ascending order of from",
		},
		"a window not known until apply": {
			coverage: map[string]any{"support_windows": []any{
				newRestrictionConfig("FRIDAY", "17:00", "MONDAY", "08:00"),
				newRestrictionConfig("MONDAY", unknownConfigValue, "MONDAY", "12:00"),
			}},
		},
		"support days": {
			coverage: map[string]any{"support_days": []any{map[string]any{"monday": supportDayConfig("09:00", "17:00")}}},
		},
		"a support day that starts where it ends": {
			coverage: map[string]any{"support_days": []any{map[string]any{"monday": supportDayConfig("09:00", "09:00")}}},
			wantErr:  "support_days.0.monday",
		},
		"a support day not known until apply": {
			coverage: map[string]any{"support_days": []any{map[string]any{"monday": supportDayConfig(unknownConfigValue, "09:00")}}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			config := terraform.NewResourceConfigRaw(supportHourConfig(tc.coverage))
			_, err := resourceSupportHour().Diff(context.Background(), nil, config, nil)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %v, want one containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestDataSourceSupportHour_SupportWindowsAreTheCompleteCoverage(t *testing.T) {
	weekend := []ilert.SupportWindow{newSupportWindow("FRIDAY", "17:00", "MONDAY", "09:00")}
	cases := map[string]struct {
		supportHour *ilert.SupportHour
		want        []string
	}{
		"returned as windows": {
			supportHour: &ilert.SupportHour{
				SupportDays:    &ilert.SupportDays{FRIDAY: &ilert.SupportDay{Start: "17:00", End: "23:59"}},
				SupportWindows: &weekend,
			},
			want: []string{"FRIDAY 17:00 to MONDAY 09:00"},
		},
		"returned as support days only": {
			supportHour: &ilert.SupportHour{SupportDays: &ilert.SupportDays{
				MONDAY:   &ilert.SupportDay{Start: "09:00", End: "17:00"},
				SATURDAY: &ilert.SupportDay{Start: "10:00", End: "14:00"},
			}},
			want: []string{"MONDAY 09:00 to MONDAY 17:00", "SATURDAY 10:00 to SATURDAY 14:00"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, dataSourceSupportHour().Schema, map[string]any{"name": "test-support-hour"})
			if err := d.Set("support_windows", flattenSupportWindows(supportHourWindows(tc.supportHour))); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := expandSupportWindows(d.Get("support_windows").([]any))
			if !reflect.DeepEqual(describeSupportWindows(got), tc.want) {
				t.Errorf("support_windows = %v, want %v", describeSupportWindows(got), tc.want)
			}
		})
	}
}

// TestAccSupportHour_coverage runs against the API: it needs TF_ACC and credentials, and creates
// and deletes a support hour in that account.
func TestAccSupportHour_coverage(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-support-hour")
	resourceName := "ilert_support_hour.test"
	var supportHourID int64

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSupportHourDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSupportHourConfig(name, testAccSupportHourMorning),
				Check: resource.ComposeTestCheckFunc(
					testAccCaptureSupportHourID(resourceName, &supportHourID),
					resource.TestCheckResourceAttr(resourceName, "support_days.0.monday.0.end", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "support_windows.#", "0"),
					testAccCheckSupportHourCoverage(resourceName, "MONDAY 09:00 to MONDAY 12:00"),
				),
			},
			{
				// a window added in the web app, which support_days cannot show, has to reach the plan
				PreConfig: func() {
					testAccSetSupportHourWindowsOutsideTerraform(t, supportHourID,
						newSupportWindow("MONDAY", "09:00", "MONDAY", "12:00"),
						newSupportWindow("MONDAY", "13:00", "MONDAY", "17:00"))
				},
				Config:             testAccSupportHourConfig(name, testAccSupportHourMorning),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// and applying the configuration replaces it, after which the plan is clean
				Config: testAccSupportHourConfig(name, testAccSupportHourMorning),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "support_days.0.monday.0.end", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "support_windows.#", "0"),
					testAccCheckSupportHourCoverage(resourceName, "MONDAY 09:00 to MONDAY 12:00"),
				),
			},
			{
				Config: testAccSupportHourConfig(name, testAccSupportHourSplitShiftAndWeekend),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "support_windows.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "support_days.#", "0"),
					testAccCheckSupportHourCoverage(resourceName,
						"MONDAY 09:00 to MONDAY 12:00", "MONDAY 13:00 to MONDAY 17:00", "FRIDAY 17:00 to MONDAY 08:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// windows that fit one per day come back from the API as support days
				Config: testAccSupportHourConfig(name, testAccSupportHourWindowPerDay),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "support_windows.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "support_days.#", "0"),
					testAccCheckSupportHourCoverage(resourceName, "MONDAY 09:00 to MONDAY 17:00"),
				),
			},
			{
				Config: testAccSupportHourConfig(name, testAccSupportHourWholeWeek),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "support_windows.#", "1"),
					testAccCheckSupportHourCoverage(resourceName, "MONDAY 00:00 to MONDAY 00:00"),
				),
			},
			{
				Config: testAccSupportHourConfig(name, testAccSupportHourMorning),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "support_days.0.monday.0.end", "12:00"),
					resource.TestCheckResourceAttr(resourceName, "support_windows.#", "0"),
					testAccCheckSupportHourCoverage(resourceName, "MONDAY 09:00 to MONDAY 12:00"),
				),
			},
		},
	})
}

const testAccSupportHourMorning = `
  support_days {
    monday {
      start = "09:00"
      end   = "12:00"
    }
  }
`

const testAccSupportHourSplitShiftAndWeekend = `
  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "09:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "12:00"
    }
  }

  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "13:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "17:00"
    }
  }

  support_windows {
    from {
      day_of_week = "FRIDAY"
      time        = "17:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "08:00"
    }
  }
`

const testAccSupportHourWindowPerDay = `
  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "09:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "17:00"
    }
  }
`

const testAccSupportHourWholeWeek = `
  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "00:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "00:00"
    }
  }
`

func testAccSupportHourConfig(name, coverage string) string {
	return fmt.Sprintf(`
resource "ilert_support_hour" "test" {
  name     = %q
  timezone = "Europe/Berlin"
%s
}
`, name, coverage)
}

func testAccCaptureSupportHourID(resourceName string, id *int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		parsed, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}
		*id = parsed
		return nil
	}
}

// testAccSetSupportHourWindowsOutsideTerraform replaces the coverage through the API, the way the
// web app does when a window is added there.
func testAccSetSupportHourWindowsOutsideTerraform(t *testing.T, id int64, windows ...ilert.SupportWindow) {
	client := testAccProvider.Meta().(*ilert.Client)

	result, err := client.GetSupportHour(&ilert.GetSupportHourInput{SupportHourID: ilert.Int64(id)})
	if err != nil {
		t.Fatalf("could not read support hour %d: %v", id, err)
	}
	supportHour := result.SupportHour
	supportHour.SupportDays = nil
	supportHour.SupportWindows = &windows

	if _, err := client.UpdateSupportHour(&ilert.UpdateSupportHourInput{SupportHourID: ilert.Int64(id), SupportHour: supportHour}); err != nil {
		t.Fatalf("could not update support hour %d: %v", id, err)
	}
}

// testAccCheckSupportHourCoverage reads the coverage of the support hour from the API, so a step
// cannot pass on a state that agrees with the configuration while the coverage stored differs.
func testAccCheckSupportHourCoverage(resourceName string, want ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		client := testAccProvider.Meta().(*ilert.Client)
		result, err := client.GetSupportHour(&ilert.GetSupportHourInput{SupportHourID: ilert.Int64(id)})
		if err != nil {
			return err
		}
		if got := describeSupportWindows(supportHourWindows(result.SupportHour)); !reflect.DeepEqual(got, want) {
			return fmt.Errorf("coverage of support hour %d = %v, want %v", id, got, want)
		}
		return nil
	}
}

func testAccCheckSupportHourDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ilert.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ilert_support_hour" {
			continue
		}
		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}
		_, err = client.GetSupportHour(&ilert.GetSupportHourInput{SupportHourID: ilert.Int64(id)})
		if err == nil {
			return fmt.Errorf("support hour %d still exists", id)
		}
		if _, ok := err.(*ilert.NotFoundAPIError); !ok {
			return err
		}
	}
	return nil
}
