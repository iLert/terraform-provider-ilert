package ilert

import (
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v3"
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
