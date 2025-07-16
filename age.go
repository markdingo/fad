package main

import (
	"fmt"
	"strconv"
	"time"
)

const (
	second        int64 = 1
	minute              = 60 * second
	hour                = 60 * minute
	day                 = 24 * hour
	week                = 7 * day
	year                = (365 * day) + (5 * hour) + (49 * minute) + (12 * second) // Gregorian
	month               = year / 12
	ageMaxSeconds       = 999 * year // Arbitrary, but pretty value
)

var (
	ageUnitToSeconds = map[string]int64{
		"s": second,
		"m": minute,
		"h": hour,
		"D": day,
		"W": week,
		"M": month,
		"Y": year,
	}

	ageSecondsToUnit = []struct {
		upper int64
		unit  string
	}{
		{year, "Y"},   // Year
		{month, "M"},  // Month
		{week, "W"},   // Week
		{day, "D"},    // Day
		{hour, "h"},   // hour
		{minute, "m"}, // minute
	}
)

// age implements flag.Value as an alternative to flag.Duration because the latter is too
// fine-grained. age accepts values that are more typical of what people care about for
// files such as days, hours and weeks since modified. The calculation for seconds is:
// time.Now().Sub(candidateTime).Seconds() which will normally be positive for anything in
// the file system since it will normally have a creation time earlier than "now".
//
// -ve <-----0-----> +ve
// Young             Old
type age struct {
	seconds    int64  // Seconds before "now" - normally positive for file-system objects
	multiplier int64  // Based on unit or defaults to 1 if zero
	value      string // Original Set value as a string - only used during flags processing
}

func (a *age) String() string {
	return a.value
}

// setFromTime stores the distance from 'now' to the value 'v'.
func (a *age) setFromTime(now, v time.Time) {
	a.seconds = int64(now.Sub(v).Seconds())
}

// gt returns true if 'a' is greater than 'u' - that is, older. Equal ages are not greater
// than.
//
// If useMultiplier is true, truncate the ages to the multiplier value. The idea is that a
// 3W value equals a 3.2W value and is thus not greater than. The goal of useMultiplier is
// to ensure that a comparison of ages which produce identical compactString() values
// result in a not greater than outcome.
func (a *age) gt(u age, truncateToMultiplier bool) bool {
	var multiplier int64
	multiplier = second // Default multiplier
	if truncateToMultiplier {
		if u.multiplier > 0 {
			multiplier = u.multiplier
		}
		if a.multiplier > 0 {
			multiplier = a.multiplier
		}
	}
	aSecs := (a.seconds + multiplier/2) / multiplier
	uSecs := (u.seconds + multiplier/2) / multiplier

	return aSecs > uSecs
}

// le returns true if 'a' is less than or equal to 'u'. Use this when you want equal age
// comparisons to favour the receiver age.
func (a *age) le(u age) bool {
	return a.seconds <= u.seconds
}

// lt returns true if 'a' is less than 'u' - that is, younger. Equal ages are not younger.
func (a *age) lt(u age) bool {
	return a.seconds < u.seconds
}

// Set helps meets the flag.Value interface. It parses and sets the age or rejects with an
// error. Parsing only accepts 1-5 decimal digits followed by a single unit character. No
// leading '0' is allowed to thwart old-timers who might try to sneak in an octal value
// reminiscent of their mainframe era.
//
// The assumption is that age is being set relative to now and thus is stored as a
// positive integer.
func (a *age) Set(s string) (err error) {
	if len(s) > 6 || len(s) < 2 {
		return fmt.Errorf("Age '%s' must be 1-5 digits + unit", s)
	}
	l := len(s)
	if s[0] == '0' && l > 2 {
		return fmt.Errorf("First digit of '%s' cannot be zero (we no grok octal)", s)
	}

	unit := string(s[l-1])
	l-- // Ensure scanner stops prior to unit

	// Check multiplier
	var ok bool
	if a.multiplier, ok = ageUnitToSeconds[unit]; !ok {
		return fmt.Errorf("Age '%s' has invalid unit '%s' - expect s,m,h,D,W,M or Y",
			s, unit)
	}

	// Convert decimal string to binary. We could use strconv here...
	val, err := strconv.ParseInt(s[:l], 10, 64)
	if err != nil {
		return strconvTrimError(err)
	}

	/*
		var val int64
		var ix int
		for ; ix < l; ix++ {
			if s[ix] < '0' || s[ix] > '9' {
				return fmt.Errorf("Age '%s' has invalid digit '%s'", s, string(s[ix]))
			}
			val *= 10
			val += int64(s[ix] - '0')
		}
	*/

	val *= a.multiplier // Scale out
	if val > ageMaxSeconds {
		return fmt.Errorf("Age '%s' exceeds maximum value of %d years",
			s, ageMaxSeconds/year)
	}

	a.seconds = val
	a.value = s

	return
}

// compactString return a compact string which tries to fit in 3 characters by adjusting
// the granularity as the age increases. If the age is in the future, return "fut".
//
// >= 365d -> nnY (Years)
// >= 365/12d -> nnM (Months)
// >= 7d -> nnW (Weeks)
// >= 1d -> nnD (Days)
// >= 1h -> nnh (Hours)
// >= 1m -> nnm (Minutes)
// < 60s -> nns (Seconds)
func (a *age) compactString() string {
	if a.seconds < 0 {
		return "fut"
	}
	for _, s := range ageSecondsToUnit {
		if a.seconds >= s.upper {
			return fmt.Sprintf("%d%s", (a.seconds+s.upper/2)/s.upper, s.unit)
		}
	}

	return fmt.Sprintf("%ds", a.seconds)
}
