package main

import (
	"strings"
	"testing"
)

func TestFlagsSet(t *testing.T) {
	var bf boolFlag
	var ui uintFlag
	var cs commaStringFlag

	testCases := []struct {
		fv     flagValue
		input  string
		error  string
		expect string // Only checked if len(expect) > 0
	}{
		{&bf, "true", "", "true"},
		{&bf, "F", "", "false"},
		{&bf, "junk", "invalid", ""},

		{&ui, "23", "", "23"},
		{&ui, "-23", "invalid", ""},
		{&ui, "xx23", "invalid", ""},

		// These test cases rely on &cs state being retained across calls
		{&cs, "a,b", "", "a,b"},
		{&cs, "+c,d", "", "a,b,c,d"},
		{&cs, "e,f", "", "e,f"},
		{&cs, "", "", ""},        // Clear
		{&cs, "+c,d", "", "c,d"}, // Append to empty
	}

	for ix, tc := range testCases {
		err := tc.fv.Set(tc.input)
		if err != nil && len(tc.error) == 0 {
			t.Error(ix, "Unexpected error", err)
			continue
		}

		got := ""
		if err != nil {
			got = err.Error()
		}
		if !strings.Contains(got, tc.error) {
			t.Errorf("%d Error mismatch. Expected '%s', got '%s'\n", ix, tc.error, got)
			continue
		}

		if len(tc.expect) > 0 {
			got := tc.fv.String()
			if got != tc.expect {
				t.Errorf("%d String() mismatch. Expected '%s', got '%s'\n",
					ix, tc.expect, got)
			}
		}
	}
}

func TestFlagMin(t *testing.T) {
	var f uintFlag
	f.min = 1         // Set minimum to 1
	err := f.Set("0") // and try and set to zero
	if err == nil {
		t.Fatal("Expected an error trying to set a min value to zero")
	}

	got := err.Error()
	exp := "Less than minimum"
	if !strings.Contains(got, exp) {
		t.Error("Expected 'Less than minimum', but got", got)
	}
}
