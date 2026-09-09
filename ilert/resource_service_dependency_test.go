package ilert

import "testing"

// The resource id pairs the service with the edge, because deleting or reading a
// dependency needs both and the edge id alone does not identify the endpoint's path.
func TestParseServiceDependencyID(t *testing.T) {
	serviceID, edgeID, err := parseServiceDependencyID("12/34")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serviceID != 12 || edgeID != 34 {
		t.Fatalf("got service %d edge %d, want service 12 edge 34", serviceID, edgeID)
	}
}

func TestParseServiceDependencyID_Rejects(t *testing.T) {
	for _, id := range []string{"", "12", "12/", "/34", "12/34/56/78", "abc/34", "12/abc"} {
		t.Run(id, func(t *testing.T) {
			if _, _, err := parseServiceDependencyID(id); err == nil {
				t.Fatalf("expected an error for id %q", id)
			}
		})
	}
}
