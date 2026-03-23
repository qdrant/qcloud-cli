package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name      string
		raw       []string
		wantSet   map[string]string
		wantRm    []string
		wantError string
	}{
		{
			name:    "single set",
			raw:     []string{"env=prod"},
			wantSet: map[string]string{"env": "prod"},
		},
		{
			name:    "empty value is valid",
			raw:     []string{"env="},
			wantSet: map[string]string{"env": ""},
		},
		{
			name:    "value with equals sign",
			raw:     []string{"a=b=c"},
			wantSet: map[string]string{"a": "b=c"},
		},
		{
			name:    "multiple sets",
			raw:     []string{"env=prod", "team=platform"},
			wantSet: map[string]string{"env": "prod", "team": "platform"},
		},
		{
			name:   "single remove",
			raw:    []string{"env-"},
			wantRm: []string{"env"},
		},
		{
			name:    "mixed set and remove",
			raw:     []string{"env=prod", "old-"},
			wantSet: map[string]string{"env": "prod"},
			wantRm:  []string{"old"},
		},
		{
			name:   "last wins: set then remove",
			raw:    []string{"env=prod", "env-"},
			wantRm: []string{"env"},
		},
		{
			name:    "last wins: remove then set",
			raw:     []string{"env-", "env=prod"},
			wantSet: map[string]string{"env": "prod"},
		},
		{
			name:    "key with prefix slash",
			raw:     []string{"my.prefix/env=prod"},
			wantSet: map[string]string{"my.prefix/env": "prod"},
		},
		{
			name:      "no equals no dash",
			raw:       []string{"badformat"},
			wantError: "invalid --label value",
		},
		{
			name:      "empty key with value",
			raw:       []string{"=value"},
			wantError: "empty key",
		},
		{
			name:      "empty string",
			raw:       []string{""},
			wantError: "empty --label value",
		},
		{
			name:      "bare dash",
			raw:       []string{"-"},
			wantError: "empty key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := util.ParseLabels(tt.raw)
			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				return
			}
			require.NoError(t, err)

			if tt.wantSet != nil {
				assert.Equal(t, tt.wantSet, changes.Set)
			} else {
				assert.Empty(t, changes.Set)
			}

			var gotRm []string
			for k := range changes.Remove {
				gotRm = append(gotRm, k)
			}
			if tt.wantRm != nil {
				assert.ElementsMatch(t, tt.wantRm, gotRm)
			} else {
				assert.Empty(t, gotRm)
			}
		})
	}
}

func TestApplyLabels(t *testing.T) {
	tests := []struct {
		name     string
		existing []*commonv1.KeyValue
		set      map[string]string
		remove   []string
		want     map[string]string
	}{
		{
			name:     "upsert existing key",
			existing: kvs("env", "staging", "team", "infra"),
			set:      map[string]string{"env": "prod"},
			want:     map[string]string{"env": "prod", "team": "infra"},
		},
		{
			name:     "add new key",
			existing: kvs("env", "staging"),
			set:      map[string]string{"team": "platform"},
			want:     map[string]string{"env": "staging", "team": "platform"},
		},
		{
			name:     "remove existing key",
			existing: kvs("env", "staging", "team", "infra"),
			remove:   []string{"team"},
			want:     map[string]string{"env": "staging"},
		},
		{
			name:     "remove nonexistent key is silent",
			existing: kvs("env", "staging"),
			remove:   []string{"nonexistent"},
			want:     map[string]string{"env": "staging"},
		},
		{
			name:     "remove all keys",
			existing: kvs("env", "staging", "team", "infra"),
			remove:   []string{"env", "team"},
			want:     map[string]string{},
		},
		{
			name:     "add to empty",
			existing: nil,
			set:      map[string]string{"env": "prod"},
			want:     map[string]string{"env": "prod"},
		},
		{
			name:     "no changes",
			existing: kvs("env", "staging"),
			want:     map[string]string{"env": "staging"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := &util.LabelChanges{
				Set:    tt.set,
				Remove: make(map[string]bool),
			}
			if changes.Set == nil {
				changes.Set = make(map[string]string)
			}
			for _, k := range tt.remove {
				changes.Remove[k] = true
			}

			result := util.ApplyLabels(tt.existing, changes)

			got := make(map[string]string)
			for _, kv := range result {
				got[kv.GetKey()] = kv.GetValue()
			}
			assert.Equal(t, tt.want, got)

			// Verify sorted by key
			for i := 1; i < len(result); i++ {
				assert.Less(t, result[i-1].GetKey(), result[i].GetKey(), "result should be sorted by key")
			}
		})
	}
}

// kvs creates key-value pairs from alternating key, value strings.
func kvs(pairs ...string) []*commonv1.KeyValue {
	var result []*commonv1.KeyValue
	for i := 0; i < len(pairs); i += 2 {
		result = append(result, &commonv1.KeyValue{Key: pairs[i], Value: pairs[i+1]})
	}
	return result
}
