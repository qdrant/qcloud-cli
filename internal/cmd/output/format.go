package output

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
)

// Duration formats d as a human-readable duration with days as the largest
// unit and at most two units (e.g. "1 day", "30 days", "1 day 12 hours").
func Duration(d time.Duration) string {
	return durafmt.Parse(d).LimitToUnit("days").LimitFirstN(2).String()
}

// CompactDuration formats d using Go duration notation without trailing zero
// units (e.g. "2m", "1h", "24h", "1h30m").
func CompactDuration(d time.Duration) string {
	s := d.String()
	s = strings.TrimSuffix(s, "m0s")
	if len(s) != len(d.String()) {
		s += "m"
	}

	if trimmed, ok := strings.CutSuffix(s, "h0m"); ok {
		s = trimmed + "h"
	}

	return s
}

// Bytes formats a measured byte size using IEC units (e.g. "0 B", "1.5 GiB").
// Use resource.ByteQuantity for configured sizes, which must round-trip.
func Bytes(v uint64) string {
	return humanize.IBytes(v)
}

// Rate formats a per-second rate with two decimals (e.g. "1.25/s").
func Rate(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64) + "/s"
}

// Milliseconds formats a value in milliseconds with one decimal (e.g. "12.3ms").
func Milliseconds(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64) + "ms"
}

// UnlimitedIfZero returns "unlimited" when v is 0 and formatted otherwise. It is
// used for limits where 0 means no limit.
func UnlimitedIfZero(formatted string, v uint64) string {
	if v == 0 {
		return "unlimited"
	}

	return formatted
}

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

// JoinOrDefault joins items with ", ", or returns fallback when items is empty.
func JoinOrDefault(items []string, fallback string) string {
	if len(items) == 0 {
		return fallback
	}

	return strings.Join(items, ", ")
}
