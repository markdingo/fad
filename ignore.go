package main

import (
	"path/filepath"
	"strings"
)

func (cfg *config) ignore(path string) string {
	if cfg.matchesBases(path) {
		return "bases"
	}

	if cfg.matchesContains(path) {
		return "contains"
	}

	if cfg.matchesRegexes(path) {
		return "regexes"
	}

	return ""
}

func (cfg *config) matchesBases(path string) bool {
	base := filepath.Base(path)
	_, ok := cfg.ignoreBasesMap[base]

	return ok
}

func (cfg *config) matchesContains(path string) bool {
	for _, s := range cfg.ignoreContainsList {
		if strings.Contains(strings.ToUpper(path), strings.ToUpper(s)) {
			return true
		}
	}

	return false
}

func (cfg *config) matchesRegexes(path string) bool {
	for _, re := range cfg.ignoreRegexesCompiled {
		if re.MatchString(path) {
			return true
		}
	}

	return false
}
