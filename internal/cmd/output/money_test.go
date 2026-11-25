package output_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

func TestUSDMillicents(t *testing.T) {
	for in, want := range map[uint64]string{
		0:                                   "$0.00",
		999:                                 "$0.00", // under a cent
		1_000:                               "$0.01",
		50 * output.MillicentsPerUSD:        "$50.00",
		4_012_999:                           "$40.12",
		1_234_567 * 1_000:                   "$12,345.67",
		1_000_000 * output.MillicentsPerUSD: "$1,000,000.00",
	} {
		assert.Equal(t, want, output.USDMillicents(in), "%d", in)
	}
}

func TestMillicents(t *testing.T) {
	for _, tc := range []struct {
		in       int64
		currency string
		want     string
	}{
		{0, "USD", "$0.00"},
		{1_234_567 * 1_000, "USD", "$12,345.67"},
		{50 * output.MillicentsPerUSD, "", "$50.00"},
		{-1_500 * 1_000, "USD", "-$15.00"},
		{1_234_567 * 1_000, "EUR", "12,345.67 EUR"},
		{-999, "EUR", "0.00 EUR"},
	} {
		assert.Equal(t, tc.want, output.Millicents(tc.in, tc.currency), "%d %s", tc.in, tc.currency)
	}
}
