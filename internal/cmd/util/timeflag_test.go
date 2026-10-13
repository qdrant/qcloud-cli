package util_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestParseTimeFlag(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		in   string
		want time.Time
	}{
		{"2026-10-01T08:30:00Z", time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)},
		{"2026-10-01", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
		{"30m", now.Add(-30 * time.Minute)},
		{"24h", now.Add(-24 * time.Hour)},
		{"7d", now.Add(-7 * 24 * time.Hour)},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := util.ParseTimeFlag(tt.in, now)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTimeFlag_Invalid(t *testing.T) {
	now := time.Now()
	for _, in := range []string{"", "not-a-date", "0d", "-1h", "0s", "1.5d", "2026-13-01"} {
		t.Run(in, func(t *testing.T) {
			_, err := util.ParseTimeFlag(in, now)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "must be RFC3339")
		})
	}
}

func TestParseUntilFlag(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		in   string
		want time.Time
	}{
		{"2026-10-01T08:30:00Z", time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)},
		{"2026-10-01", time.Date(2026, 10, 1, 23, 59, 59, 0, time.UTC)},
		{"1h", now.Add(-time.Hour)},
		{"2d", now.Add(-2 * 24 * time.Hour)},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := util.ParseUntilFlag(tt.in, now)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
