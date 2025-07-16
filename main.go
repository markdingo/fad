package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	os.Exit(realMain(time.Now(), os.Args[1:], os.UserConfigDir, os.Stdout, os.Stderr))
}

// realMain does all the work and can more easily be the target of testing
func realMain(start time.Time, args []string, confFunc userConfigDirFunc, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(Name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprintln(stderr, "Consider -h for option details") }
	cfg := newConfig(fs, confFunc)
	err := cfg.loadDefaults() // Load early so setFlags see revised defaults
	if err != nil {           // Config load errors only generate warnings
		fmt.Fprintln(stderr, "Warning:", err)
	}

	cfg.setFlags()
	err = fs.Parse(args)
	if err != nil {
		return EX_USAGE
	}

	// Documentation requests usurp all scanning options
	if cfg.help {
		printUsage(stdout, stderr, false, fs, cfg)
		return EX_OK
	}

	if cfg.version {
		printVersion(stdout)
		return EX_OK
	}

	if cfg.manpage {
		fmt.Fprint(stdout, Manpage)
		return EX_OK
	}

	// We're actually going to run a scan
	err = cfg.compile()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return EX_USAGE
	}

	scanList := fs.Args()
	if len(scanList) == 0 { // If none supplied, scan current working directory.
		scanList = append(scanList, ".")
	}

	allCandidates := newCandidates(int(cfg.maxCount.v), cfg.maxAge)
	cc := newConcurrencyController(int(cfg.maxScanners.v))
	scn := newScanner(cfg, cc, allCandidates, start, stderr)
	for _, dirName := range scanList {
		dirName = filepath.Clean(dirName) // Clean here so we can avoid .Join/Clean later
		scn.descend(0, dirName)           // Runs a goroutine
	}
	scn.wait() // Wait for all goroutines started by scn.descend()

	// Sort and print
	scn.allCandidates.sortAscending()
	end := time.Now()
	secs := end.Sub(start)

	scn.printCandidates(stdout)
	if scn.cfg.printStats.v {
		scn.printStats(stdout, secs)
	}

	// If any access errors occurred, exit non-zero
	if scn.stats.errorCount > 0 {
		return EX_OSFILE
	}

	return EX_OK
}
