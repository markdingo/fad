package main

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

// Test basic functions of load, including that all flags are over-ridden.
func TestConfigLoadBasic(t *testing.T) {
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", errors.New("An error") })
	err := cfg.loadDefaults()
	if err == nil {
		t.Error("Expected 'An error', not nil")
	}
	if cfg.maxCount.v != defaultPrintLimit || cfg.maxScanners.v != defaultScannerLimit {
		t.Error("loadDefaults did not set defaults", cfg.maxCount.v, cfg.maxScanners.v)
	}

	cfg = newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "testdata/loadall", nil })
	err = cfg.loadDefaults()
	if err != nil {
		t.Error(err)
	}

	err = cfg.flagSet.Parse([]string{})
	if err != nil {
		t.Error(err)
	}

	if cfg.printDirname.v != true {
		t.Error("printDirname should be true, not", cfg.printDirname)
	}
	if cfg.printIgnored.v != true {
		t.Error("pignored should be true, not", cfg.printIgnored)
	}
	if cfg.printStats.v != true {
		t.Error("pstats should be true, not", cfg.printStats)
	}
	if cfg.suppressErrors.v != true {
		t.Error("perrors should be true")
	}

	if cfg.maxAge.seconds != 7*86400 {
		t.Error("age should be 1week, not", cfg.maxAge.seconds)
	}
	if cfg.maxCount.v != 123 {
		t.Error("count should be 123, not", cfg.maxCount)
	}
	if cfg.maxScanners.v != 11 {
		t.Error("scanners should be 11, not", cfg.maxScanners)
	}

	if cfg.ignoreBases.v != "ignoreBases" {
		t.Error("ibase should be 'ignoreBases', not", cfg.ignoreBases)
	}
	if cfg.ignoreContains.v != "ignoreContains" {
		t.Error("icontain should be 'ignoreContains', not", cfg.ignoreContains)
	}
	if cfg.ignoreRegexes.v != "ignorePatterns" {
		t.Error("ipattern should be 'ignorePatterns', not", cfg.ignoreRegexes)
	}
	if cfg.ignoreTypes.v != "p,d" {
		t.Error("ipattern should be 'p,d', not", cfg.ignoreTypes)
	}
}

func TestConfigLoadErrors(t *testing.T) {
	testCases := []struct {
		configFile string
		error      string
		path       string
	}{
		{"notPresent", "", "not present"},
		{"notdir", "not a directory", ""},
		{"onefield", "expects one", ""},
		{"twofields", "one argument, not 2", ""},
		{"unknown", "Unknown option", ""},
		{"duplicate", "Duplicate", ""},
		{"badsetage", "invalid unit", ""},
		{"badsetbool", "invalid syntax", ""},
		{"badsetuint", "invalid syntax", ""},
	}

	for ix, tc := range testCases {
		cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
			func() (string, error) { return "testdata/" + tc.configFile, nil })
		err := cfg.loadDefaults()
		switch {
		case err == nil && len(tc.error) == 0:
			if !strings.Contains(cfg.configPathHelp, tc.path) {
				t.Error(ix, "Expected path to contain", tc.path, cfg.configPathHelp)
			}
		case err == nil && len(tc.error) > 0:
			t.Error(ix, "Expected Error for", tc.configFile)
		case err != nil && len(tc.error) == 0:
			t.Error(ix, "Unexpected error", err)
		case !strings.Contains(err.Error(), tc.error): // Must be err != nil && len(tc.error) > 0
			t.Errorf("%d Expected error to contain '%s', but got '%s'\n",
				ix, tc.error, err)
		}
	}
}

func TestConfigCompile(t *testing.T) {
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", errors.New("An error") })

	cfg.ignoreRegexes.v = `.*,data\,^dog$`
	err := cfg.compile()
	if err == nil {
		t.Error("Expected regex compile error")
	}
	exp := "does not compile"
	got := err.Error()
	if !strings.Contains(got, exp) {
		t.Error("Error does not contain", exp, got)
	}

	cfg.ignoreRegexes.v = ""
	cfg.ignoreTypes.v = "D,l,d"
	err = cfg.compile()
	if err == nil {
		t.Error("Expected ignore types compile error")
	}
	exp = "Error: -itypes"
	got = err.Error()
	if !strings.Contains(got, exp) {
		t.Error("Error does not contain", exp, got)
	}
}
