package testutil

import (
	"strings"
)

// EqualErrorMessage compares two errors by just comparing their messages.
func EqualErrorMessage(a, b error) bool {
	aMsg := "<nil>"
	if a != nil {
		aMsg = a.Error()
	}

	bMsg := "<nil>"
	if b != nil {
		bMsg = b.Error()
	}

	return strings.Compare(aMsg, bMsg) == 0
}
