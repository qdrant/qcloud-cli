package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

func TestParseUSDMillicents(t *testing.T) {
	for in, want := range map[string]uint64{
		"50":     5_000_000,
		"$50":    5_000_000,
		"49.99":  4_999_000,
		"0.5":    50_000,
		"0.01":   1_000,
		" 1000 ": 100_000_000,
	} {
		got, err := util.ParseUSDMillicents(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
}

func TestParseUSDMillicents_Invalid(t *testing.T) {
	for _, in := range []string{"", "$", "abc", "-5", "1.234", "1.", ".5", "0", "0.00", "1,000", "99999999999999999999"} {
		_, err := util.ParseUSDMillicents(in)
		assert.Error(t, err, in)
	}
}
