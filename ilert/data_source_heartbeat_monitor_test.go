package ilert

import "testing"

func TestDataSourceHeartbeatMonitor_CredentialsAreSensitive(t *testing.T) {
	dataSourceSchema := dataSourceHeartbeatMonitor().Schema

	for _, attribute := range []string{"integration_key", "integration_url"} {
		t.Run(attribute, func(t *testing.T) {
			attributeSchema, ok := dataSourceSchema[attribute]
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
