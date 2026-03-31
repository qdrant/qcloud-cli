package output

import (
	"fmt"
	"reflect"
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

const boolYes = "yes"

// BoolYesNo formats a bool as "yes" or "no".
func BoolYesNo(v bool) string {
	if v {
		return boolYes
	}
	return "no"
}

// BoolMark formats a bool as "yes" or empty string ("").
func BoolMark(v bool) string {
	if v {
		return boolYes
	}
	return ""
}

// OptionalValue formats an optional pointer value as a string.
// Returns fallback for nil pointers. Supports any pointer type.
// Booleans are formatted as "yes"/"no"; all other types use their default format.
func OptionalValue(v any, fallback string) string {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fallback
	}

	elem := rv.Elem().Interface()
	if b, ok := elem.(bool); ok {
		if b {
			return boolYes
		}
		return "no"
	}

	return fmt.Sprintf("%v", elem)
}
