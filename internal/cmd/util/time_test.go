package util_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestParseDateEndOfDay(t *testing.T) {
	got, err := util.ParseDateEndOfDay("2026-12-31")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), got)
}

func TestParseDateEndOfDay_Invalid(t *testing.T) {
	for _, s := range []string{"", "2026-13-01", "2026-12-31T00:00:00Z", "31/12/2026"} {
		_, err := util.ParseDateEndOfDay(s)
		assert.Error(t, err, s)
	}
}
