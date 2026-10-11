package ilert

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Every expectation here was checked against the server's own parser
// (TemplateDSLEngine.fromText, verified 09.10.2026): a pair is equivalent exactly when
// the server stores both templates as the same elements.
func TestSuppressEquivalentTemplateDiff(t *testing.T) {
	cases := []struct {
		name     string
		old, new string
		want     bool
	}{
		// #165: what the API returned against what the configuration said
		{"space after a comma", `{{ subject.splitTakeAt(" ",0) }}`, `{{ subject.splitTakeAt(" ", 0) }}`, true},
		{"spaces around the arguments", `{{ subject.splitTakeAt(" ",4) }}`, `{{  subject.splitTakeAt( " " , 4 )  }}`, true},
		{"no padding inside the braces", `{{ subject.splitTakeAt(" ",0) }}`, `{{subject.splitTakeAt(" ",0)}}`, true},
		{"plain variable without padding", `{{ subject }}`, `{{subject}}`, true},
		{"escaped space followed by a plain one", `{{ subject.replaceAll(a\ b,"-") }}`, `{{ subject.replaceAll(a\  b,"-") }}`, true},
		{"several variables in multiline text",
			"Alert {{ alert.id }}: {{ subject.splitTakeAt(\" \",1) }}\nsee {{ details }}",
			"Alert {{alert.id}}: {{ subject.splitTakeAt(\" \", 1) }}\nsee {{ details }}", true},

		// real changes
		{"different argument", `{{ subject.splitTakeAt(" ",0) }}`, `{{ subject.splitTakeAt(" ", 1) }}`, false},
		{"space inside a quoted argument", `{{ subject.splitTakeAt(" ",0) }}`, `{{ subject.splitTakeAt("  ",0) }}`, false},
		{"space in the text around a variable", `Alert {{ alert.id }}`, `Alert  {{ alert.id }}`, false},
		// spaces in the path turn the function call into a plain variable named
		// `subject . splitTakeAt(" ",0)`
		{"spaces in the path", `{{ subject.splitTakeAt(" ",0) }}`, `{{ subject . splitTakeAt(" ",0) }}`, false},
		{"space inside a plain variable", `{{ subject }}`, `{{ sub ject }}`, false},
		{"legacy syntax keeps argument spaces", `{{ subject##splitTakeAt(( ||0)) }}`, `{{ subject##splitTakeAt((  ||0)) }}`, false},
		{"escaped space removed", `{{ subject.replaceAll(a\ b,"-") }}`, `{{ subject.replaceAll(a\b,"-") }}`, false},
		{"template added", "", `{{ subject }}`, false},
		{"template removed", `{{ subject }}`, "", false},

		// equivalent on the server, but rewriting the port does not model, so these stay
		// a diff as before
		{"quotes the server adds around a bare argument", `{{ subject.replaceAll("a\ b","-") }}`, `{{ subject.replaceAll(a\ b,"-") }}`, false},
		{"legacy syntax the server rewrites", `{{ subject.splitTakeAt(" ",0) }}`, `{{ subject##splitTakeAt(( ||0)) }}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := suppressEquivalentTemplateDiff("text_template", tc.old, tc.new, nil); got != tc.want {
				t.Fatalf("suppressEquivalentTemplateDiff(%q, %q) = %v, want %v", tc.old, tc.new, got, tc.want)
			}
		})
	}
}

// Every expectation here was checked against Service.setAlias in the monolith (verified
// 09.10.2026): old is the value the API returned for an alias sent as new.
func TestSuppressEquivalentServiceAliasDiff(t *testing.T) {
	cases := []struct {
		name     string
		old, new string
		want     bool
	}{
		{"capitals", "my-service", "My-Service", true},
		{"surrounding spaces", "my-service", " my-service ", true},
		{"surrounding tab and newline", "my-service", "\tmy-service\n", true},
		{"several aliases", "db, api", "DB, API", true},

		{"different alias", "my-service", "other-service", false},
		{"inner space", "my-service", "my -service", false},
		// stored before the API started lowercasing aliases, so the update is real
		{"mixed case stored by an older API", "My-Service", "my-service", false},
		// Java lowercases a final sigma to ς, Go to σ
		{"non-ASCII is never normalized", "ας", "ΑΣ", false},
		{"alias added", "", "my-service", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := suppressEquivalentServiceAliasDiff("alias", tc.old, tc.new, nil); got != tc.want {
				t.Fatalf("suppressEquivalentServiceAliasDiff(%q, %q) = %v, want %v", tc.old, tc.new, got, tc.want)
			}
		})
	}
}

// Every text template of an alert source goes through the same server parser, so each
// one needs the suppression, including any added later.
func TestAlertSourceTextTemplatesSuppressReformatting(t *testing.T) {
	var found []string
	var walk func(prefix string, s map[string]*schema.Schema)
	walk = func(prefix string, s map[string]*schema.Schema) {
		for name, attr := range s {
			path := prefix + name
			if name == "text_template" {
				found = append(found, path)
				if attr.DiffSuppressFunc == nil {
					t.Errorf("%s has no DiffSuppressFunc", path)
				}
			}
			if elem, ok := attr.Elem.(*schema.Resource); ok {
				walk(path+".", elem.Schema)
			}
		}
	}
	walk("", resourceAlertSource().Schema)

	if len(found) < 9 {
		t.Fatalf("expected at least 9 text_template attributes, found %d: %v", len(found), found)
	}
}

// testPlanDiff returns the diff a plan computes between the prior state and the new
// configuration, which is where a DiffSuppressFunc takes effect.
func testPlanDiff(t *testing.T, resource *schema.Resource, priorRaw, raw map[string]any) *terraform.InstanceDiff {
	t.Helper()

	prior := schema.TestResourceDataRaw(t, resource.Schema, priorRaw)
	prior.SetId("1")

	diff, err := schema.InternalMap(resource.Schema).Diff(context.Background(), prior.State(), terraform.NewResourceConfigRaw(raw), nil, nil, true)
	if err != nil {
		t.Fatalf("unexpected error diffing: %v", err)
	}
	return diff
}

func testAlertSourceWithTemplates(valueTemplate, servicesTemplate string) map[string]any {
	return map[string]any{
		"name":              "test-alert-source",
		"integration_type":  "API",
		"escalation_policy": "1",
		"priority_template": []any{map[string]any{
			"value_template": []any{map[string]any{"text_template": valueTemplate}},
			"mapping":        []any{map[string]any{"value": "P1", "priority": "HIGH"}},
		}},
		"severity_template": []any{map[string]any{
			"value_template": []any{map[string]any{"text_template": valueTemplate}},
			"mapping":        []any{map[string]any{"value": "P1", "severity": 1}},
		}},
		"services_template": []any{map[string]any{"text_template": servicesTemplate}},
	}
}

// Regression for #165: the configuration from the issue against the state the API left
// behind. Before the fix this plan changed all three templates on every run.
func TestAlertSourcePlan_IgnoresTemplatesTheAPIReformatted(t *testing.T) {
	diff := testPlanDiff(t, resourceAlertSource(),
		testAlertSourceWithTemplates(`{{ subject.splitTakeAt(" ",0) }}`, `{{ subject.splitTakeAt(" ",4) }}`),
		testAlertSourceWithTemplates(`{{ subject.splitTakeAt(" ", 0) }}`, `{{ subject.splitTakeAt(" ", 4) }}`))

	if diff != nil {
		for key, attr := range diff.Attributes {
			if strings.HasSuffix(key, "text_template") && attr.Old != attr.New {
				t.Errorf("unexpected diff on %s: %q => %q", key, attr.Old, attr.New)
			}
		}
	}
}

func TestAlertSourcePlan_KeepsRealTemplateChanges(t *testing.T) {
	diff := testPlanDiff(t, resourceAlertSource(),
		testAlertSourceWithTemplates(`{{ subject.splitTakeAt(" ",0) }}`, `{{ subject.splitTakeAt(" ",4) }}`),
		testAlertSourceWithTemplates(`{{ subject.splitTakeAt(" ", 1) }}`, `{{ subject.splitTakeAt(" ", 4) }}`))

	for _, key := range []string{
		"priority_template.0.value_template.0.text_template",
		"severity_template.0.value_template.0.text_template",
	} {
		if diff == nil || diff.Attributes[key] == nil || diff.Attributes[key].New != `{{ subject.splitTakeAt(" ", 1) }}` {
			t.Errorf("expected %s to change to the new index", key)
		}
	}
	if diff != nil {
		if attr := diff.Attributes["services_template.0.text_template"]; attr != nil && attr.Old != attr.New {
			t.Errorf("unexpected diff on the unchanged services template: %q => %q", attr.Old, attr.New)
		}
	}
}

func TestServicePlan_AliasNormalization(t *testing.T) {
	cases := []struct {
		name, config string
		wantDiff     bool
	}{
		{"capitals the API lowercased", "My-Service", false},
		{"spaces the API trimmed", " my-service ", false},
		{"a different alias", "other-service", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diff := testPlanDiff(t, resourceService(),
				map[string]any{"name": "test-service", "alias": "my-service"},
				map[string]any{"name": "test-service", "alias": tc.config})

			attr := (*terraform.ResourceAttrDiff)(nil)
			if diff != nil {
				attr = diff.Attributes["alias"]
			}
			gotDiff := attr != nil && attr.Old != attr.New
			if gotDiff != tc.wantDiff {
				t.Fatalf("alias diff = %v, want %v (diff: %+v)", gotDiff, tc.wantDiff, attr)
			}
		})
	}
}
