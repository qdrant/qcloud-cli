package cluster

import (
	"fmt"
	"strconv"
	"strings"
)

// normalizeCPU converts shorthand CPU values to millicore format.
// "1" → "1000m", "0.5" → "500m", "1000m" → "1000m" (passthrough).
func normalizeCPU(s string) (string, error) {
	if strings.HasSuffix(s, "m") {
		return s, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", fmt.Errorf("cannot parse %q as a CPU value", s)
	}
	return fmt.Sprintf("%.0fm", v*1000), nil
}

// parseDiskGiB parses a disk string to a uint32 GiB value.
// Accepts "100GiB", "100Gi", "100G", "100" (all → 100),
// and "1TiB", "1Ti", "1T" (all → 1024).
func parseDiskGiB(s string) (uint32, error) {
	for _, suffix := range []string{"TiB", "Ti", "T"} {
		if strings.HasSuffix(s, suffix) {
			v, err := strconv.ParseUint(strings.TrimSuffix(s, suffix), 10, 32)
			if err != nil {
				return 0, fmt.Errorf("cannot parse %q as a disk value", s)
			}
			return uint32(v) * 1024, nil
		}
	}
	for _, suffix := range []string{"GiB", "Gi", "G"} {
		if strings.HasSuffix(s, suffix) {
			s = strings.TrimSuffix(s, suffix)
			break
		}
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a disk value", s)
	}
	return uint32(v), nil
}

// normalizeRAM converts shorthand RAM values to GiB binary notation.
// "8" → "8GiB", "8G" → "8GiB", "8Gi" → "8GiB", "8GiB" → "8GiB" (passthrough).
func normalizeRAM(s string) (string, error) {
	if strings.HasSuffix(s, "GiB") {
		return s, nil
	}
	if strings.HasSuffix(s, "Gi") {
		return strings.TrimSuffix(s, "Gi") + "GiB", nil
	}
	if strings.HasSuffix(s, "G") {
		return strings.TrimSuffix(s, "G") + "GiB", nil
	}
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", fmt.Errorf("cannot parse %q as a RAM value", s)
	}
	return s + "GiB", nil
}
