package main

import (
	"bytes"
	"testing"
	"time"
)

func TestPrintCandidates(t *testing.T) {
	var out bytes.Buffer
	var c1, c2, c3, c4 candidate
	c1.path = "./TODO"
	c1.age.seconds = 1
	c2.path = "/var/log/system"
	c2.age.seconds = 61
	c3.path = "/usr/local/etc/postfix/master.cf"
	c3.age.seconds = 59
	c4.path = "/usr/local/etc/nsd/nsd.conf"
	c4.age.seconds = 200 * year
	var scn scanner
	scn.cfg = &config{}
	var can candidates
	can.cf = []*candidate{&c1, &c2, &c3, &c4}
	scn.allCandidates = &can
	scn.allCandidates.sortAscending()

	scn.printCandidates(&out)
	exp := `  1s:f:TODO
 59s:f:/usr/local/etc/postfix/master.cf
  1m:f:/var/log/system
200Y:f:/usr/local/etc/nsd/nsd.conf
`
	got := out.String()
	if got != exp {
		t.Error("Print mismatch. Got\n", got, "Exp\n", exp)
	}

	scn.cfg.printDirname.v = true
	out.Reset()
	scn.printCandidates(&out)
	exp = `  1s:d:.
 59s:d:/usr/local/etc/postfix
  1m:d:/var/log
200Y:d:/usr/local/etc/nsd
`
	got = out.String()
	if got != exp {
		t.Error("Print mismatch. Got\n", got, "Exp\n", exp)
	}
}

func TestPrintStats(t *testing.T) {
	var out bytes.Buffer
	var cfg config
	var base time.Time
	allCandidates := newCandidates(10, age{})
	cc := newConcurrencyController(5)
	scn := newScanner(&cfg, cc, allCandidates, base, &out)
	scn.printStats(&out, time.Second*2)
	exp := "Elapse: 2.0s 0/5 Found: 0 Dirs: 0 Files: 0 Others: 0 Ignored: 0 Errors: 0\n"
	got := out.String()
	if got != exp {
		t.Error("Print mismatch. Got\n", got, "Exp\n", exp)
	}

	cc.minimum = 1
	var c1 candidate
	scn.allCandidates.cf = append(scn.allCandidates.cf, &c1)
	scn.dirCount = 2
	scn.fileCount = 3
	scn.otherCount = 4
	scn.ignoreCount = 5
	scn.errorCount = 6
	out.Reset()
	scn.printStats(&out, time.Minute*2)
	exp = "Elapse: 120.0s 4/5 Found: 1 Dirs: 2 Files: 3 Others: 4 Ignored: 5 Errors: 6\n"
	got = out.String()
	if got != exp {
		t.Error("Print mismatch. Got\n", got, "Exp\n", exp)
	}

}
