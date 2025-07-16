package main

import (
	"io/fs"
	"testing"
	"time"
)

func TestCandidateSet(t *testing.T) {
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	dtm := now.Add(-1 * time.Second)
	var c candidate
	c.set("/home/candidate", fs.ModeDevice, now, dtm)
	if c.isDir() || c.path != "/home/candidate" || c.age.seconds != 1 {
		t.Error("Candidate not set", c.isDir(), c)
	}

	if !c.isSet() {
		t.Error("Candidate should be set")
	}
}

func TestCandidatesSortAscending(t *testing.T) {
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	can := newCandidates(10, age{})
	var c0, c1, c2 candidate
	c0t := now.Add(-2 * time.Second)
	c1t := now.Add(-5 * time.Second)
	c2t := now.Add(-10 * time.Second)
	c0.set("c0", fs.ModeDir, now, c0t)
	c1.set("c1", fs.ModeDir, now, c1t)
	c2.set("c2", fs.ModeDir, now, c2t)
	if !can.addMaybe(&c2) {
		t.Fatal("Add failed unexpectedly for c2")
	}
	if !can.addMaybe(&c1) {
		t.Fatal("Add failed unexpectedly for c1")
	}
	if !can.addMaybe(&c0) {
		t.Fatal("Add failed unexpectedly for c0")
	}
	can.sortAscending() // Starts as c2, c1, c0 should now be c0, c1, c2
	if can.cf[0] != &c0 || can.cf[1] != &c1 || can.cf[2] != &c2 {
		t.Error("Sort failed", can.cf[0], can.cf[1], can.cf[2])
	}
}

func TestCandidatesAddMaybe(t *testing.T) {
	const maxAge = 1000
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	can := newCandidates(2, age{seconds: maxAge})
	var c0, c1, c2, c3, old candidate
	c0t := now.Add(-2 * time.Second)
	c1t := now.Add(-5 * time.Second)
	c2t := now.Add(-10 * time.Second)            // c2 is oldest valid
	oldt := now.Add(-(maxAge + 1) * time.Second) // Should be too old for maxAge

	c0.set("c0", fs.ModeDir, now, c0t)
	c1.set("c1", fs.ModeDir, now, c1t)
	c2.set("c2", fs.ModeDir, now, c2t)
	old.set("old", fs.ModeDir, now, oldt)

	if can.addMaybe(&old) {
		t.Error("Too old should not have been added at all")
	}

	if !can.addMaybe(&c0) {
		t.Error("Add failed unexpectedly for c0")
	}
	if can.oldest != 0 {
		t.Error("Solitary c0 should be oldest")
	}

	if !can.addMaybe(&c1) {
		t.Error("Add failed unexpectedly for c1")
	}
	if can.oldest != 1 {
		t.Error("c1 is oldest, but can thinks", can.oldest, "of", can.cf[0], can.cf[1])
	}

	if can.addMaybe(&c2) {
		t.Error("Add succeeded unexpectedly for c2")
	}

	c3t := now.Add(-3 * time.Second) // Should evict c1
	c3.set("c3", fs.ModeDir, now, c3t)
	if !can.addMaybe(&c3) {
		t.Error(&c3, "should have evicted c1", can.cf[0], can.cf[1])
	}
	if can.cf[can.oldest] != &c3 {
		t.Error("c3 should now be oldest, not", can.cf[can.oldest])
	}
}

func TestCandidateFType(t *testing.T) {
	var c candidate
	modes := []fs.FileMode{0, fs.ModeDir, fs.ModeTemporary, fs.ModeSymlink, fs.ModeDevice, fs.ModeNamedPipe,
		fs.ModeSocket, fs.ModeCharDevice, fs.ModeIrregular}
	expect := map[string]bool{"f": true, "d": true, "T": true, "L": true, "D": true, "p": true, "S": true, "c": true, "?": true}
	for _, m := range modes {
		c.mode = fs.FileMode(m)
		delete(expect, c.fType())
	}

	if len(expect) > 0 {
		t.Error("Should not be any residual types in", expect)
	}
}

func TestCandidatesMaxAgeWidth(t *testing.T) {
	var c1, c2, c3, c4 candidate
	c1.age.seconds = 1          // "1s" = 2
	c2.age.seconds = 61         // "1m" = 2
	c3.age.seconds = 59         // "59s" = 3
	c4.age.seconds = 200 * year // "200Y" = 4

	var can candidates
	can.cf = []*candidate{&c1, &c2}
	width := can.maxAgeWidth()
	if width != 2 {
		t.Error("Expected 2, got", width)
	}
	can.cf = append(can.cf, &c3)
	width = can.maxAgeWidth()
	if width != 3 {
		t.Error("Expected 3, got", width)
	}

	can.cf = append(can.cf, &c4)
	width = can.maxAgeWidth()
	if width != 4 {
		t.Error("Expected 4, got", width)
	}
}
