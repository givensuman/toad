package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getExitError(err error, rc int) error {
	return &exitError{rc, err}
}

func TestExitError(t *testing.T) {
	t.Run("correct error interface implementation", func(t *testing.T) {
		var err error = &exitError{0, nil}
		assert.Implements(t, (*error)(nil), err)
	})

	testCases := []struct {
		name string
		err  error
		rc   int
	}{
		{
			"errmsg empty; return code 0; casting from Error",
			nil,
			0,
		},
		{
			"errmsg empty; return code > 0; casting from Error",
			nil,
			42,
		},
		{
			"errmsg full; return code 0; casting from Error",
			errors.New("this is an error message"),
			0,
		},
		{
			"errmsg full; return code > 0; casting from Error",
			errors.New("this is an error message"),
			42,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := getExitError(tc.err, tc.rc)
			var errExit *exitError

			assert.ErrorAs(t, err, &errExit)
			assert.Equal(t, tc.rc, errExit.code)
			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), errExit.Error())
			}
		})
	}
}
