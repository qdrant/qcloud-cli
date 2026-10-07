package util

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// ListChanges holds the parsed result of a string-list flag that supports
// additions and removals (e.g. --allowed-ip, --allowed-origin).
type ListChanges struct {
	Add    []string
	Remove map[string]bool
}

// IPChanges holds the parsed result of --allowed-ip flags.
type IPChanges = ListChanges

// ParseIPs parses a slice of raw --allowed-ip flag values into additions and removals.
//
// Accepted forms:
//   - "10.0.0.0/8"  -- add the IP CIDR
//   - "10.0.0.0/8-" -- remove the IP CIDR (trailing dash)
//
// When the same IP appears multiple times, the last occurrence wins.
func ParseIPs(raw []string) (*IPChanges, error) {
	return ParseListChanges("--allowed-ip", "IP", raw)
}

// ApplyIPs applies IPChanges to an existing IP list and returns a new sorted
// slice. The input slice is not modified. Removing an IP that does not exist
// is a silent no-op.
func ApplyIPs(existing []string, changes *IPChanges) []string {
	return ApplyListChanges(existing, changes)
}

// ParseListChanges parses a slice of raw flag values into additions and
// removals. A value with a trailing '-' marks the entry for removal; any other
// value is added. flag and item are only used in error messages (e.g.
// "--allowed-origin", "origin").
//
// When the same entry appears multiple times, the last occurrence wins.
func ParseListChanges(flag, item string, raw []string) (*ListChanges, error) {
	changes := &ListChanges{
		Remove: make(map[string]bool),
	}

	for _, entry := range raw {
		if entry == "" {
			return nil, fmt.Errorf("empty %s value", flag)
		}

		if strings.HasSuffix(entry, "-") {
			v := entry[:len(entry)-1]
			if v == "" {
				return nil, fmt.Errorf("empty %s in %s %q", item, flag, entry)
			}

			// Last operation wins: remove from Add if previously added.
			changes.Add = slices.DeleteFunc(changes.Add, func(s string) bool { return s == v })
			changes.Remove[v] = true
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

// ApplyListChanges applies ListChanges to an existing list and returns a new
// sorted slice. The input slice is not modified. Removing an entry that does
// not exist is a silent no-op.
func ApplyListChanges(existing []string, changes *ListChanges) []string {
	// Start with existing, filtering out removals.
	seen := make(map[string]bool, len(existing))
	var result []string
	for _, v := range existing {
		if changes.Remove[v] {
			continue
		}

		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	// Add new entries, deduplicating against existing.
	for _, v := range changes.Add {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	sort.Strings(result)
	return result
}
