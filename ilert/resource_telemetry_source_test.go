package ilert

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v3"
)

// Labels follow the same contract as the team blocks: removing every label has to reach
// the API as an explicit empty object, or the labels stay assigned on the server while
// Terraform reports the apply as successful.
func TestBuildTelemetrySource_RemovingAllLabelsSendsEmptyObject(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceTelemetrySource(),
		map[string]any{
			"name":   "test-telemetry-source",
			"type":   "OTEL",
			"labels": map[string]any{"env": "production"},
		},
		map[string]any{
			"name": "test-telemetry-source",
			"type": "OTEL",
		},
	)

	source, err := buildTelemetrySource(d)
	if err != nil {
		t.Fatalf("unexpected error building telemetry source: %v", err)
	}
	if source.Labels == nil {
		t.Fatalf("expected labels to be set to an empty object, got no labels field")
	}
	if len(*source.Labels) != 0 {
		t.Fatalf("expected labels to be empty, got %v", *source.Labels)
	}

	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("unexpected error marshalling payload: %v", err)
	}
	if !strings.Contains(string(encoded), `"labels":{}`) {
		t.Fatalf("expected payload to contain \"labels\":{}, got %s", encoded)
	}
}

// A configuration that never declared labels is not a statement that the telemetry
// source has none, so an unrelated update must leave labels set elsewhere alone.
func TestBuildTelemetrySource_KeepsLabelsWhenNeverDeclared(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceTelemetrySource(),
		map[string]any{
			"name": "test-telemetry-source",
			"type": "OTEL",
		},
		map[string]any{
			"name": "test-telemetry-source-renamed",
			"type": "OTEL",
		},
	)

	source, err := buildTelemetrySource(d)
	if err != nil {
		t.Fatalf("unexpected error building telemetry source: %v", err)
	}
	if source.Labels != nil {
		t.Fatalf("expected no labels field, got %v", *source.Labels)
	}

	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("unexpected error marshalling payload: %v", err)
	}
	if strings.Contains(string(encoded), `"labels"`) {
		t.Fatalf("expected payload to omit labels entirely, got %s", encoded)
	}
}

// The API omits the integration key for users without update permission on the
// telemetry source. Reading as such a user must not wipe the key already in state.
func TestTransformTelemetrySource_KeepsIntegrationKeyWhenNotReturned(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTelemetrySource().Schema, map[string]any{
		"name": "test-telemetry-source",
		"type": "OTEL",
	})
	d.SetId("1")
	d.Set("integration_key", "the-key")

	if err := transformTelemetrySourceResource(&ilertapi.TelemetrySource{
		Name: "test-telemetry-source",
		Type: "OTEL",
	}, d); err != nil {
		t.Fatalf("unexpected error transforming telemetry source: %v", err)
	}

	if got := d.Get("integration_key").(string); got != "the-key" {
		t.Fatalf("integration_key = %q, want the key already in state to survive a read without one", got)
	}
}

// A key the API does return replaces what state holds, so a rotation performed outside
// Terraform converges on the next refresh.
func TestTransformTelemetrySource_TakesReturnedIntegrationKey(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTelemetrySource().Schema, map[string]any{
		"name": "test-telemetry-source",
		"type": "OTEL",
	})
	d.SetId("1")
	d.Set("integration_key", "the-old-key")

	if err := transformTelemetrySourceResource(&ilertapi.TelemetrySource{
		Name:           "test-telemetry-source",
		Type:           "OTEL",
		IntegrationKey: "the-rotated-key",
	}, d); err != nil {
		t.Fatalf("unexpected error transforming telemetry source: %v", err)
	}

	if got := d.Get("integration_key").(string); got != "the-rotated-key" {
		t.Fatalf("integration_key = %q, want the rotated key from the API", got)
	}
}

// Labels the API returns reach the state of a configuration that declares them, so drift
// shows up as a normal diff.
func TestTransformTelemetrySource_SetsDeclaredLabels(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTelemetrySource().Schema, map[string]any{
		"name":   "test-telemetry-source",
		"type":   "OTEL",
		"labels": map[string]any{"env": "staging"},
	})
	d.SetId("1")

	labels := map[string]string{"env": "production", "region": "eu"}
	if err := transformTelemetrySourceResource(&ilertapi.TelemetrySource{
		Name:   "test-telemetry-source",
		Type:   "OTEL",
		Labels: &labels,
	}, d); err != nil {
		t.Fatalf("unexpected error transforming telemetry source: %v", err)
	}

	got := d.Get("labels").(map[string]any)
	if got["env"] != "production" || got["region"] != "eu" {
		t.Fatalf("labels = %v, want the labels from the API", got)
	}
}

// A configuration that never declared labels leaves them unmanaged: the API leaves them
// alone when the field is absent from a write, so pulling them into state would make the
// next plan propose their removal and take over labels set in the web app.
func TestTransformTelemetrySource_LeavesUndeclaredLabelsUnmanaged(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTelemetrySource().Schema, map[string]any{
		"name": "test-telemetry-source",
		"type": "OTEL",
	})
	d.SetId("1")

	labels := map[string]string{"owner": "webapp-team"}
	if err := transformTelemetrySourceResource(&ilertapi.TelemetrySource{
		Name:   "test-telemetry-source",
		Type:   "OTEL",
		Labels: &labels,
	}, d); err != nil {
		t.Fatalf("unexpected error transforming telemetry source: %v", err)
	}

	if got := d.Get("labels").(map[string]any); len(got) != 0 {
		t.Fatalf("labels = %v, want them left out of state so they stay unmanaged", got)
	}
}

// The API stamps ilert.com/discovered-by onto every telemetry source, derived from its
// name. Verified against the API on 09.09.2026. Keeping it in state would make every plan
// propose deleting a label the server writes straight back, and renaming the source would
// change the label's value on its own, so it is stripped on read.
func TestTransformTelemetrySource_StripsServerStampedLabel(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTelemetrySource().Schema, map[string]any{
		"name":   "test-telemetry-source",
		"type":   "OTEL",
		"labels": map[string]any{"env": "production"},
	})
	d.SetId("1")

	labels := map[string]string{
		"env":                            "production",
		telemetrySourceDiscoveredByLabel: "test-telemetry-source",
	}
	if err := transformTelemetrySourceResource(&ilertapi.TelemetrySource{
		Name:   "test-telemetry-source",
		Type:   "OTEL",
		Labels: &labels,
	}, d); err != nil {
		t.Fatalf("unexpected error transforming telemetry source: %v", err)
	}

	got := d.Get("labels").(map[string]any)
	if _, present := got[telemetrySourceDiscoveredByLabel]; present {
		t.Fatalf("labels = %v, want the server-stamped label stripped", got)
	}
	if got["env"] != "production" {
		t.Fatalf("labels = %v, want the user's own label kept", got)
	}
}

// The same label is kept out of the payload: the API documents it as read-only derived
// metadata that clients should omit when sending labels back.
func TestBuildTelemetrySource_OmitsServerStampedLabel(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTelemetrySource().Schema, map[string]any{
		"name": "test-telemetry-source",
		"type": "OTEL",
		"labels": map[string]any{
			"env":                            "production",
			telemetrySourceDiscoveredByLabel: "somebody-typed-this",
		},
	})

	source, err := buildTelemetrySource(d)
	if err != nil {
		t.Fatalf("unexpected error building telemetry source: %v", err)
	}
	if source.Labels == nil {
		t.Fatalf("Labels = nil, want the declared labels")
	}
	if _, present := (*source.Labels)[telemetrySourceDiscoveredByLabel]; present {
		t.Fatalf("Labels = %v, want the reserved label omitted from the payload", *source.Labels)
	}
	if (*source.Labels)["env"] != "production" {
		t.Fatalf("Labels = %v, want the user's own label sent", *source.Labels)
	}
}
