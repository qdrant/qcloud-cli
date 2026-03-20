package resource

import (
	"fmt"
	"strconv"
	"strings"
)

// Binary SI size constants.
const (
	KiB ByteQuantity = 1024
	MiB              = 1024 * KiB
	GiB              = 1024 * MiB
	TiB              = 1024 * GiB
	PiB              = 1024 * TiB
	EiB              = 1024 * PiB
)

// Unit suffix constants for use with FormatByteQuantity.
const (
	UnitKiB = "KiB"
	UnitMiB = "MiB"
	UnitGiB = "GiB"
	UnitTiB = "TiB"
	UnitPiB = "PiB"
	UnitEiB = "EiB"
)

// ByteQuantity stores a memory quantity as int64 bytes (the smallest unit).
// Implements pflag.Value for use with cmd.Flags().Var().
type ByteQuantity int64

// ParseByteQuantity parses a byte quantity string.
// Accepts E/Ei/EiB, P/Pi/PiB, T/Ti/TiB, G/Gi/GiB, M/Mi/MiB, K/Ki/KiB, or bare integer (treated as GiB).
func ParseByteQuantity(s string) (ByteQuantity, error) {
	// Match longest suffix first.
	suffixes := []struct {
		suffix string
		mult   ByteQuantity
	}{
		{"EiB", EiB},
		{"Ei", EiB},
		{"E", EiB},
		{"PiB", PiB},
		{"Pi", PiB},
		{"P", PiB},
		{"TiB", TiB},
		{"Ti", TiB},
		{"T", TiB},
		{"GiB", GiB},
		{"Gi", GiB},
		{"G", GiB},
		{"MiB", MiB},
		{"Mi", MiB},
		{"M", MiB},
		{"KiB", KiB},
		{"Ki", KiB},
		{"K", KiB},
	}
	for _, e := range suffixes {
		if q, ok := strings.CutSuffix(s, e.suffix); ok {
			n, err := strconv.ParseInt(q, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot parse %q as a byte quantity", s)
			}

			return ByteQuantity(n) * e.mult, nil
		}
	}

	// Bare integer: treat as GiB.
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a byte quantity", s)
	}
	return ByteQuantity(n) * GiB, nil
}

// Set implements pflag.Value.
func (b *ByteQuantity) Set(s string) error {
	v, err := ParseByteQuantity(s)
	if err != nil {
		return err
	}
	*b = v
	return nil
}

// String implements pflag.Value.
// Auto-picks the largest binary SI unit that divides evenly: 8*GiB → "8GiB". Returns "" for zero.
func (b ByteQuantity) String() string {
	if b == 0 {
		return ""
	}
	for _, u := range unitTable {
		if b%u.mult == 0 {
			return fmt.Sprintf("%d%s", int64(b/u.mult), u.suffix)
		}
	}
	return fmt.Sprintf("%d", int64(b))
}

// Type implements pflag.Value.
func (b *ByteQuantity) Type() string {
	return "bytes"
}

// GiB returns the quantity as an integer number of gibibytes (truncating).
func (b ByteQuantity) GiB() int64 {
	return int64(b) / int64(GiB)
}

// FormatByteQuantity formats b in the given binary SI unit (use the UnitXiB constants).
// "" formats as plain bytes.
// Example: FormatByteQuantity(8*GiB, UnitGiB) → "8GiB".
func FormatByteQuantity(b ByteQuantity, unit string) string {
	for _, u := range unitTable {
		if unit == u.suffix {
			return fmt.Sprintf("%d%s", int64(b/u.mult), u.suffix)
		}
	}
	return fmt.Sprintf("%d", int64(b))
}

// unitTable lists all binary SI units from largest to smallest.
// Used by String() and FormatByteQuantity().
var unitTable = []struct {
	mult   ByteQuantity
	suffix string
}{
	{EiB, UnitEiB},
	{PiB, UnitPiB},
	{TiB, UnitTiB},
	{GiB, UnitGiB},
	{MiB, UnitMiB},
	{KiB, UnitKiB},
}
