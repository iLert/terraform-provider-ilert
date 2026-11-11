package ilert

import "testing"

// The IaC exporter resolves provider resources by the API's entity type name, so a
// resource registered in the provider but missing from the registry maps is invisible to
// it. TELEMETRY_SOURCE is the entity type the API reports for telemetry sources.
func TestGetResourceInfo_TelemetrySource(t *testing.T) {
	info, err := GetResourceInfo("TELEMETRY_SOURCE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ResourceType != "ilert_telemetry_source" {
		t.Fatalf("ResourceType = %q, want ilert_telemetry_source", info.ResourceType)
	}
	if info.NewEntity() == nil {
		t.Fatal("NewEntity returned nil")
	}
	if info.Transformer == nil {
		t.Fatal("Transformer is nil")
	}
	if info.Schema == nil {
		t.Fatal("Schema is nil")
	}
}

// AGENT and CUSTOMER_CREDENTIAL are the entity types the API reports for agents and
// credentials.
func TestGetResourceInfo_AgentAndCredential(t *testing.T) {
	for entityType, resourceType := range map[string]string{
		"AGENT":               "ilert_agent",
		"CUSTOMER_CREDENTIAL": "ilert_credential",
	} {
		info, err := GetResourceInfo(entityType)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", entityType, err)
		}
		if info.ResourceType != resourceType {
			t.Fatalf("%s: ResourceType = %q, want %s", entityType, info.ResourceType, resourceType)
		}
		if info.NewEntity() == nil || info.Transformer == nil || info.Schema == nil {
			t.Fatalf("%s: registry entry is incomplete", entityType)
		}
	}
}
