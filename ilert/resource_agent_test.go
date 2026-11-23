package ilert

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v4"
)

// storedAgent is an agent as the API returns it, with everything set that the
// resource has no argument for
func storedAgent() *ilertapi.Agent {
	threshold := 70
	return &ilertapi.Agent{
		ID:                     1,
		Name:                   "triage",
		AvatarURL:              "https://example.com/avatar.png",
		Status:                 "ENABLED",
		Teams:                  []ilertapi.TeamShort{{ID: 4, Name: "ops"}},
		McpServers:             []ilertapi.AgentMcpServer{{Name: "ilert-workforce-github-mcp", Active: true, Credentials: &ilertapi.AgentMcpServerCredentials{ID: 3}}},
		McpTools:               []ilertapi.AgentMcpTool{{Server: "ilert-workforce-github-mcp", Name: "get_file", Active: true}},
		AdditionalInstructions: "be brief",
		RcaPolicy:              &ilertapi.AgentRcaPolicy{Type: "CUSTOM", Condition: "alert.priority == 'HIGH'", ScoreThreshold: &threshold},
		AutonomyLevel:          "FULLY",
	}
}

// The API replaces an agent with what an update sends. What the configuration does not
// manage has to go back as it was read, or the update clears it.
func TestApplyAgentConfig_KeepsWhatTerraformDoesNotManage(t *testing.T) {
	d := testResourceDataForUpdate(t, resourceAgent(),
		map[string]any{"name": "triage"},
		map[string]any{"name": "triage-2", "status": "DISABLED"})

	agent := storedAgent()
	if err := applyAgentConfig(agent, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := storedAgent()
	want.Name = "triage-2"
	want.Status = "DISABLED"
	if !reflect.DeepEqual(agent, want) {
		t.Errorf("agent = %+v, want %+v", agent, want)
	}
}

// Undeclared MCP servers and tools go back exactly as the API returned them, with what
// the SDK types do not know: a server saved without its active flag must not be sent
// as inactive, and a credential reference without an id must not become id 0, which
// the API rejects.
func TestApplyAgentConfig_SendsUndeclaredMcpListsBackAsRead(t *testing.T) {
	const servers = `[{"name":"ilert-workforce-github-mcp","credentials":{"id":null},"futureField":1}]`
	agent := &ilertapi.Agent{}
	if err := json.Unmarshal([]byte(`{"id":1,"name":"triage","mcpServers":`+servers+`,"mcpTools":[]}`), agent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d := testResourceDataForUpdate(t, resourceAgent(),
		map[string]any{"name": "triage"},
		map[string]any{"name": "triage-2"})
	if err := applyAgentConfig(agent, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	encoded, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body["mcpServers"]) != servers {
		t.Errorf("mcpServers = %s, want them as read: %s", body["mcpServers"], servers)
	}
	if string(body["name"]) != `"triage-2"` {
		t.Errorf("name = %s, want the configured one", body["name"])
	}
}

// Like teams, the MCP servers and tools are managed once declared: removing every block
// clears them, which takes an empty list since the API treats null the same.
func TestApplyAgentConfig_RemovingAllMcpBlocksClearsThem(t *testing.T) {
	declared := map[string]any{
		"name":       "triage",
		"mcp_server": []any{map[string]any{"name": "ilert-workforce-github-mcp"}},
		"mcp_tool":   []any{map[string]any{"server": "ilert-workforce-github-mcp", "name": "get_file"}},
	}
	d := testResourceDataForUpdate(t, resourceAgent(), declared, map[string]any{"name": "triage"})

	agent := storedAgent()
	if err := applyAgentConfig(agent, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	encoded, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{`"mcpServers":[]`, `"mcpTools":[]`} {
		if !strings.Contains(string(encoded), want) {
			t.Errorf("payload = %s, want it to contain %s", encoded, want)
		}
	}
}

func TestApplyAgentConfig_BuildsServersAndTools(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAgent().Schema, map[string]any{
		"name": "triage",
		"mcp_server": []any{
			map[string]any{"name": "ilert-workforce-github-mcp", "credential_id": "3"},
			map[string]any{"name": "datadog", "display_name": "Datadog", "brand": "datadog", "active": false},
		},
		"mcp_tool": []any{
			map[string]any{
				"server": "ilert-workforce-github-mcp",
				"name":   "get_file",
				"input_overwrite": []any{
					map[string]any{"name": "repo", "values": []any{"iLert/ilert-go"}},
					map[string]any{"name": "limit", "type": "number", "values": []any{"5", "10"}},
				},
			},
		},
	})

	agent := &ilertapi.Agent{}
	if err := applyAgentConfig(agent, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantServers := []ilertapi.AgentMcpServer{
		{Name: "ilert-workforce-github-mcp", Active: true, Credentials: &ilertapi.AgentMcpServerCredentials{ID: 3}},
		{Name: "datadog", Active: false, DisplayName: "Datadog", Brand: "datadog"},
	}
	if !reflect.DeepEqual(agent.McpServers, wantServers) {
		t.Errorf("servers = %+v, want %+v", agent.McpServers, wantServers)
	}
	wantTools := []ilertapi.AgentMcpTool{{
		Server: "ilert-workforce-github-mcp",
		Name:   "get_file",
		Active: true,
		InputOverwrites: []ilertapi.AgentMcpToolInputOverwrite{
			{Name: "repo", Type: "string", Value: ilertapi.AgentMcpToolInputValues{"iLert/ilert-go"}},
			{Name: "limit", Type: "number", Value: ilertapi.AgentMcpToolInputValues{"5", "10"}},
		},
	}}
	if !reflect.DeepEqual(agent.McpTools, wantTools) {
		t.Errorf("tools = %+v, want %+v", agent.McpTools, wantTools)
	}
	if agent.Status != "ENABLED" {
		t.Errorf("status = %q, want the ENABLED default", agent.Status)
	}
}

// A configuration that never declared MCP servers or tools leaves those set in the web
// app alone, so they stay out of the state instead of showing up as a removal.
func TestTransformAgentResource_LeavesUndeclaredMcpBlocksOut(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAgent().Schema, map[string]any{"name": "triage"})
	d.SetId("1")

	if err := transformAgentResource(storedAgent(), d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if servers := d.Get("mcp_server").([]any); len(servers) != 0 {
		t.Errorf("mcp_server = %v, want none in the state", servers)
	}
	if tools := d.Get("mcp_tool").([]any); len(tools) != 0 {
		t.Errorf("mcp_tool = %v, want none in the state", tools)
	}
}

// Read without an id, as the IaC exporter does, the agent is taken whole. An input
// overwrite without values has no effect and cannot be configured, so it is left out.
func TestTransformAgentResource_ExportsMcpBlocks(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAgent().Schema, nil)

	agent := storedAgent()
	agent.McpTools[0].InputOverwrites = []ilertapi.AgentMcpToolInputOverwrite{
		{Name: "repo", Type: "string", Value: ilertapi.AgentMcpToolInputValues{"iLert/ilert-go"}},
		{Name: "path", Type: "string"},
		{Name: "ref", Type: "string", Value: ilertapi.AgentMcpToolInputValues{}},
	}
	if err := transformAgentResource(agent, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := d.Get("mcp_server.0.credential_id").(string); got != "3" {
		t.Errorf("credential_id = %q, want 3", got)
	}
	overwrites := d.Get("mcp_tool.0.input_overwrite").([]any)
	if len(overwrites) != 1 || overwrites[0].(map[string]any)["name"] != "repo" {
		t.Errorf("input_overwrite = %v, want only the repo overwrite", overwrites)
	}
}
