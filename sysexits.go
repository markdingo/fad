package main

// Define the subset of EX_* codes used by this program. Since the C standard apparently
// only specifics EXIT_SUCCESS and EXIT_FAILURE, the definition of EX_* codes used by
// FreeBSD are unlikely to be viewed as standard enough to ever get incorporated into the
// core go package. Thus this.

const (
	EX_OK     int = 0
	EX_USAGE      = 64
	EX_OSFILE     = 72
	EX_IOERR      = 74
	EX_CONFIG     = 78
)
