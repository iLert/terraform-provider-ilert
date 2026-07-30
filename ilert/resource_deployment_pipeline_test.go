package ilert

import "testing"

func TestResourceDeploymentPipeline_CredentialsAreSensitive(t *testing.T) {
	resourceSchema := resourceDeploymentPipeline().Schema

	for _, attribute := range []string{"integration_key", "integration_url"} {
		t.Run(attribute, func(t *testing.T) {
			attributeSchema, ok := resourceSchema[attribute]
			if !ok {
				t.Fatalf("schema is missing %q", attribute)
			}

			if !attributeSchema.Computed {
				t.Errorf("expected %q to be computed", attribute)
			}
			if !attributeSchema.Sensitive {
				t.Errorf("expected %q to be sensitive", attribute)
			}
		})
	}
}
