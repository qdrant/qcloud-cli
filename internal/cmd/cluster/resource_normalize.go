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
