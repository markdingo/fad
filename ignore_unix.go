//go:build unix && !darwin

package main

var (
	defaultIgnoreBasenames = []string{
		".git",
		".emacs.d",
		"cache",
		".cache",
	}
)
