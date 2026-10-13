package ilert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v4"
)

// The API can leave stored credentials out of its response. Every credential of the
// connector schema has to survive a refresh that returns none, or the configured value
// shows as a change on every plan.
func TestTransformConnectorResource_KeepsCredentialMissingFromResponse(t *testing.T) {
	cases := []struct {
		connectorType string
		block         string
		key           string
		config        map[string]any
		params        ilertapi.ConnectorOutputParams
	}{
		{ilertapi.ConnectorTypes.Jira, "jira", "password",
			map[string]any{"url": "https://jira.example.com", "email": "bot@example.com", "password": "configured"},
			ilertapi.ConnectorOutputParams{URL: "https://jira.example.com", Email: "bot@example.com"}},
		{ilertapi.ConnectorTypes.ServiceNow, "servicenow", "password",
			map[string]any{"url": "https://example.service-now.com", "username": "bot", "password": "configured"},
			ilertapi.ConnectorOutputParams{URL: "https://example.service-now.com", Username: "bot"}},
		{ilertapi.ConnectorTypes.Zendesk, "zendesk", "api_key",
			map[string]any{"url": "https://example.zendesk.com", "email": "bot@example.com", "api_key": "configured"},
			ilertapi.ConnectorOutputParams{URL: "https://example.zendesk.com", Email: "bot@example.com"}},
		{ilertapi.ConnectorTypes.Github, "github", "api_key",
			map[string]any{"api_key": "configured"},
			ilertapi.ConnectorOutputParams{}},
		{ilertapi.ConnectorTypes.Topdesk, "topdesk", "password",
			map[string]any{"url": "https://example.topdesk.net", "username": "bot", "password": "configured"},
			ilertapi.ConnectorOutputParams{URL: "https://example.topdesk.net", Username: "bot"}},
		{ilertapi.ConnectorTypes.Autotask, "autotask", "password",
			map[string]any{"url": "https://webservices.autotask.net", "email": "bot@example.com", "password": "configured"},
			ilertapi.ConnectorOutputParams{URL: "https://webservices.autotask.net", Email: "bot@example.com"}},
		{ilertapi.ConnectorTypes.Zammad, "zammad", "api_key",
			map[string]any{"url": "https://zammad.example.com", "api_key": "configured"},
			ilertapi.ConnectorOutputParams{URL: "https://zammad.example.com"}},
		{ilertapi.ConnectorTypes.DingTalk, "dingtalk", "secret",
			map[string]any{"url": "https://oapi.dingtalk.com/robot/send", "secret": "configured"},
			ilertapi.ConnectorOutputParams{URL: "https://oapi.dingtalk.com/robot/send"}},
	}

	for _, tc := range cases {
		t.Run(tc.block, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceConnector().Schema, map[string]any{
				"name":   "example",
				"type":   tc.connectorType,
				tc.block: []any{tc.config},
			})
			d.SetId("1")

			connector := &ilertapi.ConnectorOutput{ID: "1", Name: "example", Type: tc.connectorType, Params: tc.params}
			if err := transformConnectorResource(connector, d); err != nil {
				t.Fatalf("unexpected error transforming connector: %v", err)
			}

			if got := d.Get(tc.block + ".0." + tc.key); got != "configured" {
				t.Fatalf("expected the configured %s to be kept, got %q", tc.key, got)
			}
		})
	}
}

// A credential the API does return still replaces the one in the state, so a credential
// changed in ilert keeps showing as a change in the plan.
func TestTransformConnectorResource_TakesCredentialFromResponse(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceConnector().Schema, map[string]any{
		"name": "example",
		"type": ilertapi.ConnectorTypes.Jira,
		"jira": []any{
			map[string]any{"url": "https://jira.example.com", "email": "bot@example.com", "password": "configured"},
		},
	})
	d.SetId("1")

	connector := &ilertapi.ConnectorOutput{ID: "1", Name: "example", Type: ilertapi.ConnectorTypes.Jira,
		Params: ilertapi.ConnectorOutputParams{URL: "https://jira.example.com", Email: "bot@example.com", Password: "changed in ilert"}}
	if err := transformConnectorResource(connector, d); err != nil {
		t.Fatalf("unexpected error transforming connector: %v", err)
	}

	if got := d.Get("jira.0.password"); got != "changed in ilert" {
		t.Fatalf("expected the password from the response, got %q", got)
	}
}

// Without a credential in the state, as on import or in an export, there is nothing to
// keep, and a credential the API did not return stays empty.
func TestTransformConnectorResource_LeavesMissingCredentialEmptyWithoutState(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceConnector().Schema, map[string]any{})
	d.SetId("1")

	connector := &ilertapi.ConnectorOutput{ID: "1", Name: "example", Type: ilertapi.ConnectorTypes.Jira,
		Params: ilertapi.ConnectorOutputParams{URL: "https://jira.example.com", Email: "bot@example.com"}}
	if err := transformConnectorResource(connector, d); err != nil {
		t.Fatalf("unexpected error transforming connector: %v", err)
	}

	if got := d.Get("jira.0.password"); got != "" {
		t.Fatalf("expected an empty password, got %q", got)
	}
}
