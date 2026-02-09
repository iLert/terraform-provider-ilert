package ilert

import (
	"fmt"
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
