package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMillicores(t *testing.T) {
	tests := []struct {
		input   string
		want    Millicores
		wantErr bool
	}{
		{"1", 1000, false},
		{"2", 2000, false},
		{"0.5", 500, false},
		{"0.25", 250, false},
		{"1000m", 1000, false},
		{"500m", 500, false},
		{"4000m", 4000, false},
		// Floating-point values that would truncate incorrectly without math.Round.
		{"0.7", 700, false},
		{"1.1", 1100, false},
		{"2.3", 2300, false},
		{"0.001", 1, false},
		{"0.0005", 1, false}, // rounds up from 0.5
		{"0.0004", 0, false}, // rounds down
		{"99.999", 99999, false},
		{"bad", 0, true},
		{"badm", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseMillicores(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMillicores_Set(t *testing.T) {
	tests := []struct {
		input   string
		want    Millicores
		wantErr bool
	}{
		{"1", 1000, false},
		{"0.5", 500, false},
		{"1000m", 1000, false},
		{"bad", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var m Millicores
			err := m.Set(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, m)
		})
	}
}

func TestMillicores_String(t *testing.T) {
	tests := []struct {
		input Millicores
		want  string
	}{
		{0, ""},
		{1000, "1"},
		{2000, "2"},
		{500, "0.5"},
		{250, "0.25"},
		{1500, "1.5"},
		{1, "0.001"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.input.String())
	}
}

func TestMillicores_Cores(t *testing.T) {
	tests := []struct {
		input Millicores
		want  string
	}{
		{0, "0"},
		{1000, "1"},
		{500, "0.5"},
		{1500, "1.5"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.input.Cores())
	}
}
