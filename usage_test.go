package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	fs := flag.NewFlagSet(Name, flag.ContinueOnError)
	cfg := newConfig(fs, nil)
	printUsage(&stdout, &stderr, false, fs, cfg)
	out := stdout.String()
	err := stderr.String()
	if !strings.Contains(out, "recursively") {
		t.Error("Expected usage in stdout", out)
	}

	if len(err) > 0 {
		t.Error("Did not expect usage in stderr", len(err))
	}

	stdout.Reset()
	stderr.Reset()
	printUsage(&stdout, &stderr, true, fs, cfg)
	out = stdout.String()
	err = stderr.String()

	if !strings.Contains(err, "recursively") {
		t.Error("Expected usage in stderr", out)
	}

	if len(out) > 0 {
		t.Error("Did not expect usage in stdout", len(err))
	}

	cfg.configPathHelp = "Test"
	stdout.Reset()
	stderr.Reset()
	printUsage(&stdout, &stderr, true, fs, cfg)
	out = stdout.String()
	err = stderr.String()

	if !strings.Contains(err, "Test") {
		t.Error("Usage did not detect configPathHelpError out:", out, "err:", err)
	}
}
