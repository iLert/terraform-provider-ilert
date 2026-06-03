package ilert

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v3"
)

func TestParseAlertActionAlertSourceRelationID(t *testing.T) {
	cases := []struct {
		in           string
		wantActionID string
		wantSourceID int64
		wantErr      bool
	}{
		{"123/456", "123", 456, false},
		{"1/1", "1", 1, false},
		{"99999999999/88888888888", "99999999999", 88888888888, false},
		{"a5cc66dc-9851-4b25-8853-d65ea773924e/2344528", "a5cc66dc-9851-4b25-8853-d65ea773924e", 2344528, false},
		{"", "", 0, true},
		{"/", "", 0, true},
		{"123/", "", 0, true},
		{"/456", "", 0, true},
		{"123", "", 0, true},
		{"123/abc", "", 0, true},
		{"a5cc66dc-9851-4b25-8853-d65ea773924e/not-a-number", "", 0, true},
	}
	for _, tc := range cases {
		gotAction, gotSource, err := parseAlertActionAlertSourceRelationID(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseAlertActionAlertSourceRelationID(%q) expected error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseAlertActionAlertSourceRelationID(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if gotAction != tc.wantActionID || gotSource != tc.wantSourceID {
			t.Errorf("parseAlertActionAlertSourceRelationID(%q) = (%q, %d), want (%q, %d)",
				tc.in, gotAction, gotSource, tc.wantActionID, tc.wantSourceID)
		}
	}
}

func TestAlertActionContainsSource(t *testing.T) {
	src := func(id int64) ilertapi.AlertSource { return ilertapi.AlertSource{ID: id} }

	tests := []struct {
		name   string
		action *ilertapi.AlertActionOutput
		id     int64
		want   bool
	}{
		{"nil action", nil, 1, false},
		{"empty action", &ilertapi.AlertActionOutput{}, 1, false},
		{
			"present in AlertSources",
			&ilertapi.AlertActionOutput{AlertSources: &[]ilertapi.AlertSource{src(10), src(20)}},
			20, true,
		},
		{
			"absent from AlertSources",
			&ilertapi.AlertActionOutput{AlertSources: &[]ilertapi.AlertSource{src(10), src(20)}},
			30, false,
		},
		{
			"present in deprecated AlertSourceIDs",
			&ilertapi.AlertActionOutput{AlertSourceIDs: []int64{5, 6, 7}},
			6, true,
		},
		{
			"absent from deprecated AlertSourceIDs",
			&ilertapi.AlertActionOutput{AlertSourceIDs: []int64{5, 6, 7}},
			99, false,
		},
		{
			"present in both fields",
			&ilertapi.AlertActionOutput{
				AlertSources:   &[]ilertapi.AlertSource{src(1)},
				AlertSourceIDs: []int64{2},
			},
			2, true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := alertActionContainsSource(tc.action, tc.id)
			if got != tc.want {
				t.Errorf("alertActionContainsSource() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResourceAlertActionAlertSourceRelationImport(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertActionAlertSourceRelation().Schema, map[string]any{})
	d.SetId("a5cc66dc-9851-4b25-8853-d65ea773924e/2344528")

	states, err := resourceAlertActionAlertSourceRelationImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("unexpected error importing relation: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 state, got %d", len(states))
	}

	if got := states[0].Get("alert_action_id").(string); got != "a5cc66dc-9851-4b25-8853-d65ea773924e" {
		t.Errorf("alert_action_id = %q, want %q", got, "a5cc66dc-9851-4b25-8853-d65ea773924e")
	}
	if got := states[0].Get("alert_source_id").(string); got != "2344528" {
		t.Errorf("alert_source_id = %q, want %q", got, "2344528")
	}
}

func TestResourceAlertActionAlertSourceRelationImport_InvalidID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertActionAlertSourceRelation().Schema, map[string]any{})
	d.SetId("missing-source-id")

	if _, err := resourceAlertActionAlertSourceRelationImport(context.Background(), d, nil); err == nil {
		t.Fatal("expected error for malformed import ID, got nil")
	}
}

func TestAlertSourceIDValidation(t *testing.T) {
	validate := resourceAlertActionAlertSourceRelation().Schema["alert_source_id"].ValidateFunc
	if validate == nil {
		t.Fatal("expected a ValidateFunc on alert_source_id")
	}

	cases := []struct {
		in      string
		wantErr bool
	}{
		{"2344528", false},
		{"1", false},
		{"", true},
		{"abc", true},
		{"123abc", true},
		{"12.3", true},
		{"a5cc66dc-9851-4b25-8853-d65ea773924e", true},
	}
	for _, tc := range cases {
		_, errs := validate(tc.in, "alert_source_id")
		if tc.wantErr && len(errs) == 0 {
			t.Errorf("alert_source_id=%q: expected validation error, got none", tc.in)
		}
		if !tc.wantErr && len(errs) > 0 {
			t.Errorf("alert_source_id=%q: unexpected validation error: %v", tc.in, errs)
		}
	}
}

// TestKeyedMutex_SerializesSameKey asserts the keyed mutex provides mutual
// exclusion per key, so concurrent attach/detach on one alert_action_id cannot
// interleave (the bug guarded against). Run with -race to catch interleaving.
func TestKeyedMutex_SerializesSameKey(t *testing.T) {
	km := newKeyedMutex()
	const goroutines, increments = 20, 1000
	counter := 0

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				km.Lock("same-action")
				counter++ // unguarded read-modify-write, serialized only by the keyed mutex
				km.Unlock("same-action")
			}
		}()
	}
	wg.Wait()

	if want := goroutines * increments; counter != want {
		t.Errorf("counter = %d, want %d (lost updates => mutex not serializing)", counter, want)
	}
}

func TestIsAlreadyAttachedError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("boom"), false},
		{"not-found", &ilertapi.NotFoundAPIError{Message: "Could not find alert action"}, false},
		{"unrelated bad request", &ilertapi.BadRequestAPIError{Message: "invalid alertSourceId"}, false},
		{"already attached", &ilertapi.BadRequestAPIError{Message: "This alert action is already attached to this alert source"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAlreadyAttachedError(tc.err); got != tc.want {
				t.Errorf("isAlreadyAttachedError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsAlreadyDetachedError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("boom"), false},
		{"not-found", &ilertapi.NotFoundAPIError{Message: "Could not find alert source"}, false},
		{"unrelated bad request", &ilertapi.BadRequestAPIError{Message: "invalid alertSourceId"}, false},
		{"not attached", &ilertapi.BadRequestAPIError{Message: "This alert action is not attached to this alert source"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAlreadyDetachedError(tc.err); got != tc.want {
				t.Errorf("isAlreadyDetachedError() = %v, want %v", got, tc.want)
			}
		})
	}
}
