package main

import (
	"strings"
	"testing"
	"time"
)

func TestAgeSet(t *testing.T) {
	testCases := []struct {
		in  string
		err string // Contained in error text
		out int64  // Only checked if no error (natch)
	}{
		{"", "1-5 digits", 0}, // # 0
		{"123456s", "1-5 digits", 0},
		{"01s", "octal", 0},
		{"1S", "invalid unit", 0},
		{"1t", "invalid unit", 0},
		{"1xs", "invalid syntax", 0},
		{"1234xs", "invalid syntax", 0},
		{"x1234s", "invalid syntax", 0},
		{"3000Y", "exceeds", 0},
		{"16000M", "exceeds", 0}, // # 9

		{"0s", "", 0 * second}, // # 10
		{"1s", "", 1 * second},
		{"59s", "", 59 * second},
		{"60s", "", 1 * minute},
		{"61s", "", 61 * second},
		{"1m", "", 1 * minute},
		{"59m", "", 59 * minute},
		{"1h", "", 1 * hour},
		{"23h", "", 23 * hour},
		{"1D", "", 1 * day},
		{"7D", "", 1 * week},
		{"1W", "", 1 * week},
		{"1Y", "", 1 * year},
	}

	for ix, tc := range testCases {
		var a age
		err := a.Set(tc.in)
		if err != nil {
			if len(tc.err) == 0 {
				t.Error(ix, "Unexpected error", err)
				continue
			}
			if !strings.Contains(err.Error(), tc.err) {
				t.Errorf("%d Wrong error returned. Want '%s' got '%s'\n", ix, tc.err, err.Error())
				continue
			}
			continue
		}
		if len(tc.err) > 0 {
			t.Error(ix, "Expected error", tc.err)
			continue
		}

		if tc.out != a.seconds {
			t.Error(ix, "Wrong value parsed. Expect", tc.out, "got", a.seconds)
			continue
		}

		if a.String() != tc.in {
			t.Error(ix, "Set value of", a.String(), "differs from input", tc.in)
			continue
		}
	}
}

func TestAgeGreaterThan(t *testing.T) {
	var baby, twin, granny age
	baby.seconds = -3
	twin.seconds = -3 // Equal ages are not greater than each other
	granny.seconds = 1
	if !granny.gt(baby, false) {
		t.Error("Granny should be GT baby")
	}
	if baby.gt(twin, false) {
		t.Error("Baby should not be greater than twin")
	}
	if twin.gt(baby, false) {
		t.Error("Twin should not be greater than baby")
	}

	baby.seconds = -3*day + 11*hour
	twin.seconds = -3 * day
	baby.multiplier = day
	twin.multiplier = day

	if baby.gt(twin, true) {
		t.Error("Truncated baby should not be greater than twin")
	}
	if twin.gt(baby, true) {
		t.Error("Truncated twin should not be greater than baby")
	}
}

func TestAgeLessThan(t *testing.T) {
	var baby, twin, granny age
	baby.seconds = -3
	twin.seconds = -3 // Equal ages are not younger than each other
	granny.seconds = 1
	if !baby.lt(granny) {
		t.Error("Baby should be younger than granny")
	}
	if baby.lt(twin) {
		t.Error("Baby should not be younger than twin")
	}
	if twin.lt(baby) {
		t.Error("Twin should not be younger than baby")
	}
}

func TestAgeSetFromTime(t *testing.T) {
	var baby, granny age

	base := time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)
	year2000 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	year3000 := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)

	granny.setFromTime(base, year2000)
	baby.setFromTime(base, year3000)

	if granny.lt(baby) {
		t.Error("Granny", granny, "should not be younger than baby", baby)
	}
}

func TestAgeCompactString(t *testing.T) {
	testCases := []struct {
		seconds int64
		expect  string
	}{
		{86401 * 366, "1Y"},
		{86401 * 99, "3M"},
		{86401 * 7 * 2, "2W"},
		{86401, "1D"},
		{60*60 + 1, "1h"},
		{121, "2m"},
		{61, "1m"},
		{59, "59s"},
		{1, "1s"},
		{0, "0s"},
		{-5, "fut"},
	}

	var a age
	for ix, tc := range testCases {
		a.seconds = tc.seconds
		s := a.compactString()
		if s != tc.expect {
			t.Error(ix, "Expect", tc.expect, "got", s)
		}
	}
}
