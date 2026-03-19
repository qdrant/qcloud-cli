package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, KiB, ByteQuantity(1024))
	assert.Equal(t, MiB, ByteQuantity(1024*1024))
	assert.Equal(t, GiB, ByteQuantity(1024*1024*1024))
	assert.Equal(t, TiB, ByteQuantity(1024*1024*1024*1024))
	assert.Equal(t, PiB, ByteQuantity(1024*1024*1024*1024*1024))
	assert.Equal(t, EiB, ByteQuantity(1024*1024*1024*1024*1024*1024))
}

func TestParseByteQuantity(t *testing.T) {
	tests := []struct {
		input   string
		want    ByteQuantity
		wantErr bool
	}{
		{"8GiB", 8 * GiB, false},
		{"8Gi", 8 * GiB, false},
		{"8G", 8 * GiB, false},
		{"8", 8 * GiB, false},
		{"512MiB", 512 * MiB, false},
		{"512Mi", 512 * MiB, false},
		{"512M", 512 * MiB, false},
		{"1TiB", 1 * TiB, false},
		{"1Ti", 1 * TiB, false},
		{"1T", 1 * TiB, false},
		{"2PiB", 2 * PiB, false},
		{"2Pi", 2 * PiB, false},
		{"2P", 2 * PiB, false},
		{"1EiB", 1 * EiB, false},
		{"1Ei", 1 * EiB, false},
		{"1E", 1 * EiB, false},
		{"4KiB", 4 * KiB, false},
		{"4Ki", 4 * KiB, false},
		{"4K", 4 * KiB, false},
		{"bad", 0, true},
		{"badGiB", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseByteQuantity(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestByteQuantity_Set(t *testing.T) {
	tests := []struct {
		input   string
		want    ByteQuantity
		wantErr bool
	}{
		{"8GiB", 8 * GiB, false},
		{"512MiB", 512 * MiB, false},
		{"1PiB", 1 * PiB, false},
		{"bad", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var b ByteQuantity
			err := b.Set(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, b)
		})
	}
}

func TestByteQuantity_String(t *testing.T) {
	tests := []struct {
		input ByteQuantity
		want  string
	}{
		{0, ""},
		{8 * GiB, "8GiB"},
		{16 * GiB, "16GiB"},
		{1 * TiB, "1TiB"},
		{512 * MiB, "512MiB"},
		{4 * KiB, "4KiB"},
		{2 * PiB, "2PiB"},
		{1 * EiB, "1EiB"},
		{1500, "1500"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.input.String())
	}
}

func TestByteQuantity_GiB(t *testing.T) {
	assert.Equal(t, int64(8), (8 * GiB).GiB())
	assert.Equal(t, int64(0), (512 * MiB).GiB())
	assert.Equal(t, int64(1024), TiB.GiB())
}

func TestFormatByteQuantity(t *testing.T) {
	tests := []struct {
		b    ByteQuantity
		unit string
		want string
	}{
		{8 * GiB, UnitGiB, "8GiB"},
		{512 * MiB, UnitMiB, "512MiB"},
		{1 * TiB, UnitTiB, "1TiB"},
		{4 * KiB, UnitKiB, "4KiB"},
		{2 * PiB, UnitPiB, "2PiB"},
		{1 * EiB, UnitEiB, "1EiB"},
		{8 * GiB, "", "8589934592"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, FormatByteQuantity(tt.b, tt.unit))
	}
}
