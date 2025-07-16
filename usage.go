package main

import (
	"flag"
	"fmt"
	"io"
)

func printUsage(stdout, stderr io.Writer, hasError bool, fs *flag.FlagSet, cfg *config) {
	var out io.Writer
	if hasError {
		out = stderr
		fmt.Fprintln(out) // Separate flag error from our usage
	} else {
		out = stdout
	}
	fs.SetOutput(out)

	fmt.Fprintf(out,
		`NAME
  %s - find active directories

SYNOPSIS
  %s [options] [path...]

DESCRIPTION
  %s recursively searches each path to find directories with the most recent
  'activity date'. The 'activity date' is derived from the date-time-modified
  of the most recently modified entry within each directory. The default path
  is the current working directory.

  %s differs from 'find -mtime' and 'find -newer' in that it compares the
  activity *within* each directory as opposed to comparing the directory's
  date-time-modified against a fixed age or file.

  Output consists of directory details and their most recently modified
  file in 'activity date' order.  The number of directories listed is
  controlled by --count and --age.

OPTIONS
`, Name, Name, Name, Name)
	fs.PrintDefaults()
	fmt.Fprintln(out)
	fmt.Fprint(out,
		`All 'Ignore paths' values are comma-strings allowing multiple values separated
by commas. To add to defaults instead of replacing them, prefix with '+' such
as "-ibase +.ssh,.local".

The -itypes values can be `, validFTypesString, ` as described in the manpage.

`)
	fmt.Fprintln(out, "Config:", cfg.configPathHelp)
	printVersion(out)
}

func printVersion(out io.Writer) {
	fmt.Fprintln(out, "Project:", Project)
	fmt.Fprintln(out, "Version:", Version)
	fmt.Fprintln(out, "Release:", ReleaseDate)
}
