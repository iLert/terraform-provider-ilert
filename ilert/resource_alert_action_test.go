package ilert

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v4"
)

func TestTransformAlertActionResource_DoesNotPanicOnEmptyAlertSourceIDs(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertAction().Schema, map[string]any{
		"name": "test-alert-action",
		"connector": []any{
			map[string]any{
				"id":   "1",
				"type": "SLACK",
			},
		},
		"trigger_mode": "ALL",
		"alert_source": []any{
			map[string]any{
				"id": "123",
			},
		},
	})
	d.SetId("1")

	alertAction := &ilertapi.AlertActionOutput{
		Name:           "test-alert-action",
		ConnectorType:  "SLACK",
		AlertSourceIDs: []int64{},
	}

	if err := transformAlertActionResource(alertAction, d); err != nil {
		t.Fatalf("unexpected error transforming alert action: %v", err)
	}

	alertSources := d.Get("alert_source").([]any)
	if len(alertSources) != 0 {
		t.Fatalf("expected 0 alert sources in state, got %d", len(alertSources))
	}
}

func TestFlattenAlertActionAlertSourcesListSorted_UsesConfigOrder(t *testing.T) {
	serverAlertSources := []ilertapi.AlertSource{
		{ID: 100},
		{ID: 200},
	}

	configAlertSources := []any{
		map[string]any{"id": "200"},
		map[string]any{"id": "100"},
	}

	got, err := flattenAlertActionAlertSourcesListSorted(serverAlertSources, configAlertSources)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	wantOrder := []string{"200", "100"}
	gotOrder := alertSourceIDOrder(got)
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("expected source order %v, got %v", wantOrder, gotOrder)
	}
}

func TestFlattenAlertActionAlertSourcesListSorted_AppendsUnmatchedServerAlertSource(t *testing.T) {
	serverAlertSources := []ilertapi.AlertSource{
		{ID: 100},
		{ID: 200},
		{ID: 300},
	}

	configAlertSources := []any{
		map[string]any{"id": "200"},
	}

	got, err := flattenAlertActionAlertSourcesListSorted(serverAlertSources, configAlertSources)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	gotOrder := alertSourceIDOrder(got)
	if len(gotOrder) != 3 {
		t.Fatalf("expected 3 alert sources, got %d (%v)", len(gotOrder), gotOrder)
	}
	if gotOrder[0] != "200" {
		t.Fatalf("expected configured source first, got %v", gotOrder)
	}
	if !slices.Contains(gotOrder, "100") || !slices.Contains(gotOrder, "300") {
		t.Fatalf("expected unmatched server sources to be preserved, got %v", gotOrder)
	}
}

func TestBuildAlertAction_WebhookHeaders(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertAction().Schema, map[string]any{
		"name": "test-alert-action",
		"connector": []any{
			map[string]any{
				"type": ilertapi.ConnectorTypes.Webhook,
			},
		},
		"webhook": []any{
			map[string]any{
				"url":           "https://example.com/webhook",
				"body_template": "body-template",
				"headers": []any{
					map[string]any{
						"key":   "Authorization",
						"value": "Bearer test-token",
					},
				},
			},
		},
	})

	alertAction, err := buildAlertAction(d)
	if err != nil {
		t.Fatalf("unexpected error building alert action: %v", err)
	}

	params, ok := alertAction.Params.(*ilertapi.AlertActionParamsWebhook)
	if !ok {
		t.Fatalf("expected webhook params type, got: %#v", alertAction.Params)
	}

	if len(params.Headers) != 1 {
		t.Fatalf("expected 1 webhook header, got %d", len(params.Headers))
	}

	if params.Headers[0].Key != "Authorization" {
		t.Fatalf("expected header key Authorization, got %s", params.Headers[0].Key)
	}

	if params.Headers[0].Value != "Bearer test-token" {
		t.Fatalf("expected header value Bearer test-token, got %s", params.Headers[0].Value)
	}
}

func TestFlattenAlertActionWebhookHeadersList(t *testing.T) {
	webhookHeaders := []ilertapi.AlertActionParamsWebhookHeader{
		{
			Key:   "X-Header-Key",
			Value: "X-Header-Value",
		},
	}

	got := flattenAlertActionWebhookHeadersList(webhookHeaders)

	want := []any{
		map[string]any{
			"key":   "X-Header-Key",
			"value": "X-Header-Value",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected webhook headers to be %#v, got %#v", want, got)
	}
}

func alertSourceIDOrder(alertSources []any) []string {
	ids := make([]string, 0, len(alertSources))
	for _, alertSource := range alertSources {
		alertSourceMap, ok := alertSource.(map[string]any)
		if !ok {
			continue
		}
		sourceID, ok := alertSourceMap["id"].(string)
		if ok {
			ids = append(ids, sourceID)
		}
	}
	return ids
}

func TestTransformAlertActionResource_AutotaskParams(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertAction().Schema, map[string]any{
		"name": "test-alert-action",
		"connector": []any{
			map[string]any{
				"id":   "1",
				"type": "autotask",
			},
		},
		"trigger_mode": "ALL",
		"alert_source": []any{
			map[string]any{
				"id": "123",
			},
		},
	})
	d.SetId("1")

	alertAction := &ilertapi.AlertActionOutput{
		Name:           "test-alert-action",
		ConnectorType:  ilertapi.ConnectorTypes.Autotask,
		AlertSourceIDs: []int64{},
		Params: &ilertapi.AlertActionOutputParams{
			CompanyID: "12345",
			QueueID:   8,
			NoteType:  "1",
		},
	}

	// company_id (string) passes through, queue_id (int64) surfaces as an int.
	if err := transformAlertActionResource(alertAction, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	at := d.Get("autotask").([]any)[0].(map[string]any)
	if at["company_id"] != "12345" {
		t.Errorf("company_id = %v, want \"12345\"", at["company_id"])
	}
	if at["queue_id"] != 8 {
		t.Errorf("queue_id = %v, want 8", at["queue_id"])
	}
	if at["note_type"] != "1" {
		t.Errorf("note_type = %v, want \"1\"", at["note_type"])
	}

	// A non-numeric company_id must round-trip untouched, and an unset one
	// must stay an empty string.
	alertAction.Params.CompanyID = "ACME-42"
	if err := transformAlertActionResource(alertAction, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	at = d.Get("autotask").([]any)[0].(map[string]any)
	if at["company_id"] != "ACME-42" {
		t.Errorf("company_id = %v, want \"ACME-42\"", at["company_id"])
	}

	alertAction.Params.CompanyID = ""
	if err := transformAlertActionResource(alertAction, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	at = d.Get("autotask").([]any)[0].(map[string]any)
	if at["company_id"] != "" {
		t.Errorf("company_id = %v, want empty string when unset", at["company_id"])
	}
}

func TestBuildAlertAction_Reroute(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertAction().Schema, map[string]any{
		"name": "test-alert-action",
		"connector": []any{
			map[string]any{
				"type": ilertapi.ConnectorTypes.Reroute,
			},
		},
		"trigger_types":          []any{ilertapi.AlertActionTriggerTypes.AlertNotResolved},
		"not_resolved_delay_sec": 600,
		"conditions":             `alert.summary contains_any ["PROD"]`,
		"reroute": []any{
			map[string]any{
				"escalation_policy": []any{
					map[string]any{
						"id": 2262584,
					},
				},
			},
		},
	})

	alertAction, err := buildAlertAction(d)
	if err != nil {
		t.Fatalf("unexpected error building alert action: %v", err)
	}

	body, err := json.Marshal(alertAction)
	if err != nil {
		t.Fatalf("unexpected error marshalling alert action: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unexpected error unmarshalling alert action: %v", err)
	}

	// The API rejects a connectorId for reroute and only accepts the AUTOMATIC trigger mode.
	if _, ok := got["connectorId"]; ok {
		t.Errorf("connectorId = %v, want it left out", got["connectorId"])
	}
	if got["connectorType"] != "reroute" {
		t.Errorf("connectorType = %v, want \"reroute\"", got["connectorType"])
	}
	if got["triggerMode"] != "AUTOMATIC" {
		t.Errorf("triggerMode = %v, want \"AUTOMATIC\"", got["triggerMode"])
	}
	wantParams := map[string]any{"escalationPolicyId": float64(2262584)}
	if !reflect.DeepEqual(got["params"], wantParams) {
		t.Errorf("params = %v, want %v", got["params"], wantParams)
	}
	if !reflect.DeepEqual(got["triggerTypes"], []any{"v-alert-not-resolved"}) {
		t.Errorf("triggerTypes = %v, want [v-alert-not-resolved]", got["triggerTypes"])
	}
	if got["notResolvedDelaySec"] != float64(600) {
		t.Errorf("notResolvedDelaySec = %v, want 600", got["notResolvedDelaySec"])
	}
}

func TestTransformAlertActionResource_RerouteParams(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAlertAction().Schema, map[string]any{
		"name": "test-alert-action",
		"connector": []any{
			map[string]any{
				"type": ilertapi.ConnectorTypes.Reroute,
			},
		},
	})
	d.SetId("1")

	alertAction := &ilertapi.AlertActionOutput{
		Name:                "test-alert-action",
		ConnectorType:       ilertapi.ConnectorTypes.Reroute,
		TriggerMode:         ilertapi.AlertActionTriggerModes.Automatic,
		TriggerTypes:        []string{ilertapi.AlertActionTriggerTypes.AlertNotResolved},
		NotResolvedDelaySec: 600,
		Params: &ilertapi.AlertActionOutputParams{
			EscalationPolicyID: 2262584,
		},
	}

	if err := transformAlertActionResource(alertAction, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := d.Get("reroute.0.escalation_policy.0.id"); got != 2262584 {
		t.Errorf("reroute escalation_policy id = %v, want 2262584", got)
	}
	if got := d.Get("connector.0.id"); got != "" {
		t.Errorf("connector id = %v, want empty for reroute", got)
	}
	if got := d.Get("trigger_types"); !reflect.DeepEqual(got, []any{ilertapi.AlertActionTriggerTypes.AlertNotResolved}) {
		t.Errorf("trigger_types = %v, want [v-alert-not-resolved]", got)
	}
	if got := d.Get("not_resolved_delay_sec"); got != 600 {
		t.Errorf("not_resolved_delay_sec = %v, want 600", got)
	}
}
