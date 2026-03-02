package ilert

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
