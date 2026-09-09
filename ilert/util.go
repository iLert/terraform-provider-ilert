package ilert

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/iLert/ilert-go/v3"
)

func unconvertibleIDErr(id string, err error) *unconvertibleIDError {
	return &unconvertibleIDError{OriginalID: id, OriginalError: err}
}

type unconvertibleIDError struct {
	OriginalID    string
	OriginalError error
}

func (e *unconvertibleIDError) Error() string {
	return fmt.Sprintf("Unexpected ID format (%q), expected numerical ID. %s",
		e.OriginalID, e.OriginalError.Error())
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// flattenLabels turns the API's label map into the shape a TypeMap attribute expects,
// following the same rule as flattenTeamShortList: labels only reach the state once the
// configuration declares them. The API leaves the labels untouched when the field is
// absent from a write (verified 09.09.2026), so a configuration that never declared them
// keeps them unmanaged — writing them into state would make the next plan propose their
// removal and the next apply carry it out, silently taking over labels set in the web
// app. A configuration that does declare them sees drift as a normal diff.
func flattenLabels(labels *map[string]string, d *schema.ResourceData, key string) map[string]any {
	result := make(map[string]any)
	if labels == nil {
		return result
	}
	if _, declared := d.GetOk(key); !declared && d.Id() != "" {
		return result
	}
	for k, v := range *labels {
		result[k] = v
	}
	return result
}

// derefTeamShortList unwraps the pointer-typed Teams of the endpoints that need an
// absent field to mean "leave the teams untouched", so the read path can share
// flattenTeamShortList with the resources whose Teams is a plain slice.
func derefTeamShortList(teams *[]ilert.TeamShort) []ilert.TeamShort {
	if teams == nil {
		return nil
	}
	return *teams
}

// flattenLabelsAll exposes every label the API returned. Data sources report what is on
// the server rather than reconciling a configuration, so they have nothing to gate on.
func flattenLabelsAll(labels *map[string]string) map[string]any {
	result := make(map[string]any)
	if labels == nil {
		return result
	}
	for k, v := range *labels {
		result[k] = v
	}
	return result
}
