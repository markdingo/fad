package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	defaultUserFile     = "defaults.conf"
	defaultPrintLimit   = 23  // Fits nicely on a modern 24x80 Uniscope 100
	defaultScannerLimit = 10  // A semi-empirical, gut-feel guesstimate
	defaultIgnoreTypes  = "d" // Considering directories as "active" is a two-edged sword
	commaDelimiter      = "," // Comma-Strings split/joined on this character
	plusAppend          = '+' // If a comma-string starts with this, append rather than replace
)

const (
	fTypeDevice     = "D"
	fTypeSymlink    = "L"
	fTypeSocket     = "S"
	fTypeTemporary  = "T"
	fTypeCharDevice = "c"
	fTypeDir        = "d"
	fTypeFile       = "f"
	fTypeNamedPipe  = "p"
	fTypeUnknown    = "?"
)

var (
	validFTypes = map[string]fs.FileMode{ // Valid File Types
		fTypeDevice:     fs.ModeDevice,
		fTypeSymlink:    fs.ModeSymlink,
		fTypeSocket:     fs.ModeSocket,
		fTypeTemporary:  fs.ModeTemporary,
		fTypeCharDevice: fs.ModeCharDevice,
		fTypeDir:        fs.ModeDir,
		fTypeFile:       0, // Regular file
		fTypeNamedPipe:  fs.ModeNamedPipe,
	}
	validFTypesString string
)

func init() {
	var ar []string
	for k := range validFTypes {
		ar = append(ar, k)
	}
	sort.Strings(ar)
	last := ar[len(ar)-1]
	ar = ar[:len(ar)-1]
	validFTypesString = strings.Join(ar, ",") + " or " + last
}

func fTypeString(mode fs.FileMode) string {
	switch {
	case (mode & fs.ModeTemporary) != 0: // A Plan 9 thing - must preceed IsRegular() case
		return fTypeTemporary
	case (mode & fs.ModeSymlink) != 0:
		return fTypeSymlink
	case (mode & fs.ModeDevice) != 0:
		return fTypeDevice
	case (mode & fs.ModeNamedPipe) != 0:
		return fTypeNamedPipe
	case (mode & fs.ModeSocket) != 0:
		return fTypeSocket
	case (mode & fs.ModeCharDevice) != 0:
		return fTypeCharDevice
	case mode.IsRegular(): // This subsumes ModeTemporary on Unix - don't let it
		return fTypeFile
	case mode.IsDir():
		return fTypeDir
	}

	return fTypeUnknown
}

// docFlags are only valid on the command-line and result in a documentation printout of
// somesort followed by an exit.
type docFlags struct {
	help    bool
	manpage bool
	version bool
}

// configFlags can be changed by user configuration or command-line flags
type configFlags struct {
	printDirname boolFlag // Print just the dirname of the path
	printIgnored boolFlag // Print file system objects ignored by ignore filters
	printStats   boolFlag // Print scanning stats at end of program

	suppressErrors boolFlag // Don't print errors if file-system access fails

	maxAge      ageFlag  // Age limit of paths to print
	maxCount    uintFlag // How many paths to print
	maxDepth    uintFlag // Descend depth
	maxScanners uintFlag // Maximum number of concurrent directory scanners

	ignoreBases    commaStringFlag // Exact `basename` values to ignore
	ignoreContains commaStringFlag // Caseless strings to ignore in full path
	ignoreRegexes  commaStringFlag // Regexes to ignore in full path
	ignoreTypes    commaStringFlag // Ignore file system types base on our notation (validFTypes)
}

// derivedConfig values are built from configFlags
type derivedConfig struct {
	ignoreBasesMap        map[string]any
	ignoreContainsList    []string
	ignoreRegexesList     []string
	ignoreRegexesCompiled []*regexp.Regexp
	ignoreTypesMap        map[string]any
}

// userConfigDirFunc defines the function which returns the location of the default
// configuration directory. Defaults to os.UserConfigDir but is replaced by tests.
type userConfigDirFunc func() (string, error)

type config struct {
	confFunc       userConfigDirFunc
	configPathHelp string // Only used by -h

	flagSet *flag.FlagSet

	docFlags    // Command-line only
	configFlags // Can be in config file or command-line
	derivedConfig
}

// newConfig constructs a skeletal config struct and determines the path of the user
// default configuration file.
func newConfig(fs *flag.FlagSet, confFunc userConfigDirFunc) *config {
	cfg := &config{flagSet: fs, confFunc: confFunc}
	cfg.ignoreBasesMap = make(map[string]any)
	cfg.ignoreTypesMap = make(map[string]any)

	return cfg
}

func (cfg *config) setFlags() {
	cfg.flagSet.BoolVar(&cfg.help, "h", false, "Print usage, defaults, version info, then exit")
	cfg.flagSet.BoolVar(&cfg.help, "help", false, "Print usage, defaults, version info, then exit")
	cfg.flagSet.BoolVar(&cfg.manpage, "manpage", false, "Print manpage and exit - perhaps pipe into mandoc(1)")
	cfg.flagSet.BoolVar(&cfg.version, "v", false, "Print version details and exit")
	cfg.flagSet.BoolVar(&cfg.version, "version", false, "Print version details and exit")

	cfg.flagSet.Var(&cfg.maxAge, "age",
		"Print paths no older than value (e.g: 1s, 2h, 3d, 4w, 5y)")
	cfg.flagSet.Var(&cfg.maxCount, "count", "Maximum paths to print")
	cfg.flagSet.Var(&cfg.maxDepth, "depth",
		"Maximum depth to descend below command line paths (default of 0 is unlimited)")

	cfg.flagSet.Var(&cfg.ignoreBases, "ibases", "Ignore paths which matching 'basename'")
	cfg.flagSet.Var(&cfg.ignoreContains, "icontains",
		"Ignore paths containing case-insensistive string ('"+string(os.PathSeparator)+"' allowed)")
	cfg.flagSet.Var(&cfg.ignoreRegexes, "iregexes",
		"Ignore paths matching patterns (see regexp.MatchString())")
	cfg.flagSet.Var(&cfg.ignoreTypes, "itypes", "Ignore file system types")

	cfg.flagSet.Var(&cfg.printDirname, "pdirname", "Print just the 'dirname' of paths")
	cfg.flagSet.Var(&cfg.printIgnored, "pignored", "Print paths ignored by filters")
	cfg.flagSet.Var(&cfg.printStats, "pstats", "Print summary statistics")

	cfg.flagSet.Var(&cfg.suppressErrors, "q", "Suppress error messages when file-system access fails")
	cfg.flagSet.Var(&cfg.maxScanners, "scanners", "Number directories to scan concurrently")
}

// Set values which have not been previously set by caller
func (cfg *config) setInternalDefaults() {
	if cfg.maxCount.v == 0 {
		cfg.maxCount.v = defaultPrintLimit
	}
	if cfg.maxCount.min == 0 {
		cfg.maxCount.min = 1
	}
	if cfg.maxScanners.v == 0 {
		cfg.maxScanners.v = defaultScannerLimit
	}
	if cfg.maxScanners.min == 0 {
		cfg.maxScanners.min = 1
	}

	if len(cfg.ignoreBases.v) == 0 {
		cfg.ignoreBases.v = strings.Join(defaultIgnoreBasenames, commaDelimiter)
	}
	if len(cfg.ignoreTypes.v) == 0 {
		cfg.ignoreTypes.v = defaultIgnoreTypes
	}
}

// loadDefaults loads the default values from the user-provided config file. If the config
// file does not exist, that's not considered an error.
//
// If present, the config file contains lines of text with with each line contains a flag
// name (sans '-') and the value to use. The value overrides the compiled in default value
// for the corresponding flag excepting in the case of comma-strings which are appended if
// they are prefixed with '+'.
//
// Duplicate flag names are an error. Whitespace-only lines are ignored. Text beyond the
// comment-delimiter of '#' is ignored. There is no quoting mechanism nor
// line-continuation support.
//
// Unlike command-line options, bools must be supplied with a true/false argument. This is
// an arbitrary decision made the author as the visual of an isolated option seems
// ambiguous.
//
// In all cases, cfg.configPathHelp is set to something useful for -h to print out.
func (cfg *config) loadDefaults() error {
	cfg.setInternalDefaults()

	dir, err := cfg.confFunc()
	if err != nil {
		cfg.configPathHelp = err.Error() // For -h
		return err                       // This is serious enough to warrant returning
	}
	if len(dir) == 0 { // Not sure this can occur in real-life, but treat as not existing
		cfg.configPathHelp = "No UserConfigDir configure for this user"
		return nil
	}
	path := filepath.Join(dir, Name, defaultUserFile)
	cfg.configPathHelp = path

	configText, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.configPathHelp += " (not present)"
			return nil // Not existing is not really an error
		}
		cfg.configPathHelp += " " + err.Error()
		return err // But all other errors are real errors
	}

	validOptions := map[string]flagValue{ // Listed in same order as configFlags
		"pdirname": &cfg.printDirname,
		"pignored": &cfg.printIgnored,
		"pstats":   &cfg.printStats,

		"q": &cfg.suppressErrors,

		"age":      &cfg.maxAge,
		"count":    &cfg.maxCount,
		"depth":    &cfg.maxDepth,
		"scanners": &cfg.maxScanners,

		"ibases":    &cfg.ignoreBases,
		"icontains": &cfg.ignoreContains,
		"iregexes":  &cfg.ignoreRegexes,
		"itypes":    &cfg.ignoreTypes,
	}

	// Parse config file
	dupes := make(map[string]any)
	for lno, line := range strings.Split(string(configText), "\n") {
		line, _, _ = strings.Cut(line, "#")
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		option := fields[0]
		args := fields[1:]

		fv, ok := validOptions[option]
		if !ok {
			return fmt.Errorf("Unknown option '%s' at %s:%d",
				option, cfg.configPathHelp, lno+1)
		}
		if _, ok := dupes[option]; ok {
			return fmt.Errorf("Duplicate option '%s' at %s:%d",
				option, cfg.configPathHelp, lno+1)
		}
		dupes[option] = true

		if len(args) != 1 {
			return fmt.Errorf("Option '%s' expects one argument, not %d at %s:%d",
				option, len(args), cfg.configPathHelp, lno+1)
		}

		err := fv.Set(args[0])
		if err != nil {
			return err
		}
	}

	return nil
}

// Determine derived values from base config values. Generally be tolerant of "errors"
// which make no semantic difference, such as duplicates - caseless or otherwise.
func (cfg *config) compile() error {
	if len(cfg.ignoreBases.v) > 0 { // Split Idiosyncrasies
		for _, f := range strings.Split(cfg.ignoreBases.v, commaDelimiter) {
			cfg.ignoreBasesMap[f] = true
		}
	}

	if len(cfg.ignoreContains.v) > 0 {
		cfg.ignoreContainsList = strings.Split(cfg.ignoreContains.v, commaDelimiter)
	}

	if len(cfg.ignoreRegexes.v) > 0 {
		cfg.ignoreRegexesList = strings.Split(cfg.ignoreRegexes.v, commaDelimiter)
		for _, res := range cfg.ignoreRegexesList {
			re, err := regexp.Compile(res)
			if err != nil {
				return fmt.Errorf("Error: -iregexes '%s' does not compile: %w",
					res, err)
			}
			cfg.ignoreRegexesCompiled = append(cfg.ignoreRegexesCompiled, re)
		}
	}

	if len(cfg.ignoreTypes.v) > 0 {
		for _, f := range strings.Split(cfg.ignoreTypes.v, commaDelimiter) {
			if _, ok := validFTypes[f]; !ok {
				return fmt.Errorf("Error: -itypes '%s' is not one of '%s'",
					f, validFTypesString)
			}
			cfg.ignoreTypesMap[f] = true
		}
	}

	return nil
}
