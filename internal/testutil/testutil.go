package testutil

import "github.com/google/go-cmp/cmp"

// ErrorComparer provides a way to compare errors using cmp.Diff
var ErrorComparer = cmp.Comparer(func(a, b error) bool {
	aE := a.Error()
	bE := b.Error()

	return aE == bE
})
