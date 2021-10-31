package testutil

import (
	"errors"
	"testing"
)

type testError struct{}

func (e testError) Error() string {
	return "test message"
}

func TestEqualErrorMessage(t *testing.T) {
	tt := []struct {
		desc      string
		a         error
		b         error
		wantEqual bool
	}{
		{
			desc:      "two nil",
			a:         nil,
			b:         nil,
			wantEqual: true,
		},
		{
			desc:      "a not nil",
			a:         errors.New("error A"),
			b:         nil,
			wantEqual: false,
		},
		{
			desc:      "b not nil",
			a:         nil,
			b:         errors.New("error B"),
			wantEqual: false,
		},
		{
			desc:      "both not nil",
			a:         errors.New("error A"),
			b:         errors.New("error B"),
			wantEqual: false,
		},
		{
			desc:      "equal message, same type",
			a:         errors.New("test message"),
			b:         errors.New("test message"),
			wantEqual: true,
		},
		{
			desc:      "equal message, different type",
			a:         errors.New("test message"),
			b:         testError{},
			wantEqual: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			equal := EqualErrorMessage(tc.a, tc.b)
			if equal != tc.wantEqual {
				t.Errorf("got equal %v, want %v", equal, tc.wantEqual)
			}
		})
	}
}
