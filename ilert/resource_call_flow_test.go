package ilert

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v4"
)

func TestBuildCallFlow_RequiresCallStyleForRouteCallNode(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceCallFlow().Schema, map[string]any{
		"name":     "test-call-flow",
		"language": "en",
		"root_node": []any{
			map[string]any{
				"node_type": "ROOT",
				"branches": []any{
					map[string]any{
						"branch_type": "ANSWERED",
						"target": []any{
							map[string]any{
								"node_type": "ROUTE_CALL",
								"metadata": []any{
									map[string]any{
										"targets": []any{
											map[string]any{
												"target": "1",
												"type":   "USER",
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

	_, err := buildCallFlow(d)
	if err == nil {
		t.Fatal("expected error for ROUTE_CALL node without metadata.call_style")
	}
	if !strings.Contains(err.Error(), "requires 'call_style'") {
		t.Fatalf("expected metadata.call_style validation error, got: %v", err)
	}
}

func TestBuildCallFlow_CreateAlertAcceptAlertOnAnswer(t *testing.T) {
	build := func(accept bool) *ilert.CallFlowNodeMetadata {
		d := schema.TestResourceDataRaw(t, resourceCallFlow().Schema, map[string]any{
			"name":     "test-call-flow",
			"language": "en",
			"root_node": []any{
				map[string]any{
					"node_type": "ROOT",
					"branches": []any{
						map[string]any{
							"branch_type": "ANSWERED",
							"target": []any{
								map[string]any{
									"node_type": "CREATE_ALERT",
									"metadata": []any{
										map[string]any{
											"alert_source_id":        1,
											"accept_alert_on_answer": accept,
										},
									},
								},
							},
						},
					},
				},
			},
		})

		cf, err := buildCallFlow(d)
		if err != nil {
			t.Fatalf("unexpected error building call flow: %v", err)
		}
		md, ok := cf.RootNode.Branches[0].Target.Metadata.(*ilert.CallFlowNodeMetadata)
		if !ok {
			t.Fatalf("expected *ilert.CallFlowNodeMetadata, got %T", cf.RootNode.Branches[0].Target.Metadata)
		}
		return md
	}

	if md := build(true); !md.AcceptAlertOnAnswer {
		t.Error("expected AcceptAlertOnAnswer to be true when accept_alert_on_answer = true")
	}
	if md := build(false); md.AcceptAlertOnAnswer {
		t.Error("expected AcceptAlertOnAnswer to be false when accept_alert_on_answer = false")
	}
}

func TestFlattenCallFlowNodeMetadata_AcceptAlertOnAnswer(t *testing.T) {
	// true is surfaced into state.
	result, err := flattenCallFlowNodeMetadata(&ilert.CallFlowNodeMetadata{AcceptAlertOnAnswer: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := result[0].(map[string]any)["accept_alert_on_answer"]; v != true {
		t.Errorf("expected accept_alert_on_answer = true in state, got %v", v)
	}

	// false is omitted (the backend default), so the key must be absent.
	result, err = flattenCallFlowNodeMetadata(&ilert.CallFlowNodeMetadata{AcceptAlertOnAnswer: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result[0].(map[string]any)["accept_alert_on_answer"]; ok {
		t.Error("expected accept_alert_on_answer to be absent from state when false")
	}
}

func TestFlattenCallFlowNodeMetadata_DisableTranscription(t *testing.T) {
	// true is surfaced into state.
	result, err := flattenCallFlowNodeMetadata(&ilert.CallFlowNodeMetadata{DisableTranscription: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := result[0].(map[string]any)["disable_transcription"]; v != true {
		t.Errorf("expected disable_transcription = true in state, got %v", v)
	}

	// false is omitted (the backend default), so the key must be absent.
	result, err = flattenCallFlowNodeMetadata(&ilert.CallFlowNodeMetadata{DisableTranscription: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result[0].(map[string]any)["disable_transcription"]; ok {
		t.Error("expected disable_transcription to be absent from state when false")
	}
}

func TestBuildCallFlow_DefineBranchesDefinitions(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceCallFlow().Schema, map[string]any{
		"name":     "test-call-flow",
		"language": "en",
		"root_node": []any{
			map[string]any{
				"node_type": "ROOT",
				"branches": []any{
					map[string]any{
						"branch_type": "ANSWERED",
						"target": []any{
							map[string]any{
								"node_type": "DEFINE_BRANCHES",
								"metadata": []any{
									map[string]any{
										"definitions": []any{
											map[string]any{
												"branch_name": "germany",
												"conditions":  "context.fromCountry == 'DE'",
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

	cf, err := buildCallFlow(d)
	if err != nil {
		t.Fatalf("unexpected error building call flow: %v", err)
	}
	md, ok := cf.RootNode.Branches[0].Target.Metadata.(*ilert.CallFlowNodeMetadata)
	if !ok {
		t.Fatalf("expected *ilert.CallFlowNodeMetadata, got %T", cf.RootNode.Branches[0].Target.Metadata)
	}
	want := []ilert.CallFlowNodeMetadataDefinition{{BranchName: "germany", Conditions: "context.fromCountry == 'DE'"}}
	if !reflect.DeepEqual(md.Definitions, want) {
		t.Errorf("Definitions = %v, want %v", md.Definitions, want)
	}
}

// TestFlattenCallFlowNode_DefineBranchesDefinitions verifies that the definitions
// of a branch node are read into state. Nested nodes decode their metadata as a
// plain map, and the definitions were left out of it, so a branch node always
// read back without them and exported without its conditions.
// See iLert/engineering-tasks#2769.
func TestFlattenCallFlowNode_DefineBranchesDefinitions(t *testing.T) {
	node := &ilert.CallFlowNode{
		NodeType: "DEFINE_BRANCHES",
		Metadata: map[string]any{
			"definitions": []any{
				map[string]any{"branchName": "germany", "conditions": "context.fromCountry == 'DE'"},
			},
		},
	}

	result, err := flattenCallFlowNode(&node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	md := result[0].(map[string]any)["metadata"].([]any)[0].(map[string]any)
	want := []any{map[string]any{"branch_name": "germany", "conditions": "context.fromCountry == 'DE'"}}
	if !reflect.DeepEqual(md["definitions"], want) {
		t.Errorf("definitions = %v, want %v", md["definitions"], want)
	}
}

func TestCallFlowDefinitionValidation(t *testing.T) {
	definition := resourceCallFlowNodeMetadata().Schema["definitions"].Elem.(*schema.Resource).Schema

	cases := []struct {
		field   string
		in      string
		wantErr bool
	}{
		{"branch_name", "germany", false},
		{"branch_name", "", true},
		{"branch_name", "  ", true},
		{"branch_name", "it's", true},
		{"branch_name", `a\b`, true},
		{"conditions", "context.fromCountry == 'DE'", false},
		{"conditions", "", true},
		{"conditions", "  ", true},
	}
	for _, tc := range cases {
		_, errs := definition[tc.field].ValidateFunc(tc.in, tc.field)
		if tc.wantErr && len(errs) == 0 {
			t.Errorf("%s=%q: expected validation error, got none", tc.field, tc.in)
		}
		if !tc.wantErr && len(errs) > 0 {
			t.Errorf("%s=%q: unexpected validation error: %v", tc.field, tc.in, errs)
		}
	}
}
