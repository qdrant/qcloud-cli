package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestParseIPs(t *testing.T) {
	tests := []struct {
		name      string
		raw       []string
		wantAdd   []string
		wantRm    []string
		wantError string
	}{
		{
			name:    "single add",
			raw:     []string{"10.0.0.0/8"},
			wantAdd: []string{"10.0.0.0/8"},
		},
		{
			name:   "single remove",
			raw:    []string{"10.0.0.0/8-"},
			wantRm: []string{"10.0.0.0/8"},
		},
		{
			name:    "mixed add and remove",
			raw:     []string{"10.0.0.0/8", "172.16.0.0/12-"},
			wantAdd: []string{"10.0.0.0/8"},
			wantRm:  []string{"172.16.0.0/12"},
		},
		{
			name:   "last wins: add then remove",
			raw:    []string{"10.0.0.0/8", "10.0.0.0/8-"},
			wantRm: []string{"10.0.0.0/8"},
		},
		{
			name:    "last wins: remove then add",
			raw:     []string{"10.0.0.0/8-", "10.0.0.0/8"},
			wantAdd: []string{"10.0.0.0/8"},
		},
		{
			name:    "duplicate adds",
			raw:     []string{"10.0.0.0/8", "10.0.0.0/8"},
			wantAdd: []string{"10.0.0.0/8"},
		},
		{
			name:    "multiple adds",
			raw:     []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
			wantAdd: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		},
		{
			name:      "empty string",
			raw:       []string{""},
			wantError: "empty --allowed-ip value",
		},
		{
			name:      "bare dash",
			raw:       []string{"-"},
			wantError: "empty IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := util.ParseIPs(tt.raw)
			if tt.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
				return
			}
			require.NoError(t, err)

			if tt.wantAdd != nil {
				assert.Equal(t, tt.wantAdd, changes.Add)
			} else {
				assert.Empty(t, changes.Add)
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

func TestApplyIPs(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		add      []string
		remove   []string
		want     []string
	}{
		{
			name: "add to empty",
			add:  []string{"10.0.0.0/8"},
			want: []string{"10.0.0.0/8"},
		},
		{
			name:     "add new IP to existing",
			existing: []string{"10.0.0.0/8"},
			add:      []string{"172.16.0.0/12"},
			want:     []string{"10.0.0.0/8", "172.16.0.0/12"},
		},
		{
			name:     "remove existing IP",
			existing: []string{"10.0.0.0/8", "172.16.0.0/12"},
			remove:   []string{"172.16.0.0/12"},
			want:     []string{"10.0.0.0/8"},
		},
		{
			name:     "remove nonexistent IP is silent",
			existing: []string{"10.0.0.0/8"},
			remove:   []string{"192.168.0.0/16"},
			want:     []string{"10.0.0.0/8"},
		},
		{
			name:     "remove all IPs",
			existing: []string{"10.0.0.0/8", "172.16.0.0/12"},
			remove:   []string{"10.0.0.0/8", "172.16.0.0/12"},
			want:     []string{},
		},
		{
			name:     "add duplicate of existing",
			existing: []string{"10.0.0.0/8"},
			add:      []string{"10.0.0.0/8"},
			want:     []string{"10.0.0.0/8"},
		},
		{
			name:     "no changes",
			existing: []string{"10.0.0.0/8"},
			want:     []string{"10.0.0.0/8"},
		},
		{
			name:     "result is sorted",
			existing: []string{"192.168.0.0/16"},
			add:      []string{"10.0.0.0/8"},
			want:     []string{"10.0.0.0/8", "192.168.0.0/16"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := &util.IPChanges{
				Add:    tt.add,
				Remove: make(map[string]bool),
			}
			for _, ip := range tt.remove {
				changes.Remove[ip] = true
			}

			result := util.ApplyIPs(tt.existing, changes)

			if len(tt.want) == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.want, result)
			}

			// Verify sorted
			for i := 1; i < len(result); i++ {
				assert.Less(t, result[i-1], result[i], "result should be sorted")
			}
		})
	}
}
