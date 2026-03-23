package util

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

// LabelChanges holds the parsed result of --label flags.
type LabelChanges struct {
	Set    map[string]string
	Remove map[string]bool
}

// ParseLabels parses a slice of raw --label flag values into upserts and removals.
//
// Accepted forms:
//   - "key=value" -- upsert (empty value is valid: "key=")
//   - "key-"      -- remove the label with that key
//
// Returns an error for malformed entries like "key" (no = and no trailing -),
// "=value" (empty key), or "" (empty string).
//
// When the same key appears multiple times, the last occurrence wins.
func ParseLabels(raw []string) (*LabelChanges, error) {
	changes := &LabelChanges{
		Set:    make(map[string]string),
		Remove: make(map[string]bool),
	}

	for _, entry := range raw {
		if entry == "" {
			return nil, fmt.Errorf("empty --label value")
		}

		if key, value, ok := strings.Cut(entry, "="); ok {
			if key == "" {
				return nil, fmt.Errorf("empty key in --label %q", entry)
			}

			// Last operation wins, hence if a label is set after it's removal, clean it from the Remove field.
			delete(changes.Remove, key)
			changes.Set[key] = value
			continue
		}

		if strings.HasSuffix(entry, "-") {
			key := entry[:len(entry)-1]
			if key == "" {
				return nil, fmt.Errorf("empty key in --label %q", entry)
			}

			// Last operation wins, hence if a label is set after it's set, clean it from the Set field.
			delete(changes.Set, key)
			changes.Remove[key] = true
			continue
		}

		return nil, fmt.Errorf("invalid --label value %q: use 'key=value' to set or 'key-' to remove", entry)
	}

	return changes, nil
}

// ApplyLabels applies LabelChanges to an existing label set and returns a new
// slice sorted by key. The input slice is not modified. Removing a key that
// does not exist is a silent no-op.
func ApplyLabels(existing []*commonv1.KeyValue, changes *LabelChanges) []*commonv1.KeyValue {
	merged := make(map[string]string, len(existing))
	for _, kv := range existing {
		merged[kv.GetKey()] = kv.GetValue()
	}

	for key := range changes.Remove {
		delete(merged, key)
	}
	maps.Copy(merged, changes.Set)

	result := make([]*commonv1.KeyValue, 0, len(merged))
	for k, v := range merged {
		result = append(result, &commonv1.KeyValue{Key: k, Value: v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].GetKey() < result[j].GetKey()
	})

	return result
}
