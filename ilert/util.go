package ilert

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Validate a value against a set of possible values
func validateStringValueFunc(values []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		value := v.(string)
		valid := false
		for _, val := range values {
			if value == val {
				valid = true
				break
			}
		}

		if !valid {
			errors = append(errors, fmt.Errorf("%#v is an invalid value for argument %s. Must be one of %#v", value, k, values))
		}
		return
	}
}

// Validate a int value against a set of possible values
func validateIntValueFunc(values []int) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		value := v.(int)
		valid := false
		for _, val := range values {
			if value == val {
				valid = true
				break
			}
		}

		if !valid {
			errors = append(errors, fmt.Errorf("%#v is an invalid value for argument %s. Must be one of %#v", value, k, values))
		}
		return
	}
}

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
