package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeMillicores(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"1", "1000m", false},
		{"2", "2000m", false},
		{"0.5", "500m", false},
		{"0.25", "250m", false},
		{"1000m", "1000m", false},
		{"500m", "500m", false},
		{"bad", "", true},
		{"1.2.3", "", true},
	}
	for _, tt := range tests {
		got, err := normalizeMillicores(tt.input)
		if tt.wantErr {
			require.Error(t, err, "input: %q", tt.input)
			continue
		}
		require.NoError(t, err, "input: %q", tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}

func TestParseDiskGiB(t *testing.T) {
	tests := []struct {
		input   string
		want    uint32
		wantErr bool
	}{
		{"100GiB", 100, false},
		{"100Gi", 100, false},
		{"100G", 100, false},
		{"100", 100, false},
		{"1TiB", 1024, false},
		{"1Ti", 1024, false},
		{"1T", 1024, false},
		{"2TiB", 2048, false},
		{"bad", 0, true},
		{"100X", 0, true},
	}
	for _, tt := range tests {
		got, err := parseDiskGiB(tt.input)
		if tt.wantErr {
			require.Error(t, err, "input: %q", tt.input)
			continue
		}
		require.NoError(t, err, "input: %q", tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}

func TestParseCPUMillicores(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"1", 1000, false},
		{"2", 2000, false},
		{"0.5", 500, false},
		{"0.25", 250, false},
		{"1000m", 1000, false},
		{"500m", 500, false},
		{"4000m", 4000, false},
		{"bad", 0, true},
		{"badm", 0, true},
	}
	for _, tt := range tests {
		got, err := parseCPUMillicores(tt.input)
		if tt.wantErr {
			require.Error(t, err, "input: %q", tt.input)
			continue
		}
		require.NoError(t, err, "input: %q", tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}

func TestParseRAMGiB(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"4GiB", 4, false},
		{"4Gi", 4, false},
		{"4G", 4, false},
		{"4", 4, false},
		{"8GiB", 8, false},
		{"8Gi", 8, false},
		{"16", 16, false},
		{"bad", 0, true},
		{"badGiB", 0, true},
	}
	for _, tt := range tests {
		got, err := parseRAMGiB(tt.input)
		if tt.wantErr {
			require.Error(t, err, "input: %q", tt.input)
			continue
		}
		require.NoError(t, err, "input: %q", tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}

func TestParseGPUMillicores(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"1000m", 1000, false},
		{"2000m", 2000, false},
		{"1", 1000, false},
		{"2", 2000, false},
		{"bad", 0, true},
		{"badm", 0, true},
	}
	for _, tt := range tests {
		got, err := parseGPUMillicores(tt.input)
		if tt.wantErr {
			require.Error(t, err, "input: %q", tt.input)
			continue
		}
		require.NoError(t, err, "input: %q", tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}

func TestGpuMillicoresToCount(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "1 GPU", in: "1000m", want: "1"},
		{name: "2 GPUs", in: "2000m", want: "2"},
		{name: "4 GPUs", in: "4000m", want: "4"},
		{name: "empty string", in: "", want: ""},
		{name: "not a number", in: "abc", want: ""},
		{name: "zero", in: "0m", want: ""},
		{name: "negative", in: "-1000m", want: ""},
		{name: "fractional GPU", in: "500m", want: ""},
		{name: "no m suffix", in: "1000", want: "1"},
		{name: "non-multiple of 1000", in: "1500m", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, gpuMillicoresToCount(tt.in))
		})
	}
}

func TestNormalizeRAM(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"8", "8GiB", false},
		{"16", "16GiB", false},
		{"8G", "8GiB", false},
		{"8Gi", "8GiB", false},
		{"8GiB", "8GiB", false},
		{"bad", "", true},
		{"8X", "", true},
		{"100iB", "", true},
		{"512MiB", "", true},
		{"1TiB", "", true},
	}
	for _, tt := range tests {
		got, err := normalizeRAM(tt.input)
		if tt.wantErr {
			require.Error(t, err, "input: %q", tt.input)
			continue
		}
		require.NoError(t, err, "input: %q", tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}
