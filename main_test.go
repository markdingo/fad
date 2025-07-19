package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMainOptions(t *testing.T) {
	testCases := []struct {
		options  []string
		config   string
		excode   int
		out, err string // Contained in error text
	}{
		{[]string{"--unknown"}, "", EX_USAGE, "", "Consider -h for option"},
		{[]string{""}, "testdata/main1", EX_OK, "main1", "at testdata"}, // Bad config
		{[]string{"-v"}, "", EX_OK, "Version:", ""},
		{[]string{"-h"}, "", EX_OK, "SYNOPSIS", ""},
		{[]string{"--manpage"}, "", EX_OK, ".Nm fad", ""},
		{[]string{"--count", "-1"}, "", EX_USAGE, "", "invalid value"},
		{[]string{}, "", EX_OK, ":f:", ""}, // scanList default to "."
		{[]string{"-pstats"}, "", EX_OK, "Elapse:", ""},
		{[]string{"/dev/null"}, "", EX_OSFILE, "", "/dev/null is not a directory"},
		{[]string{"--iregexes", `aa\`}, "", EX_USAGE, "", "does not compile"},
	}

	for ix, tc := range testCases {
		var stdout, stderr bytes.Buffer
		ex := realMain(time.Now(), tc.options,
			func() (string, error) { return tc.config, nil },
			&stdout, &stderr)
		if ex != tc.excode {
			t.Error(ix, "Expected exit code of", tc.excode, "got", ex)
		}

		out := stdout.String()
		if len(tc.out) == 0 {
			if len(out) > 0 {
				t.Errorf("%d Unexpected stdout: '%s'\n", ix, out)
			}
		} else {
			if !strings.Contains(out, tc.out) {
				t.Errorf("%d stdout does not contain '%s' got '%s'\n", ix, tc.out, out)
			}
		}

		err := stderr.String()
		if len(tc.err) == 0 {
			if len(err) > 0 {
				t.Errorf("%d Unexpected stderr: '%s'\n", ix, err)
			}
		} else {
			if !strings.Contains(err, tc.err) {
				t.Errorf("%d stderr does not contain '%s' got '%s'\n", ix, tc.err, err)
			}
		}
	}
}
