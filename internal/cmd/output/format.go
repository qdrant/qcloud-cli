package output

import (
	"fmt"
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

// DiffValue formats a field value as "old => new" when the value changes, or just "val" when unchanged.
func DiffValue(oldVal, newVal string) string {
	if oldVal == newVal {
		return newVal
	}
	return oldVal + " => " + newVal
}

// OptionalValue formats an optional pointer value as a string.
// Returns fallback for nil pointers. Supports *uint32, *int32, and *bool.
func OptionalValue(v any, fallback string) string {
	switch val := v.(type) {
	case *uint32:
		if val == nil {
			return fallback
		}
		return fmt.Sprintf("%d", *val)
	case *int32:
		if val == nil {
			return fallback
		}
		return fmt.Sprintf("%d", *val)
	case *bool:
		if val == nil {
			return fallback
		}
		if *val {
			return "yes"
		}
		return "no"
	default:
		return fallback
	}
}
