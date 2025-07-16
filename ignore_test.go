package main

import (
	"flag"
	"testing"
)

func testNOPConfigfunc() (string, error) { return "", nil }

func TestIgnore(t *testing.T) {
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError), testNOPConfigfunc)
	cfg.ignoreBases.v = ".profile,.ds_store,.bashrc"
	cfg.ignoreContains.v = "/pkg/mod/,tmp"
	cfg.ignoreRegexes.v = `.*\/Library\/.*Mobile.*\/`
	err := cfg.compile()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		dir    string
		base   string
		expect string
	}{
		{"/home/user", ".bashrc", "bases"},
		{"/home/user", ".profile", "bases"},
		{"/home/user", ".ds_store", "bases"},
		{"/pkg", "mod/cache/download", "contains"},
		{"/var/tmp", "testfile.gz", "contains"},
		{"~/Library", "Mobile Documents/phone.txt", "regexes"},
	}

	for ix, tc := range testCases {
		got := cfg.ignore(tc.dir + "/" + tc.base)
		if got != tc.expect {
			t.Error(ix, tc.dir, tc.base, "Got", got, "Expect", tc.expect)
		}
	}
}

func TestIgnoreBases(t *testing.T) {
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError), testNOPConfigfunc)
	cfg.ignoreBases.v = ".profile,.DS_Store:cache"
	err := cfg.compile()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		base    string
		ignored bool
	}{
		{".bashrc", false},
		{".profile", true},
		{".ds_store", false}, // Must be exact match, including case
	}

	for ix, tc := range testCases {
		if cfg.matchesBases(tc.base) != tc.ignored {
			t.Error(ix, tc.base, "Expected", tc.ignored)
		}
	}
}

func TestIgnoreContains(t *testing.T) {
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError), testNOPConfigfunc)
	cfg.ignoreContains.v = ".bashr,/go/pkg/mod/,s_sto" // Include case-mismatches
	err := cfg.compile()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		path    string
		ignored bool
	}{
		{".profile", false},
		{"/home/findactive/.bashrc", true},
		{"/home/findactive/Desktop/.DS_Store", true},
		{"/home/findactive/go/pkg/mod/cache/download/sumdb/sum.golang.org/lookup/golang.org/x/tools@v0.6.0", true},
		{"/home/findactive/go/pkg/mod/cache/download/sumdb/sum.golang.org/lookup/golang.org/x/term@v0.5.0", true},
		{"/home/findactive/go/pkg/mod/cache/download/sumdb/sum.golang.org/lookup/golang.org/x/text@v0.14.0", true},
	}

	for ix, tc := range testCases {
		if cfg.matchesContains(tc.path) != tc.ignored {
			t.Error(ix, tc.path, "Expected", tc.ignored)
		}
	}
}

func TestIgnoreRegexes(t *testing.T) {
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError), testNOPConfigfunc)
	cfg.ignoreRegexes.v = `.*profile$,\.bashrc$,.*\/go\/pkg\/.*`
	err := cfg.compile()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		path    string
		ignored bool
	}{
		{".profile", true},
		{".profile/no", false},
		{"/home/findactive/.bashrc", true},
		{"/home/findactive/Desktop/.DS_Store", false},
		{"/home/findactive/go/pkg/mod/cache/download/text@v0.14.0", true},
	}

	for ix, tc := range testCases {
		if cfg.matchesRegexes(tc.path) != tc.ignored {
			t.Error(ix, tc.path, "Expected", tc.ignored)
		}
	}
}
