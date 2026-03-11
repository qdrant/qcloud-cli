package output

import (
	"time"

	"github.com/dustin/go-humanize"
)

// HumanTime returns a relative human-readable time (e.g., "3 hours ago").
// Returns empty string for zero time.
func HumanTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return humanize.Time(t)
}

// FullDateTime formats t as "2006-01-02 15:04:05 UTC".
// Returns empty string for zero time.
func FullDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05 UTC")
}
