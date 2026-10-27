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
// Formats as (fractional) cores: 1000 → "1", 500 → "0.5". Returns "" for zero.
func (m Millicores) String() string {
	if m == 0 {
		return ""
	}

	return m.Cores()
}

// Cores formats m as a (fractional) number of cores: 1000 → "1", 1500 → "1.5", 0 → "0".
func (m Millicores) Cores() string {
	return strconv.FormatFloat(float64(m)/1000, 'f', -1, 64)
}

// Type implements pflag.Value.
func (m *Millicores) Type() string {
	return "cores"
}
