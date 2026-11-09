package util_test

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
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

func TestReadPastTimeRange(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		args      []string
		wantSince time.Time
		wantUntil time.Time
		wantErr   string
	}{
		"defaults":              {nil, now.Add(-3 * time.Hour), now, ""},
		"only since":            {[]string{"--since", "6h"}, now.Add(-6 * time.Hour), now, ""},
		"only until":            {[]string{"--until", "2026-10-01T12:00:00Z"}, time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC), time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC), ""},
		"future until":          {[]string{"--since", "1h", "--until", "2099-01-01"}, now.Add(-time.Hour), now, ""},
		"since in the future":   {[]string{"--since", "2099-01-01"}, time.Time{}, time.Time{}, "--since must be in the past"},
		"since after until":     {[]string{"--since", "1h", "--until", "2h"}, time.Time{}, time.Time{}, "--since must be before --until"},
		"since after cut until": {[]string{"--since", "2026-10-07T13:00:00Z", "--until", "2099-01-01"}, time.Time{}, time.Time{}, "--since must be in the past"},
		"invalid since":         {[]string{"--since", "yesterday"}, time.Time{}, time.Time{}, "invalid --since"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("since", "", "")
			cmd.Flags().String("until", "", "")
			require.NoError(t, cmd.Flags().Parse(tt.args))

			since, until, err := util.ReadPastTimeRange(cmd, now, 3*time.Hour)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSince, since)
			assert.Equal(t, tt.wantUntil, until)
		})
	}
}

func TestAddTimeRangeFlags(t *testing.T) {
	for _, tt := range []struct {
		defaultRange time.Duration
		want         string
	}{
		{time.Hour, "default: 1h before --until"},
		{10 * time.Minute, "default: 10m before --until"},
		{24 * time.Hour, "default: 1 day before --until"},
		{7 * 24 * time.Hour, "default: 7 days before --until"},
	} {
		cmd := &cobra.Command{}
		util.AddTimeRangeFlags(cmd, tt.defaultRange)
		assert.Contains(t, cmd.Flags().Lookup("since").Usage, tt.want)
		assert.Equal(t, "s", cmd.Flags().Lookup("since").Shorthand)
		assert.Equal(t, "u", cmd.Flags().Lookup("until").Shorthand)
	}
}
