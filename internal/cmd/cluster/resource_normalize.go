package cluster

import (
	"fmt"
	"strconv"
	"strings"
)

// normalizeMillicores converts shorthand CPU/GPU values to millicore format.
// "1" → "1000m", "0.5" → "500m", "1000m" → "1000m" (passthrough).
func normalizeMillicores(s string) (string, error) {
	if strings.HasSuffix(s, "m") {
		return s, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", fmt.Errorf("cannot parse %q as a millicore value", s)
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

// parseCPUMillicores parses a CPU string to millicores.
// "1" -> 1000, "0.5" -> 500, "1000m" -> 1000.
func parseCPUMillicores(s string) (int64, error) {
	if strings.HasSuffix(s, "m") {
		v, err := strconv.ParseInt(strings.TrimSuffix(s, "m"), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as a CPU value", s)
		}
		return v, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a CPU value", s)
	}
	return int64(v * 1000), nil
}

// parseRAMGiB parses a RAM string to GiB.
// "4GiB" -> 4, "4Gi" -> 4, "4G" -> 4, "8" -> 8 (assumes GiB).
func parseRAMGiB(s string) (int64, error) {
	for _, suffix := range []string{"GiB", "Gi", "G"} {
		if strings.HasSuffix(s, suffix) {
			v, err := strconv.ParseInt(strings.TrimSuffix(s, suffix), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot parse %q as a RAM value", s)
			}
			return v, nil
		}
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a RAM value", s)
	}
	return v, nil
}

// parseGPUMillicores parses a GPU string to millicores.
// "1000m" -> 1000, "1" -> 1000.
func parseGPUMillicores(s string) (int64, error) {
	if strings.HasSuffix(s, "m") {
		v, err := strconv.ParseInt(strings.TrimSuffix(s, "m"), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as a GPU value", s)
		}
		return v, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as a GPU value", s)
	}
	return v * 1000, nil
}

// gpuMillicoresToCount converts a millicore GPU string (e.g. "1000m") to an integer string (e.g. "1").
// Returns "" if the value cannot be parsed or is not a whole number of GPUs.
func gpuMillicoresToCount(v string) string {
	s := strings.TrimSuffix(v, "m")
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 || n%1000 != 0 {
		return ""
	}
	return strconv.Itoa(n / 1000)
}
