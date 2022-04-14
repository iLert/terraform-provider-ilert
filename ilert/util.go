package ilert

import (
	"fmt"
	"strconv"
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

func validateEntityIDFunc(v interface{}, keyName string) (we []string, errors []error) {
	entityIDString, ok := v.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", keyName)}
	}
	// Check that the entity ID can be converted to an int64
	if _, err := strconv.ParseInt(entityIDString, 10, 64); err != nil {
		return nil, []error{unconvertibleIDErr(entityIDString, err)}
	}

	return
}
