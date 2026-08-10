package ilert

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

func TestFlattenRestrictionListSorted_UsesConfigOrder(t *testing.T) {
	serverRestrictions := []ilert.LayerRestriction{
		newLayerRestriction("THURSDAY", "07:00", "THURSDAY", "18:00"),
		newLayerRestriction("MONDAY", "07:00", "MONDAY", "18:00"),
		newLayerRestriction("TUESDAY", "07:00", "TUESDAY", "18:00"),
	}

	configRestrictions := []any{
		newRestrictionConfig("MONDAY", "07:00", "MONDAY", "18:00"),
		newRestrictionConfig("TUESDAY", "07:00", "TUESDAY", "18:00"),
		newRestrictionConfig("THURSDAY", "07:00", "THURSDAY", "18:00"),
	}

	got, err := flattenRestrictionListSorted(serverRestrictions, configRestrictions)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	gotOrder := restrictionFromDayOrder(got)
	wantOrder := []string{"MONDAY", "TUESDAY", "THURSDAY"}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("expected order %v, got %v", wantOrder, gotOrder)
	}
}

func TestFlattenRestrictionListSorted_AppendsUnmatchedServerRestrictions(t *testing.T) {
	serverRestrictions := []ilert.LayerRestriction{
		newLayerRestriction("WEDNESDAY", "07:00", "WEDNESDAY", "18:00"),
		newLayerRestriction("MONDAY", "07:00", "MONDAY", "18:00"),
		newLayerRestriction("FRIDAY", "07:00", "FRIDAY", "18:00"),
	}

	configRestrictions := []any{
		newRestrictionConfig("MONDAY", "07:00", "MONDAY", "18:00"),
		newRestrictionConfig("TUESDAY", "07:00", "TUESDAY", "18:00"),
	}

	got, err := flattenRestrictionListSorted(serverRestrictions, configRestrictions)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	gotOrder := restrictionFromDayOrder(got)
	wantOrder := []string{"MONDAY", "WEDNESDAY", "FRIDAY"}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("expected order %v, got %v", wantOrder, gotOrder)
	}
}

func TestFlattenRestrictionListSorted_SkipsInvalidRestrictionEntries(t *testing.T) {
	validFrom := ilert.TimeOfWeek{DayOfWeek: "MONDAY", Time: "07:00"}
	serverRestrictions := []ilert.LayerRestriction{
		{
			From: &validFrom,
			To:   nil,
		},
		newLayerRestriction("TUESDAY", "07:00", "TUESDAY", "18:00"),
	}

	configRestrictions := []any{
		newRestrictionConfig("MONDAY", "07:00", "MONDAY", "18:00"),
		newRestrictionConfig("TUESDAY", "07:00", "TUESDAY", "18:00"),
	}

	got, err := flattenRestrictionListSorted(serverRestrictions, configRestrictions)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	gotOrder := restrictionFromDayOrder(got)
	wantOrder := []string{"TUESDAY"}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("expected order %v, got %v", wantOrder, gotOrder)
	}
}

func TestTransformScheduleResource_KeepsDefaultShiftDurationOnImport(t *testing.T) {
	// On import there is no configuration to read the duration back from. Dropping it
	// left the attribute null in state, so an exported configuration naming the duration
	// the API reported read as a permanent diff on every plan after the import.
	d := schema.TestResourceDataRaw(t, resourceSchedule().Schema, map[string]any{})

	schedule := &ilert.Schedule{
		Name:                 "test-schedule",
		Type:                 ilert.ScheduleType.Static,
		Timezone:             "Europe/Berlin",
		DefaultShiftDuration: "PT8H",
	}

	if err := transformScheduleResource(schedule, d); err != nil {
		t.Fatalf("unexpected error transforming schedule: %v", err)
	}

	if got := d.Get("default_shift_duration").(string); got != "PT8H" {
		t.Fatalf("expected default_shift_duration to be \"PT8H\", got %q", got)
	}
}

func TestTransformScheduleResource_DefaultShiftDurationFollowsTheAPI(t *testing.T) {
	// A configured duration must not survive a change made outside of Terraform, otherwise
	// the drift never shows up in the plan.
	d := schema.TestResourceDataRaw(t, resourceSchedule().Schema, map[string]any{
		"name":                   "test-schedule",
		"type":                   ilert.ScheduleType.Static,
		"timezone":               "Europe/Berlin",
		"default_shift_duration": "PT8H",
	})

	schedule := &ilert.Schedule{
		Name:                 "test-schedule",
		Type:                 ilert.ScheduleType.Static,
		Timezone:             "Europe/Berlin",
		DefaultShiftDuration: "PT24H",
	}

	if err := transformScheduleResource(schedule, d); err != nil {
		t.Fatalf("unexpected error transforming schedule: %v", err)
	}

	if got := d.Get("default_shift_duration").(string); got != "PT24H" {
		t.Fatalf("expected default_shift_duration to follow the API and be \"PT24H\", got %q", got)
	}
}

func newLayerRestriction(fromDay, fromTime, toDay, toTime string) ilert.LayerRestriction {
	from := ilert.TimeOfWeek{
		DayOfWeek: fromDay,
		Time:      fromTime,
	}
	to := ilert.TimeOfWeek{
		DayOfWeek: toDay,
		Time:      toTime,
	}
	return ilert.LayerRestriction{
		From: &from,
		To:   &to,
	}
}

func newRestrictionConfig(fromDay, fromTime, toDay, toTime string) map[string]any {
	return map[string]any{
		"from": []any{
			map[string]any{
				"day_of_week": fromDay,
				"time":        fromTime,
			},
		},
		"to": []any{
			map[string]any{
				"day_of_week": toDay,
				"time":        toTime,
			},
		},
	}
}

func restrictionFromDayOrder(restrictions []any) []string {
	order := make([]string, 0, len(restrictions))
	for _, restriction := range restrictions {
		restrictionMap, ok := restriction.(map[string]any)
		if !ok {
			continue
		}
		fromList, ok := restrictionMap["from"].([]any)
		if !ok || len(fromList) == 0 {
			continue
		}
		fromMap, ok := fromList[0].(map[string]any)
		if !ok {
			continue
		}
		dayOfWeek, ok := fromMap["day_of_week"].(string)
		if ok {
			order = append(order, dayOfWeek)
		}
	}
	return order
}
