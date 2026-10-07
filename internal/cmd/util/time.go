package util

import "time"

// ParseDateEndOfDay parses a YYYY-MM-DD date and returns the last second of
// that day in UTC, so that the date is inclusive when used as an upper bound.
func ParseDateEndOfDay(s string) (time.Time, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return time.Time{}, err
	}

	return t.AddDate(0, 0, 1).Add(-time.Second), nil
}
