package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
)

type testDirEntry struct {
	name     string
	isDir    bool
	fileMode fs.FileMode
	fileInfo fs.FileInfo
	err      error
}

func (tde *testDirEntry) Name() string               { return tde.name }
func (tde *testDirEntry) IsDir() bool                { return tde.isDir }
func (tde *testDirEntry) Type() fs.FileMode          { return tde.fileMode }
func (tde *testDirEntry) Info() (fs.FileInfo, error) { return tde.fileInfo, tde.err }

type testFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

func (tfi *testFileInfo) Name() string       { return tfi.name }
func (tfi *testFileInfo) Size() int64        { return tfi.size }
func (tfi *testFileInfo) Mode() fs.FileMode  { return tfi.mode }
func (tfi *testFileInfo) ModTime() time.Time { return tfi.modTime }
func (tfi *testFileInfo) IsDir() bool        { return tfi.isDir }
func (tfi *testFileInfo) Sys() any           { return tfi.sys }

type testDir struct {
	err     error
	dirents []fs.DirEntry
}

func (td *testDir) readDir() ([]fs.DirEntry, error) {
	return td.dirents, td.err
}

func testScannerSetup(cfg *config, stderr io.Writer, ccCount int) (*scanner, *candidates, error) {
	cfg.setInternalDefaults()
	err := cfg.compile()
	if err != nil {
		return nil, nil, err
	}
	cc := newConcurrencyController(ccCount)
	now := time.Now()
	can := newCandidates(20, age{})
	scn := newScanner(cfg, cc, can, now, stderr)

	return scn, can, err
}

func TestStats(t *testing.T) {
	var s1 stats
	if s1.sum() != 0 {
		t.Error("Initial state should be zero, not", s1.sum())
	}

	s1.errorCount++
	if s1.sum() != 1 {
		t.Error("error++ should be 1, not", s1.sum())
	}

	s1.zero()
	if s1.sum() != 0 {
		t.Error("zero() did not zero", s1.sum(), s1)
	}
}

// Test for readDir error return
func TestScannerReaddir(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	scn, _, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}

	td := testDir{err: errors.New("Error One")}
	scn.rdf = func(name string) ([]fs.DirEntry, error) { return td.readDir() }
	scn.scan(0, "testdata")
	scn.wait()

	if scn.stats.errorCount != 1 || scn.stats.ignoreCount != 1 || scn.stats.sum() != 2 {
		t.Error("Expected error,ignore,sum == '1 1 2' not",
			scn.stats.errorCount, scn.stats.ignoreCount, scn.stats.sum(), scn.stats)
	}
	got := stderr.String()
	exp := "Error One\n"
	if !strings.Contains(got, exp) {
		t.Errorf("Expected stderr to contain '%s' got '%s'\n", exp, got)
	}
}

// Test for empty directory
func TestScannerEmpty(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	scn, _, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}

	td := testDir{}
	scn.rdf = func(name string) ([]fs.DirEntry, error) { return td.readDir() }
	scn.scan(0, "testdata")
	scn.wait()

	if scn.stats.ignoreCount != 1 || scn.stats.sum() != 1 {
		t.Error("Empty dir should only set ignoreCount", scn.stats)
	}
}

// Test for dirents.Info() error return
func TestScannerDirents(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	scn, _, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}

	td := testDir{}
	td.dirents = append(td.dirents, &testDirEntry{err: errors.New("td error one")})
	scn.rdf = func(name string) ([]fs.DirEntry, error) { return td.readDir() }
	scn.scan(0, "testdata")
	scn.wait()

	exp := stderr.String()
	if scn.stats.errorCount != 1 || !strings.Contains(exp, "error one") {
		t.Error("dirents.Info should have returned error", scn.stats, exp)
	}
}

// Test for ignoring irregular
func TestScannerIrregular(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	scn, _, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}

	td := testDir{}
	tde := &testDirEntry{fileMode: fs.ModeIrregular,
		fileInfo: &testFileInfo{name: "testfile", mode: fs.ModeIrregular}}
	td.dirents = append(td.dirents, tde)
	scn.rdf = func(name string) ([]fs.DirEntry, error) { return td.readDir() }
	scn.scan(0, "testdata")
	scn.wait()

	if scn.stats.errorCount != 0 || scn.stats.otherCount != 1 {
		t.Error("scan did not detect irregular", scn.stats)
	}
}

// Test for ignoring candidate
func TestScannerCandidate(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	cfg.ignoreBases.v = "testfileIgnore"
	cfg.printIgnored.v = true
	scn, _, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}

	td := testDir{}
	tde := &testDirEntry{fileInfo: &testFileInfo{name: "testfileIgnore"}}
	td.dirents = append(td.dirents, tde)
	scn.rdf = func(name string) ([]fs.DirEntry, error) { return td.readDir() }
	scn.scan(0, "testdata")
	scn.wait()

	if scn.stats.errorCount != 0 || scn.stats.ignoreCount != 2 {
		t.Error("scan didn't ignore parent dir or ignoreBases string", scn.stats)
	}

	got := stderr.String()
	for _, exp := range []string{"Ignored bases:", "Ignored type d:"} {
		if !strings.Contains(got, exp) {
			t.Errorf("Ignore message is not prefixed with '%s' got '%s'\n", exp, got)
		}
	}
}

func TestScannerMaxDepth(t *testing.T) {
	testCases := []struct {
		maxDepth           uint
		expectedCandidates int
	}{
		{0, 3},
		{1, 0}, // No files in testdata/maxdir, just subdirs
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 3},
		{10, 3},
	}

	for ix, tc := range testCases {
		var stderr bytes.Buffer
		cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
			func() (string, error) { return "", nil })
		cfg.suppressErrors.v = false
		cfg.maxDepth.v = tc.maxDepth
		scn, can, err := testScannerSetup(cfg, &stderr, 10)
		if err != nil {
			t.Fatal(err)
		}

		scn.descend(0, "testdata/maxdir")
		scn.wait()

		got := len(can.cf)
		if got != tc.expectedCandidates {
			t.Error(ix, "Candidates expected", tc.expectedCandidates, "got", got)
		}
	}
}

func TestScannerMaxAge(t *testing.T) {
	testCases := []struct {
		name   string
		offset time.Duration
	}{
		{"3min", -3 * time.Minute},
		{"chickenDinner", -1 * time.Minute},
		{"2min", -2 * time.Minute},
	}
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	cfg.maxAge.seconds = 120
	scn, can, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	td := testDir{}
	for _, tc := range testCases {
		tde := &testDirEntry{fileInfo: &testFileInfo{name: tc.name, modTime: now.Add(tc.offset)}}
		td.dirents = append(td.dirents, tde)
	}
	scn.rdf = func(name string) ([]fs.DirEntry, error) { return td.readDir() }
	scn.scan(0, ".")
	scn.wait()

	if len(can.cf) != 1 {
		t.Error("Expected one candidate got", len(can.cf))
	}
	if can.cf[0].path != "chickenDinner" { // Is it a "winner winner"?
		t.Error("Expected a Chicken Dinner. Got", can.cf[0])
	}
}

func TestScannerScanError(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	scn, _, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}
	scn.scan(0, "testdata/noexist")
	scn.wait()
	if scn.stats.errorCount != 1 || scn.stats.sum() != 1 {
		t.Error("Expected error,sum == '1 1' not",
			scn.stats.errorCount, scn.stats.sum(), scn.stats)
	}
	got := stderr.String()
	exp := "no such file"
	if !strings.Contains(got, exp) {
		t.Error("Expected", exp, "Got", got)
	}
}

// Test that starting directory gets set as youngest if directory entries are not ignored.
func TestScannerScanYoungest(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	cfg.ignoreTypes.v = fTypeTemporary // Needs to be set to avoid defaults
	cfg.maxDepth.v = 1
	scn, can, err := testScannerSetup(cfg, &stderr, 10)
	if err != nil {
		t.Fatal(err)
	}
	scn.descend(0, "testdata")
	scn.wait()
	if scn.stats.dirCount != 1 || scn.stats.sum() != 1 {
		t.Error("Expected dir,sum == '1 1' not",
			scn.stats.dirCount, scn.stats.sum(), scn.stats)
	}

	if len(can.cf) != 1 {
		t.Fatal("Expected one candidate, not", len(can.cf))
	}
}

// Test that starting directory gets set as youngest if directory entries are not ignored.
func TestScannerIgnoreTypes(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	cfg.ignoreTypes.v = fTypeDir + "," + fTypeFile
	cfg.printIgnored.v = true
	cfg.maxDepth.v = 2
	scn, can, err := testScannerSetup(cfg, &stderr, 1)
	if err != nil {
		t.Fatal(err)
	}
	scn.descend(0, "testdata")
	scn.wait()
	if scn.stats.dirCount != scn.stats.ignoreCount {
		t.Error("Expected dir == ignore, not",
			scn.stats.dirCount, scn.stats.ignoreCount, scn.stats)
	}

	if len(can.cf) != 0 {
		t.Fatal("Did not expect any candidates, but got", len(can.cf))
	}

	got := stderr.String()
	exp := "Ignored type d"
	if !strings.Contains(got, exp) {
		t.Error("Expected", exp, "got", got)
	}
}

func testFsfError(f *os.File) (fs.FileInfo, error) {
	return nil, errors.New("fsf Failed")
}

// Test for fstat() failure
func TestScannerStatFail(t *testing.T) {
	var stderr bytes.Buffer
	cfg := newConfig(flag.NewFlagSet(Name, flag.ContinueOnError),
		func() (string, error) { return "", nil })
	scn, _, err := testScannerSetup(cfg, &stderr, 1)
	if err != nil {
		t.Fatal(err)
	}
	scn.fsf = testFsfError
	scn.descend(0, "testdata")
	scn.wait()

	got := stderr.String()
	exp := "Getting File Stat"
	if !strings.Contains(got, exp) {
		t.Error("Expected", exp, "got", got)
	}
}
