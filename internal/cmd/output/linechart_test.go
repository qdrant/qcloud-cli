package output_test

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

func chartPoints(since time.Time, values ...float64) []output.TimePoint {
	points := make([]output.TimePoint, 0, len(values))
	for i, v := range values {
		points = append(points, output.TimePoint{Time: since.Add(time.Duration(i) * time.Hour), Value: v})
	}

	return points
}

func chartOpts(since time.Time, hours int, yMax float64) output.TimeChartOptions {
	return output.TimeChartOptions{
		Since:  since,
		Until:  since.Add(time.Duration(hours) * time.Hour),
		YMax:   yMax,
		Width:  60,
		Height: 14,
		YLabel: func(v float64) string { return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64) },
	}
}

func TestTimeChart_RoundsYAxis(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	chart := output.TimeChart([]output.ChartSeries{{Name: "a", Points: chartPoints(since, 1, 2, 1.5)}}, chartOpts(since, 2, 2))

	lines := strings.Split(chart, "\n")
	assert.Len(t, lines, 14)
	// 2 with 5% headroom is 2.1, which rounds up to 4 steps of 0.6.
	assert.True(t, strings.HasPrefix(lines[0], "2.4│"), lines[0])
	assert.Contains(t, chart, "1.8│")
	assert.Contains(t, chart, "1.2│")
	assert.Contains(t, chart, "0.6│")
	assert.Contains(t, chart, "0└")
}

func TestTimeChart_RaisesYAxisToValues(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	chart := output.TimeChart([]output.ChartSeries{{Name: "a", Points: chartPoints(since, 1, 30)}}, chartOpts(since, 1, 2))

	// 30 with 5% headroom is 31.5, which rounds up to 4 steps of 8.
	assert.True(t, strings.HasPrefix(chart, "32│"), chart)
}

func TestTimeChart_YUnit(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	opts := chartOpts(since, 1, 8*(1<<30))
	opts.YUnit = 1 << 30
	opts.YLabel = func(v float64) string { return output.Bytes(uint64(v)) }

	chart := output.TimeChart([]output.ChartSeries{{Name: "a", Points: chartPoints(since, 1<<30)}}, opts)

	assert.True(t, strings.HasPrefix(chart, " 10 GiB│"), chart)
	assert.Contains(t, chart, "7.5 GiB│")
}

func TestTimeChart_TimeLabels(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	series := []output.ChartSeries{{Name: "a", Points: chartPoints(since, 1, 2)}}

	day := output.TimeChart(series, chartOpts(since, 24, 2))
	assert.Contains(t, day, "00:00")
	assert.NotContains(t, day, "10/20")

	week := output.TimeChart(series, chartOpts(since, 7*24, 2))
	assert.Contains(t, week, "10/20 00:00")
}

func TestTimeChart_SkipsEmptySeries(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	withEmpty := output.TimeChart([]output.ChartSeries{
		{Name: "empty"},
		{Name: "a", Points: chartPoints(since, 1, 2, 1.5)},
	}, chartOpts(since, 2, 2))
	without := output.TimeChart([]output.ChartSeries{
		{Name: "a", Points: chartPoints(since, 1, 2, 1.5)},
	}, chartOpts(since, 2, 2))

	assert.Equal(t, without, withEmpty)
	assert.Regexp(t, "[⠁-⣿]", withEmpty, "the line is drawn")
}

func TestTimeChart_Color(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	opts := chartOpts(since, 2, 2)
	opts.Color = true

	chart := output.TimeChart([]output.ChartSeries{
		{Name: "capacity", Color: output.ChartColorCapacity, Points: chartPoints(since, 2, 2, 2)},
		{Name: "a", Color: output.BarColorData, Points: chartPoints(since, 1, 1.5, 1)},
	}, opts)

	assert.Contains(t, chart, "\x1b[31", "capacity is red")
	assert.Contains(t, chart, "\x1b[32", "data is green")
}

func TestChartLegend(t *testing.T) {
	series := []output.ChartSeries{
		{Name: "capacity 2 vCPU", Color: output.ChartColorCapacity},
		{Name: "+ usage", Color: output.BarColorData},
	}

	assert.Equal(t, "capacity 2 vCPU  + usage", output.ChartLegend(series, false))
	assert.Equal(t, "■ capacity 2 vCPU  ■ + usage", stripANSI(output.ChartLegend(series, true)))
}

func TestTerminalWidth_NotATerminal(t *testing.T) {
	assert.Equal(t, 100, output.TerminalWidth(&bytes.Buffer{}, 100))
}

func TestTimeChart_FillWithColor(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	opts := chartOpts(since, 2, 4)
	opts.Color = true

	chart := output.TimeChart([]output.ChartSeries{
		{Name: "capacity", Color: output.ChartColorCapacity, Points: chartPoints(since, 4, 4, 4)},
		{Name: "system", Color: output.BarColorSystem, Points: chartPoints(since, 1, 1, 1), Fill: true},
		{Name: "+ data", Color: output.BarColorData, Points: chartPoints(since, 2, 3, 2.5), Fill: true},
	}, opts)

	plain := stripANSI(chart)
	assert.Contains(t, plain, "█", "areas are filled with blocks")
	assert.Regexp(t, "[⠁-⣿]", plain, "the capacity is still a line")
	assert.Contains(t, chart, "\x1b[34", "system area is blue")
	assert.Contains(t, chart, "\x1b[32", "data area is green")
	assert.Contains(t, chart, "\x1b[31", "capacity line is red")

	// Every column of the plot is filled, since the bottom row lies within the
	// system area everywhere.
	lines := strings.Split(plain, "\n")
	_, area, _ := strings.Cut(lines[len(lines)-3], "│")
	assert.NotEmpty(t, area)
	assert.Equal(t, strings.Repeat("█", len([]rune(area))), area)
}

func TestTimeChart_FillSparsePointsHasNoGaps(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	opts := chartOpts(since, 1, 2)
	opts.Color = true

	// Two points an hour apart, over a chart dozens of columns wide.
	chart := stripANSI(output.TimeChart([]output.ChartSeries{
		{Name: "a", Color: output.BarColorData, Points: chartPoints(since, 1, 2), Fill: true},
	}, opts))

	lines := strings.Split(chart, "\n")
	_, area, _ := strings.Cut(lines[len(lines)-3], "│")
	assert.NotContains(t, area, " ")
}

func TestTimeChart_FillOnlyDrawsAxes(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	opts := chartOpts(since, 2, 2)
	opts.Color = true

	chart := stripANSI(output.TimeChart([]output.ChartSeries{
		{Name: "a", Color: output.BarColorData, Points: chartPoints(since, 1, 2, 1.5), Fill: true},
	}, opts))

	assert.Contains(t, chart, "2.4│")
	assert.Contains(t, chart, "0└")
	assert.Contains(t, chart, "█")
}

func TestTimeChart_FillWithoutColorDrawsLines(t *testing.T) {
	since := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

	chart := output.TimeChart([]output.ChartSeries{
		{Name: "a", Points: chartPoints(since, 1, 2, 1.5), Fill: true},
	}, chartOpts(since, 2, 2))

	assert.NotContains(t, chart, "█")
	assert.Regexp(t, "[⠁-⣿]", chart)
}
