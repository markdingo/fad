//go:build !darwin && !unix && !windows

package main

const (
	defaultIgnoreBasenames = []string{
		".git",
		".emacs.d",
		"cache",
	}
)
