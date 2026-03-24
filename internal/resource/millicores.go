package resource

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Millicores stores a CPU/GPU quantity as int64 millicores (the smallest unit).
// Implements pflag.Value for use with cmd.Flags().Var().
type Millicores int64

// ParseMillicores parses a millicore string.
// "1" → 1000, "0.5" → 500, "1000m" → 1000.
func ParseMillicores(s string) (Millicores, error) {
	if rv, ok := strings.CutSuffix(s, "m"); ok {
		v, err := strconv.ParseInt(rv, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as a millicore value", s)
		}

		return Millicores(v), nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a millicore value", s)
	}

	return Millicores(math.Round(v * 1000)), nil
}

// Set implements pflag.Value.
func (m *Millicores) Set(s string) error {
	v, err := ParseMillicores(s)
	if err != nil {
		return err
	}
	*m = v
	return nil
}

// String implements pflag.Value.
// Auto-picks the best decimal SI unit: 1000 → "1", 500 → "500m". Returns "" for zero.
func (m Millicores) String() string {
	if m == 0 {
		return ""
	}
	if int64(m)%1000 == 0 {
		return strconv.FormatInt(int64(m)/1000, 10)
	}
	return fmt.Sprintf("%dm", int64(m))
}

// Type implements pflag.Value.
func (m *Millicores) Type() string {
	return "millicores"
}

// FormatMillicores formats m in the given decimal SI unit prefix.
// unit: "m" (millicores), "" (whole cores).
// Example: FormatMillicores(1000, "m") → "1000m", FormatMillicores(1000, "") → "1".
func FormatMillicores(m Millicores, unit string) string {
	switch unit {
	case "m":
		return fmt.Sprintf("%dm", int64(m))
	default:
		return strconv.FormatInt(int64(m)/1000, 10)
	}
}
