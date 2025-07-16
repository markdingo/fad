//go:build darwin

package main

var (
	defaultIgnoreBasenames = []string{
		"Library",
		"Caches",
		".git",
		".emacs.d",
		"cache",
		".DS_Store",
	}
)
