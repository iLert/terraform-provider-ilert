package ilert

import (
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
