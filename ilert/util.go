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

func removeStringsFromSlice(l []string, s ...string) []string {
	n := make([]string, 0)
	for _, v := range l {
		if !StringSliceContains(s, v) {
			n = append(n, v)
		}
	}
	return n
}

func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
