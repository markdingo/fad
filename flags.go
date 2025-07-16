package main

import (
	"fmt"
	"strconv"
)

// Our version of flags more or less mimics flag.Values but allows us to set them up with
// defaults from the user config file prior to adding them into the flagSet. There is also
// additional validating in some cases.

type flagValue interface {
	String() string
	Set(string) error
}

// Bool
type boolFlag struct {
	v bool
}

func (bo *boolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return strconvTrimError(err)
	}
	bo.v = v

	return nil
}

func (bo *boolFlag) String() string { return strconv.FormatBool(bo.v) }

// The existence of IsBoolFlag  tells the "flag" package that no value is expected.
func (bo *boolFlag) IsBoolFlag() bool { return true }

// Uint
type uintFlag struct {
	v   uint
	min uint // min > 0 && v < min -> error
}

func (io *uintFlag) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		return strconvTrimError(err)
	}
	if io.min > 0 && uint(v) < io.min {
		return fmt.Errorf("Less than minimum of '%d'", io.min)
	}

	io.v = uint(v)

	return nil
}

// String
func (io *uintFlag) String() string { return strconv.Itoa(int(io.v)) }

type commaStringFlag struct {
	v string
}

// Set is an append if the string starts with a "+" or a replace otherwise.
func (csf *commaStringFlag) Set(s string) error {
	if len(s) == 0 || s[0] != plusAppend { // Empty or not "+", replace
		csf.v = s
		return nil
	}

	// Must be a +string
	if len(csf.v) == 0 { // Append to an empty string removes "+"
		csf.v = s[1:]
	} else {
		csf.v += "," + s[1:] // Otherwise replace "+" with ","
	}

	return nil
}

func (csf *commaStringFlag) String() string { return csf.v }

// Age
type ageFlag = age
