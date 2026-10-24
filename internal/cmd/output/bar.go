package output

import (
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/barchart"
	"github.com/charmbracelet/colorprofile"
)

// ANSI colors used for the segments of a usage bar.
const (
	BarColorSystem = "4" // blue
	BarColorCache  = "3" // yellow
	BarColorData   = "2" // green
	BarColorFree   = "8" // gray
)

// minUsageBarWidth is the narrowest a usage bar is drawn.
const minUsageBarWidth = 3

// BarSegment is one stacked part of a usage bar.
type BarSegment struct {
	Value float64
	Color string
}

// ColorEnabled reports whether w is a terminal that supports colors, honoring
// NO_COLOR and TERM=dumb.
func ColorEnabled(w io.Writer) bool {
	return colorprofile.Detect(w, os.Environ()) >= colorprofile.ANSI
}

// UsageBar renders segments as a horizontal stacked bar of width cells, scaled
// so that a full bar equals total. With color, every segment has its own color
// and the remaining capacity is drawn as a gray track. Without color, the used
// part is drawn inside brackets, since uncolored segments look the same.
func UsageBar(segments []BarSegment, total float64, width int, color bool) string {
	// Leave room for at least one cell between the brackets of an uncolored bar.
	width = max(width, minUsageBarWidth)
	if !color {
		width -= 2
	}

	values := make([]barchart.BarValue, 0, len(segments)+1)
	used := 0.0
	for _, s := range segments {
		if s.Value <= 0 {
			continue
		}

		values = append(values, barchart.BarValue{Value: s.Value, Style: barStyle(s.Color, color)})
		used += s.Value
	}

	if !color {
		// Uncolored segments are indistinguishable, and their boundaries would
		// show up as partial blocks in the middle of the bar.
		values = []barchart.BarValue{{Value: used}}
	} else if total > used {
		values = append(values, barchart.BarValue{Value: total - used, Style: barStyle(BarColorFree, color)})
	}

	// A partially filled cell shows the next segment's color behind it, so
	// adjacent segments meet without a gap.
	for i := 0; color && i < len(values)-1; i++ {
		values[i].Style = values[i].Style.Background(values[i+1].Style.GetForeground())
	}

	chart := barchart.New(width, 1,
		barchart.WithHorizontalBars(),
		barchart.WithNoAxis(),
		barchart.WithNoAutoMaxValue(),
		barchart.WithMaxValue(total),
	)
	chart.Push(barchart.BarData{Values: values})
	chart.Draw()
	bar := chart.View()

	if !color {
		return "[" + bar + strings.Repeat(" ", max(width-lipgloss.Width(bar), 0)) + "]"
	}

	return bar
}

// BarLegend renders a colored swatch followed by label, or just the label
// without color.
func BarLegend(label, barColor string, color bool) string {
	if !color {
		return label
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render("■") + " " + label
}

func barStyle(barColor string, color bool) lipgloss.Style {
	if !color {
		return lipgloss.NewStyle()
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
}
