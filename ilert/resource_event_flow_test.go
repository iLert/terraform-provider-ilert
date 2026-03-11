package ilert

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ilertapi "github.com/iLert/ilert-go/v3"
)

func TestBuildEventFlow_BuildsTransformMetadata(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceEventFlow().Schema, map[string]any{
		"name": "test-event-flow",
		"root_node": []any{
			map[string]any{
				"node_type": "ROOT",
				"branches": []any{
					map[string]any{
						"branch_type": "ACCEPTED",
						"target": []any{
							map[string]any{
								"node_type": "TRANSFORM",
								"metadata": []any{
									map[string]any{
										"condition": "context.event != null",
										"rules": []any{
											map[string]any{
												"name":     "Rule 1",
												"target":   "context.event.summary",
												"operator": "SET",
												"value":    "hello",
												"default":  "fallback",
												"source":   "context.event.details",
												"mapping": map[string]any{
													"OPEN": "PENDING",
												},
												"properties": map[string]any{
													"severity": "SEV2",
												},
												"items": []any{
													map[string]any{
														"href": "https://example.com",
														"text": "example",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	eventFlow, err := buildEventFlow(d)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if eventFlow == nil || eventFlow.RootNode == nil {
		t.Fatal("expected root node to be set")
	}
	if len(eventFlow.RootNode.Branches) != 1 || eventFlow.RootNode.Branches[0].Target == nil {
		t.Fatal("expected accepted branch target to be built")
	}

	targetNode := eventFlow.RootNode.Branches[0].Target
	if targetNode.NodeType != ilertapi.EventFlowNodeType.Transform {
		t.Fatalf("expected target node_type TRANSFORM, got: %s", targetNode.NodeType)
	}

	md, ok := targetNode.Metadata.(*ilertapi.EventFlowNodeMetadata)
	if !ok || md == nil {
		t.Fatal("expected metadata to be *EventFlowNodeMetadata")
	}
	if md.Condition != "context.event != null" {
		t.Fatalf("expected condition to be mapped, got: %s", md.Condition)
	}
	if len(md.Rules) != 1 {
		t.Fatalf("expected one rule, got: %d", len(md.Rules))
	}

	rule := md.Rules[0]
	if rule.Operator != ilertapi.EventFlowNodeRuleOperator.Set {
		t.Fatalf("expected operator SET, got: %s", rule.Operator)
	}

	value, ok := rule.Value.(string)
	if !ok || value != "hello" {
		t.Fatalf("expected rule value to be string 'hello', got: %#v", rule.Value)
	}
	defaultValue, ok := rule.Default.(string)
	if !ok || defaultValue != "fallback" {
		t.Fatalf("expected rule default to be string 'fallback', got: %#v", rule.Default)
	}
	if rule.Mapping == nil || rule.Mapping["OPEN"] == nil || *rule.Mapping["OPEN"] != "PENDING" {
		t.Fatalf("expected mapping OPEN=>PENDING, got: %#v", rule.Mapping)
	}
	if rule.Properties == nil || rule.Properties["severity"] == nil || *rule.Properties["severity"] != "SEV2" {
		t.Fatalf("expected properties severity=>SEV2, got: %#v", rule.Properties)
	}
	if len(rule.Items) != 1 {
		t.Fatalf("expected one item, got: %d", len(rule.Items))
	}
}
