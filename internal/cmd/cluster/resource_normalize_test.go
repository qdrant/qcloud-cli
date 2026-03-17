package cluster

import (
	"testing"
)

func TestNormalizeCPU(t *testing.T) {
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
		got, err := normalizeCPU(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("normalizeCPU(%q): expected error, got %q", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeCPU(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("normalizeCPU(%q) = %q, want %q", tt.input, got, tt.want)
		}
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
			if err == nil {
				t.Errorf("normalizeRAM(%q): expected error, got %q", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeRAM(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("normalizeRAM(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
