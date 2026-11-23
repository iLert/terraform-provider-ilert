package ilert

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ilertapi "github.com/iLert/ilert-go/v4"
)

// The API replaces the config with what an update sends, so the values a template
// wrote into it have to go back, while a key the configuration removed goes away.
func TestMergeActionConfig(t *testing.T) {
	read := json.RawMessage(`{"url":"https://old.example.com","method":"POST","webhookId":"written by ilert"}`)
	for _, tc := range []struct {
		name, old, configured, want string
	}{
		{"unchanged", `{"url":"https://old.example.com","method":"POST"}`, `{"url":"https://old.example.com","method":"POST"}`,
			`{"method":"POST","url":"https://old.example.com","webhookId":"written by ilert"}`},
		{"changed value", `{"url":"https://old.example.com","method":"POST"}`, `{"url":"https://new.example.com","method":"POST"}`,
			`{"method":"POST","url":"https://new.example.com","webhookId":"written by ilert"}`},
		{"key removed", `{"url":"https://old.example.com","method":"POST"}`, `{"url":"https://old.example.com"}`,
			`{"url":"https://old.example.com","webhookId":"written by ilert"}`},
		{"key added", `{"url":"https://old.example.com"}`, `{"url":"https://old.example.com","method":"PUT"}`,
			`{"method":"PUT","url":"https://old.example.com","webhookId":"written by ilert"}`},
		{"config removed", `{"url":"https://old.example.com","method":"POST"}`, ``,
			`{"webhookId":"written by ilert"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := mergeActionConfig(read, tc.old, tc.configured)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("config = %s, want %s", got, tc.want)
			}
		})
	}

	// nothing left is sent as no config, which ilert-go sends as {} to clear it
	got, err := mergeActionConfig(json.RawMessage(`{"url":"x"}`), `{"url":"x"}`, ``)
	if err != nil || got != nil {
		t.Errorf("config = %s, %v, want none", got, err)
	}
}

// The state holds only the keys the configuration sets, so the values a template
// writes into the config do not show as a change. After an import none, and all for
// the IaC exporter.
func TestFlattenActionConfig(t *testing.T) {
	read := json.RawMessage(`{"url":"https://example.com","method":"POST","webhookId":"written by ilert"}`)
	d := resourceAction().TestResourceData()
	d.SetId("1")
	d.Set("config", `{"method":"GET","url":"https://example.com"}`)

	got, err := flattenActionConfig(read, d)
	if err != nil || got != `{"method":"POST","url":"https://example.com"}` {
		t.Errorf("managed keys: config = %s, %v, want the configured keys as read", got, err)
	}

	exporter := resourceAction().TestResourceData()
	got, err = flattenActionConfig(read, exporter)
	if err != nil || got != `{"method":"POST","url":"https://example.com","webhookId":"written by ilert"}` {
		t.Errorf("exporter: config = %s, %v, want every key", got, err)
	}

	imported := resourceAction().TestResourceData()
	imported.SetId("1")
	got, err = flattenActionConfig(read, imported)
	if err != nil || got != "" {
		t.Errorf("imported, or no config managed: config = %s, %v, want none", got, err)
	}
}

func TestSuppressEquivalentJSON(t *testing.T) {
	for _, tc := range []struct {
		old, new string
		want     bool
	}{
		{`{"a":1,"b":[1,2]}`, `{ "b": [1, 2], "a": 1 }`, true},
		{`{"a":1}`, `{"a":2}`, false},
		{`[{"id":"x"}]`, `[{"id":"x"}]`, true},
		{`[1,2]`, `[2,1]`, false},
		{``, `{}`, true},
		{``, `[]`, true},
		{``, `{"a":1}`, false},
		{`{"a":"1"}`, `{"a":1}`, false},
	} {
		if got := suppressEquivalentJSON("config", tc.old, tc.new, nil); got != tc.want {
			t.Errorf("suppress(%s, %s) = %t, want %t", tc.old, tc.new, got, tc.want)
		}
	}
}

func TestBuildAction(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAction().Schema, map[string]any{
		"name":             "page ops",
		"template_key":     "webhook_alert",
		"template_version": 2,
		"entity_kind":      "ALERT",
		"trigger_events":   []any{"alert.created"},
		"trigger_event_delay": []any{
			map[string]any{"event": "alert.not_resolved", "delay_sec": 600},
		},
		"steps":         `[{"id":"post"}]`,
		"credential_id": "42",
		"team":          []any{map[string]any{"id": 7}},
		"labels":        map[string]any{"env": "prod"},
	})
	action, err := buildAction(d, json.RawMessage(`{"url":"https://example.com"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Name != "page ops" || action.TemplateKey != "webhook_alert" || action.EntityKind != "ALERT" ||
		action.DispatchMode != "AUTOMATIC" || action.Enabled == nil || !*action.Enabled {
		t.Errorf("action = %+v, want the configured fields and the defaults", action)
	}
	if action.TemplateVersion == nil || *action.TemplateVersion != 2 || action.CredentialID == nil || *action.CredentialID != 42 {
		t.Errorf("template version and credential = %v %v, want 2 and 42", action.TemplateVersion, action.CredentialID)
	}
	if len(action.TriggerEvents) != 1 || len(action.TriggerEventDelays) != 1 || action.TriggerEventDelays[0].DelaySec != 600 {
		t.Errorf("triggers = %v %v, want alert.created and a delay of 600", action.TriggerEvents, action.TriggerEventDelays)
	}
	if string(action.Steps) != `[{"id":"post"}]` || action.FieldMapping != nil || action.Inbound != nil {
		t.Errorf("definition = %s %s %s, want only the steps", action.Steps, action.FieldMapping, action.Inbound)
	}
	if len(action.Teams) != 1 || action.Teams[0].ID != 7 || action.Labels["env"] != "prod" {
		t.Errorf("teams and labels = %v %v, want team 7 and env=prod", action.Teams, action.Labels)
	}

	// left out: no template version, credential, teams or labels, which ilert-go sends
	// empty to clear them
	d = schema.TestResourceDataRaw(t, resourceAction().Schema, map[string]any{
		"name": "page ops", "template_key": "webhook_alert", "entity_kind": "ALERT",
	})
	action, err = buildAction(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.TemplateVersion != nil || action.CredentialID != nil || action.Teams != nil || action.Labels != nil || action.Config != nil {
		t.Errorf("action = %+v, want nothing set that was left out", action)
	}
}

// The plan rejects what the API would: an action without a template or steps, and
// triggers of the other entity kind or with a delay the API does not take.
func TestValidateActionDiff(t *testing.T) {
	client := ilertapi.NewClient(ilertapi.WithAPIToken("test-token"), ilertapi.WithEarlyAccess("actions"))
	plan := func(config map[string]any) error {
		_, err := resourceAction().Diff(context.Background(), nil, terraform.NewResourceConfigRaw(config), client)
		return err
	}
	base := func(extra map[string]any) map[string]any {
		config := map[string]any{"name": "page ops", "template_key": "webhook_alert", "entity_kind": "ALERT"}
		for key, value := range extra {
			config[key] = value
		}
		return config
	}

	if err := plan(base(map[string]any{
		"trigger_events":      []any{"alert.created", "alert.not_resolved"},
		"trigger_event_delay": []any{map[string]any{"event": "alert.not_resolved", "delay_sec": 600}},
	})); err != nil {
		t.Errorf("valid action: err = %v", err)
	}
	if err := plan(map[string]any{"name": "custom", "entity_kind": "ALERT", "steps": `[{"id":"post"}]`}); err != nil {
		t.Errorf("action with its own steps: err = %v", err)
	}

	for name, tc := range map[string]struct {
		config map[string]any
		want   string
	}{
		"no template or steps": {map[string]any{"name": "x", "entity_kind": "ALERT"}, "needs a template_key, or its own steps"},
		"incident event":       {base(map[string]any{"trigger_events": []any{"incident.opened"}}), "is not an event of an ALERT action"},
		"incident delay": {base(map[string]any{"trigger_event_delay": []any{map[string]any{"event": "incident.not_resolved", "delay_sec": 600}}}),
			"is not an event of an ALERT action"},
		"short delay": {base(map[string]any{"trigger_event_delay": []any{map[string]any{"event": "alert.not_resolved", "delay_sec": 30}}}),
			"takes 60 to 172800 seconds"},
		"long escalation delay": {base(map[string]any{"trigger_event_delay": []any{map[string]any{"event": "alert.escalation_ended", "delay_sec": 9000}}}),
			"takes 30 to 7200 seconds"},
		"delay twice": {base(map[string]any{"trigger_event_delay": []any{
			map[string]any{"event": "alert.not_resolved", "delay_sec": 600},
			map[string]any{"event": "alert.not_resolved", "delay_sec": 900},
		}}), "is delayed twice"},
		"not_resolved without a delay": {base(map[string]any{"trigger_events": []any{"alert.created", "alert.not_resolved"}}),
			"needs a trigger_event_delay"},
	} {
		if err := plan(tc.config); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to say %q", name, err, tc.want)
		}
	}

	for name, config := range map[string]map[string]any{
		"escalation delay below 30": base(map[string]any{"trigger_event_delay": []any{map[string]any{"event": "alert.escalation_ended", "delay_sec": 10}}}),
		"config not an object":      base(map[string]any{"config": `["a"]`}),
		"steps not an array":        base(map[string]any{"steps": `{"id":"post"}`}),
		"inbound not an object":     base(map[string]any{"inbound": `[]`}),
	} {
		if diags := resourceAction().Validate(terraform.NewResourceConfigRaw(config)); !diags.HasError() {
			t.Errorf("%s: accepted, want a validation error", name)
		}
	}
}

func TestActionsUnavailableAndMirrors(t *testing.T) {
	unavailable := &ilertapi.NotFoundAPIError{Status: 404, Message: "Global actions are not available for this account"}
	err := actionsUnavailable(unavailable)
	var notFound *ilertapi.NotFoundAPIError
	if errors.As(err, &notFound) || !strings.Contains(err.Error(), "not available for this ilert account") {
		t.Errorf("err = %v, want an error that is not a not found, so the action stays in the state", err)
	}
	missing := &ilertapi.NotFoundAPIError{Status: 404, Message: "Could not find action with id 9"}
	if err := actionsUnavailable(missing); err != error(missing) {
		t.Errorf("err = %v, want the not found of a missing action as it is", err)
	}

	if err := assertNativeAction(&ilertapi.Action{ID: 9, Origin: ilertapi.ActionOrigin.Native}); err != nil {
		t.Errorf("native action: err = %v", err)
	}
	err = assertNativeAction(&ilertapi.Action{ID: 9, Origin: ilertapi.ActionOrigin.LegacyRef, LegacyAlertActionID: 5})
	if err == nil || !strings.Contains(err.Error(), "ilert_alert_action") {
		t.Errorf("mirror of an alert action: err = %v, want it refused", err)
	}
}

func TestActionCredentialIDIsValidatedAtPlan(t *testing.T) {
	validate := resourceAction().Schema["credential_id"].ValidateFunc
	if _, errs := validate("189530", "credential_id"); len(errs) != 0 {
		t.Errorf("an id was refused: %v", errs)
	}
	for _, bad := range []string{"abc", "12a", "-1", " 1"} {
		if _, errs := validate(bad, "credential_id"); len(errs) == 0 {
			t.Errorf("credential_id %q was accepted, want it refused at plan time", bad)
		}
	}
}
