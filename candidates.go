package main

import (
	"io/fs"
	"sort"
	"sync"
	"time"
)

// candidates contains a slice of all current candidates that will be printed.
type candidates struct {
	mu         sync.Mutex
	maxEntries int
	maxAge     age
	oldest     int // Index into cf of oldest candidate
	cf         []*candidate
}

func newCandidates(maxEntries int, maxAge age) *candidates {
	return &candidates{maxEntries: maxEntries, maxAge: maxAge}
}

// candidate contains the details of the candidate path/file which will potentially be
// printed at the end of all scanning. It's called a candidate because it can be evicted
// by a younger candidate.
type candidate struct {
	path string
	mode fs.FileMode
	age  age
}

func (c *candidate) set(path string, mode fs.FileMode, baseTime, modTime time.Time) {
	c.path = path
	c.mode = mode
	c.age.setFromTime(baseTime, modTime)
}

// isSet says whether this candidate is meaningful or not.
func (c *candidate) isSet() bool {
	return len(c.path) > 0
}

func (c *candidate) isDir() bool {
	return c.mode.IsDir()
}

// ftype returns a printable rendition of the file system type of the candidate
func (c *candidate) fType() string {
	return fTypeString(c.mode)
}

// sortAscending sort candidates from youngest to oldest
func (can *candidates) sortAscending() {
	sort.Slice(can.cf,
		func(i, j int) bool {
			return can.cf[i].age.seconds < can.cf[j].age.seconds
		})
}

// addMaybe conditionally adds the candidate depending on maxAge, maxEntries, the count of
// current entries and the age of the oldest entry.
//
// 1) if candidate is older than maxAge (and maxAge is set), discard.
// 2) if entryCount < maxEntries (or maxEntries not set), add.
// 3) if candidate is younger than oldest, replace.
// 4) Discard.
//
// Return true if the candidate is added
func (can *candidates) addMaybe(c *candidate) bool {
	can.mu.Lock()
	defer can.mu.Unlock()
	if can.maxAge.seconds > 0 && c.age.gt(can.maxAge, true) { // 1)
		return false // Discard
	}

	if can.maxEntries == 0 || len(can.cf) < can.maxEntries { // 2)
		can.cf = append(can.cf, c) // Add
		can.setOldest(len(can.cf) - 1)
		return true
	}

	// 3) Is candidate younger than oldest?
	if !c.age.lt(can.cf[can.oldest].age) {
		return false // No
	}

	// Candidate is younger so evict oldest by replacement.
	can.cf[can.oldest] = c

	// Eviction invalidates can.oldest so re-established by linear search.
	can.oldest = 0 // default to first
	for ix, c := range can.cf {
		if c.age.gt(can.cf[can.oldest].age, false) {
			can.oldest = ix // then replace
		}
	}

	return true
}

// Set can.oldest based on the possibility that newIx is older than can.oldest. The reason
// for tracking oldest as that is the candidate that will be replaced in the event of an
// otherwise fill candidate set.
func (can *candidates) setOldest(newIx int) {
	c := can.cf[newIx]
	if can.cf[can.oldest].age.lt(c.age) { // If candidate is new oldest then
		can.oldest = newIx
	}
}

// maxAgeWidth returns the number of format character positions needed for the largest "age".
func (can *candidates) maxAgeWidth() (maxWidth int) {
	for _, cf := range can.cf {
		l := len(cf.age.compactString())
		if l > maxWidth {
			maxWidth = l
		}
	}

	return
}
