package util

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// errInvalidTimeFlag describes the formats accepted by ParseTimeFlag.
var errInvalidTimeFlag = errors.New("must be RFC3339, YYYY-MM-DD, or a duration ago such as 30m, 24h or 7d")

// ParseTimeFlag parses a time flag value used as a lower bound. It accepts an
// RFC3339 timestamp, a date in YYYY-MM-DD format (start of the day in UTC), or a
// positive duration that is subtracted from now: any Go duration (e.g. "90m",
// "24h") or a number of days with a "d" suffix (e.g. "7d").
func ParseTimeFlag(s string, now time.Time) (time.Time, error) {
	return parseTimeFlag(s, now, false)
}

// ParseUntilFlag is like ParseTimeFlag but for upper bounds: a YYYY-MM-DD date
// resolves to the end of that day, so that the whole day is included.
func ParseUntilFlag(s string, now time.Time) (time.Time, error) {
	return parseTimeFlag(s, now, true)
}

func parseTimeFlag(s string, now time.Time, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	if endOfDay {
		if t, err := ParseDateEndOfDay(s); err == nil {
			return t, nil
		}
	} else if t, err := time.Parse(time.DateOnly, s); err == nil {
		return t, nil
	}

	d, err := parseAgo(s)
	if err != nil {
		return time.Time{}, errInvalidTimeFlag
	}

	return now.Add(-d), nil
}

// parseAgo parses a positive Go duration or a whole number of days ("7d").
func parseAgo(s string) (time.Duration, error) {
	if days, ok := strings.CutSuffix(s, "d"); ok {
		n, err := strconv.ParseUint(days, 10, 32)
		if err != nil || n == 0 {
			return 0, errInvalidTimeFlag
		}

		return time.Duration(n) * 24 * time.Hour, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, errInvalidTimeFlag
	}

	return d, nil
}
