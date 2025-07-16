package main

import (
	"fmt"
	"io"
	"path/filepath"
	"time"
)

func (scn *scanner) printCandidates(out io.Writer) {
	fmtString := fmt.Sprintf("%%%ds:%%s:%%s\n", scn.allCandidates.maxAgeWidth()) // Determine age format
	for _, cf := range scn.allCandidates.cf {
		p := cf.path
		p = filepath.Clean(p) // Trim off any leading "./" or ".\" or whatever the OS prefers
		fType := cf.fType()
		if scn.cfg.printDirname.v { // If printing just dirname(path) then
			p = filepath.Dir(p) // Trim path
			fType = "d"         // and force type
		}
		fmt.Fprintf(out, fmtString, cf.age.compactString(), fType, p)
	}
}

func (scn *scanner) printStats(out io.Writer, secs time.Duration) {
	fmt.Fprintf(out, "Elapse: %0.1fs %d/%d Found: %d Dirs: %d Files: %d Others: %d Ignored: %d Errors: %d\n",
		secs.Seconds()+0.05,
		scn.cc.limit-scn.cc.minimum, scn.cc.limit,
		len(scn.allCandidates.cf),
		scn.dirCount, scn.fileCount, scn.otherCount, scn.ignoreCount, scn.errorCount)
}
