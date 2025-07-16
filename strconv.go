package main

import (
	"errors"
	"strings"
)

// The errors returned by strconv and flag result in a lot of duplicate info presented to
// the user so this function trims up the strconv error. It assumes a standard strconv
// error syntax of "function: Action: error message" thus we extract the third colon
// delimited value and turn that into an error. However, given that "Action" could contain
// a colon, what we actually do is select the last colon delimited value. If the syntax is
// not as expected, return the whole error as at worst it makes for a verbose error
// message.
func strconvTrimError(err error) error {
	es := strings.Split(err.Error(), ":")
	if len(es) >= 3 {
		return errors.New(strings.TrimSpace(es[len(es)-1]))
	}

	return err
}
