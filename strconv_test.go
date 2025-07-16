package main

import (
	"errors"
	"testing"
)

func TestStrconv(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{"func: Action: message", "message"},
		{"Action: message", "Action: message"},
		{"func: Action:SubAction: message", "message"},
	}

	for ix, tc := range testCases {
		err := strconvTrimError(errors.New(tc.input))
		got := err.Error()
		if got != tc.output {
			t.Error(ix, "Mismatch. Expect", tc.output, "Got", got)
		}
	}
}
