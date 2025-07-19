//go:build local

package main

import (
	"bytes"
	"flag"
	"testing"
)

// Check basic scanner functionality. That scan returns a single entry which is also the
// one with the most recent DTM. Unfortunately this test fails on github for reasons to do
// with their funky file-system so it's only run on local development systems.
func TestScannerBasic(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	scn, can, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}

	scn.descend(0, "testdata/scan10")
	scn.wait()

	if len(can.cf) != 1 {
		t.Fatal("Expected 1 entry from testdata/scan10, got", len(can.cf))
	}
	if can.cf[0].path != "testdata/scan10/README" {
		t.Error("Expected scan to return README, not", can.cf[0].path)
	}
}
