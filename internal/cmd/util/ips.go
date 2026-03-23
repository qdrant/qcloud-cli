package util

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// IPChanges holds the parsed result of --allowed-ip flags.
type IPChanges struct {
	Add    []string
	Remove map[string]bool
}

// ParseIPs parses a slice of raw --allowed-ip flag values into additions and removals.
//
// Accepted forms:
//   - "10.0.0.0/8"  -- add the IP CIDR
//   - "10.0.0.0/8-" -- remove the IP CIDR (trailing dash)
//
// When the same IP appears multiple times, the last occurrence wins.
func ParseIPs(raw []string) (*IPChanges, error) {
	changes := &IPChanges{
		Remove: make(map[string]bool),
	}

	for _, entry := range raw {
		if entry == "" {
			return nil, fmt.Errorf("empty --allowed-ip value")
		}

		if strings.HasSuffix(entry, "-") {
			ip := entry[:len(entry)-1]
			if ip == "" {
				return nil, fmt.Errorf("empty IP in --allowed-ip %q", entry)
			}

			// Last operation wins: remove from Add if previously added.
			changes.Add = slices.DeleteFunc(changes.Add, func(s string) bool { return s == ip })
			changes.Remove[ip] = true
			continue
		}

		// Last operation wins: remove from Remove if previously marked for removal.
		delete(changes.Remove, entry)
		// Avoid duplicates in Add.
		if !slices.Contains(changes.Add, entry) {
			changes.Add = append(changes.Add, entry)
		}
	}

	return changes, nil
}

// ApplyIPs applies IPChanges to an existing IP list and returns a new sorted
// slice. The input slice is not modified. Removing an IP that does not exist
// is a silent no-op.
func ApplyIPs(existing []string, changes *IPChanges) []string {
	// Start with existing, filtering out removals.
	seen := make(map[string]bool, len(existing))
	var result []string
	for _, ip := range existing {
		if changes.Remove[ip] {
			continue
		}
		if !seen[ip] {
			seen[ip] = true
			result = append(result, ip)
		}
	}

	// Add new IPs, deduplicating against existing.
	for _, ip := range changes.Add {
		if !seen[ip] {
			seen[ip] = true
			result = append(result, ip)
		}
	}

	sort.Strings(result)
	return result
}
