package resource

import (
	"fmt"
	"strconv"
	"strings"
)

// Convenience constants (binary SI).
const (
	KiB ByteQuantity = 1024
	MiB              = 1024 * KiB
	GiB              = 1024 * MiB
	TiB              = 1024 * GiB
)

// ByteQuantity stores a memory quantity as int64 bytes (the smallest unit).
// Implements pflag.Value for use with cmd.Flags().Var().
type ByteQuantity int64

// ParseByteQuantity parses a byte quantity string.
// Accepts T/Ti/TiB, G/Gi/GiB, M/Mi/MiB, K/Ki/KiB, or bare integer (treated as GiB).
func ParseByteQuantity(s string) (ByteQuantity, error) {
	// Match longest suffix first.
	suffixes := []struct {
		suffix string
		mult   ByteQuantity
	}{
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
		if strings.HasSuffix(s, e.suffix) {
			n, err := strconv.ParseInt(strings.TrimSuffix(s, e.suffix), 10, 64)
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
	units := []struct {
		mult   ByteQuantity
		suffix string
	}{
		{TiB, "TiB"},
		{GiB, "GiB"},
		{MiB, "MiB"},
		{KiB, "KiB"},
	}
	for _, u := range units {
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

// FormatByteQuantity formats b in the given binary SI unit.
// unit: "Ki"/"KiB", "Mi"/"MiB", "Gi"/"GiB", "Ti"/"TiB", "" for plain bytes.
// Example: FormatByteQuantity(8*GiB, "GiB") → "8GiB".
func FormatByteQuantity(b ByteQuantity, unit string) string {
	switch unit {
	case "Ti", "TiB":
		return fmt.Sprintf("%dTiB", int64(b/TiB))
	case "Gi", "GiB":
		return fmt.Sprintf("%dGiB", int64(b/GiB))
	case "Mi", "MiB":
		return fmt.Sprintf("%dMiB", int64(b/MiB))
	case "Ki", "KiB":
		return fmt.Sprintf("%dKiB", int64(b/KiB))
	default:
		return fmt.Sprintf("%d", int64(b))
	}
}
