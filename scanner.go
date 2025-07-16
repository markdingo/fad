package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Scan statistics - only accessed via atomic functions when running concurrently
type stats struct {
	errorCount  uint32 // Primarily permissions
	dirCount    uint32
	fileCount   uint32
	otherCount  uint32
	ignoreCount uint32
}

func (s *stats) zero() {
	*s = stats{}
}

// Not concurrency-safe. Only to be called by main goroutine when all other goroutines
// have completed.
func (s *stats) sum() uint32 {
	return s.errorCount + s.dirCount + s.fileCount + s.otherCount + s.ignoreCount
}

type readDirFunc func(name string) ([]fs.DirEntry, error)
type fStatFunc func(f *os.File) (os.FileInfo, error)

func defaultFStatFunc(f *os.File) (os.FileInfo, error) {
	return f.Stat()
}

// scanner encapsulates the common structs used over the program lifetime.
type scanner struct {
	cfg           *config
	cc            *concurrencyController
	allCandidates *candidates
	baseTime      time.Time

	rdf readDirFunc // Overrides of system functions for
	fsf fStatFunc   // _testing.go functions

	wg     sync.WaitGroup
	stderr io.Writer
	stats
}

func newScanner(cfg *config, cc *concurrencyController, allCandidates *candidates,
	baseTime time.Time, stderr io.Writer) *scanner {
	return &scanner{cfg: cfg, cc: cc, allCandidates: allCandidates,
		baseTime: baseTime,
		rdf:      os.ReadDir, fsf: defaultFStatFunc,
		stderr: stderr}
}

// descend starts a new goroutine to scan the directory. Use concurrency control to limit the
// maximum number of concurrent scanners and thus how much i/o thrashing we impose on the
// file system(s).
//
// All "ignore" filters have already been applied, or appropriately ignored, by the time
// descend() is called. In particular the directories on the command line ignore
// filtering.
//
// All directories that are discovered or listed for scanning have their goroutine started
// immediately but the goroutines block on concurrencyControl. In short we use goroutines
// as our "pending" work queue.
//
// The depth parameter says how far below the starting point the dirName is from the
// starting directory. A value of zero means it is at the starting point. Since the
// minimum relevant value of maxDepth is 1, that means that when depth reaches or exceeds
// maxDepth, the descending stops.
func (scn *scanner) descend(depth uint, dirName string) {
	if scn.cfg.maxDepth.v > 0 && depth >= scn.cfg.maxDepth.v {
		return
	}

	scn.wg.Add(1) // Synchronously bump prior to starting goroutine
	atomic.AddUint32(&scn.dirCount, 1)
	go func() {
		scn.cc.start()
		scn.scan(depth, dirName)
		scn.cc.done()
		scn.wg.Done()
	}()
}

func (scn *scanner) wait() {
	scn.wg.Wait()
}

// scan dirName to find the youngest entry to add to the candidate list. During scanning,
// if a subdir is found, a new asynchronous scanner is started.
//
// Scanning starts with dirName as the primordial youngest candidate on the basis that if
// it has the most recent DTM that means that the last action in this directory was a
// deletion. Otherwise some other file entry with a more recent DTM will replace it.
//
// All "ignore" filters apply before each entry is considered as a candidate file or a
// sub-directory to scan. If maxAge is configured, files are age-checked before
// considering as a candidate.
//
// readDirFunc enables testing of error conditions which are otherwise hard to synthesize
// with testdata directories.
func (scn *scanner) scan(depth uint, dirName string) {
	var youngest candidate // Almost always populated with something

	// Populate "youngest" with parent dirName to capture possible deletion
	// activity. dirName has already been vetted by ignore if it's a subdir and is
	// purposely bypassed if it's a commandline dir.

	dirFi, err := scn.getFileInfo(dirName)
	if err != nil {
		atomic.AddUint32(&scn.errorCount, 1)
		if !scn.cfg.suppressErrors.v {
			fmt.Fprintln(scn.stderr, "Error:", err)
		}
		return
	}

	if !dirFi.IsDir() { // This is only possible if CLI path is not a dir
		atomic.AddUint32(&scn.errorCount, 1)
		if !scn.cfg.suppressErrors.v {
			fmt.Fprintln(scn.stderr, "Error:", dirName, "is not a directory")
		}
		return
	}

	// Only set youngest if this type is not being ignored
	if _, ok := scn.cfg.ignoreTypesMap[fTypeString(dirFi.Mode())]; !ok {
		youngest.set(dirName, dirFi.Mode(), scn.baseTime, dirFi.ModTime())
	} else {
		atomic.AddUint32(&scn.ignoreCount, 1)
		if scn.cfg.printIgnored.v {
			fmt.Fprintf(scn.stderr, "Ignored type %s:%s\n", fTypeString(dirFi.Mode()), dirName)
		}
	}

	// Get and scan of directory entries
	dirents, err := scn.rdf(dirName)
	if err != nil {
		atomic.AddUint32(&scn.errorCount, 1)
		if !scn.cfg.suppressErrors.v {
			fmt.Fprintln(scn.stderr, "Error: Reading Directory", dirName, err)
		}
		dirents = []fs.DirEntry{} // Set to a known quantity and continue
	}

	for _, de := range dirents {
		fi, err := de.Info() // Exclusively use FileInfo for file entry details
		if err != nil {
			atomic.AddUint32(&scn.errorCount, 1)
			if !scn.cfg.suppressErrors.v {
				fmt.Fprintln(scn.stderr, "Error:", err)
			}
			continue
		}

		path := filepath.Join(dirName, fi.Name())
		ignored := scn.cfg.ignore(path) // Apply ignore filters
		if len(ignored) > 0 {
			atomic.AddUint32(&scn.ignoreCount, 1)
			if scn.cfg.printIgnored.v {
				fmt.Fprintf(scn.stderr, "Ignored %s:%s\n", ignored, path)
			}
			continue
		}

		if fi.IsDir() { // If it's a sub-directory, descend and scan
			scn.descend(depth+1, path)
			continue
		}

		if _, ok := scn.cfg.ignoreTypesMap[fTypeString(fi.Mode())]; ok {
			if scn.cfg.printIgnored.v {
				fmt.Fprintf(scn.stderr, "Ignored type %s:%s\n", fTypeString(fi.Mode()), path)
			}
			continue
		}

		if fi.Mode().IsRegular() {
			atomic.AddUint32(&scn.fileCount, 1)
		} else {
			atomic.AddUint32(&scn.otherCount, 1)
		}
		var current candidate
		current.set(path, fi.Mode(), scn.baseTime, fi.ModTime())

		// Is current younger or equal to the previously discovered youngster?
		// With equal ages, the preference is given to the later entry. This is
		// particularly useful of the current youngest is the parent directory.
		if !youngest.isSet() || current.age.le(youngest.age) {
			youngest = current
		}
	}

	// Scan done. If a youngest was found, conditionally add to allCandidates.
	if youngest.isSet() {
		scn.allCandidates.addMaybe(&youngest)
	}
}

// getFileInfo returns the os.FileInfo of path
func (scn *scanner) getFileInfo(path string) (os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Getting File Info: %w", err)
	}
	defer f.Close()

	fi, err := scn.fsf(f)
	if err != nil {
		return nil, fmt.Errorf("Getting File Stat: %w", err)
	}

	return fi, nil
}
