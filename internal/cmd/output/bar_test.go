package output_test

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

func TestUsageBar_Plain(t *testing.T) {
	bar := output.UsageBar([]output.BarSegment{
		{Value: 1, Color: output.BarColorSystem},
		{Value: 4, Color: output.BarColorData},
	}, 10, 12, false)

	assert.Equal(t, "[█████     ]", bar)
}

func TestUsageBar_PlainOverCapacity(t *testing.T) {
	bar := output.UsageBar([]output.BarSegment{{Value: 15, Color: output.BarColorData}}, 10, 12, false)

	assert.Equal(t, "[██████████]", bar)
}

func TestUsageBar_Color(t *testing.T) {
	bar := output.UsageBar([]output.BarSegment{
		{Value: 2, Color: output.BarColorSystem},
		{Value: 0, Color: output.BarColorCache},
		{Value: 3, Color: output.BarColorData},
	}, 10, 10, true)

	assert.Equal(t, 10, lipgloss.Width(bar))
	assert.Equal(t, strings.Repeat("█", 10), stripANSI(bar))
	assert.Contains(t, bar, "\x1b[34", "system segment is blue")
	assert.Contains(t, bar, "\x1b[32", "data segment is green")
	assert.Contains(t, bar, "\x1b[90", "free capacity is gray")
	assert.NotContains(t, bar, "\x1b[33", "empty cache segment is not drawn")
}

func TestBarLegend(t *testing.T) {
	assert.Equal(t, "cache 1 GiB", output.BarLegend("cache 1 GiB", output.BarColorCache, false))
	assert.Equal(t, "■ cache 1 GiB", stripANSI(output.BarLegend("cache 1 GiB", output.BarColorCache, true)))
}

func TestColorEnabled_NotATerminal(t *testing.T) {
	assert.False(t, output.ColorEnabled(&bytes.Buffer{}))
}

func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			b.WriteRune(r)
		}
	}

	return b.String()
}
