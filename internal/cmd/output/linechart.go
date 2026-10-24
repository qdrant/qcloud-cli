package output

import (
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/v2/canvas"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/graph"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	"github.com/charmbracelet/x/term"
)

// chartYStep is the number of rows between two labels on the Y axis.
const chartYStep = 3

// Colors of the lines in a time chart.
const (
	ChartColorCapacity = "1" // red
	ChartColorRequests = "3" // yellow
	ChartColorLatency  = "1" // red
)

// TimePoint is a value at a point in time.
type TimePoint struct {
	Time  time.Time
	Value float64
}

// ChartSeries is one line of a time chart.
type ChartSeries struct {
	Name   string
	Color  string
	Points []TimePoint
	// Fill draws the area below the line instead of the line. Filled series
	// are stacked: each one must include the values of the filled series
	// before it, so it is drawn on top of them. Without color, a filled series
	// is drawn as a line, since areas of the same look would hide each other.
	Fill bool
}

// TimeChartOptions configures TimeChart.
type TimeChartOptions struct {
	Since, Until time.Time
	// YMax is the highest value the Y axis must show. It is raised to fit
	// every value, then rounded up to a value that gives even labels.
	YMax float64
	// YUnit is the unit labels are rounded to, such as 1 GiB for bytes.
	// It defaults to 1.
	YUnit  float64
	Width  int
	Height int
	// YLabel formats the values on the Y axis.
	YLabel func(float64) string
	Color  bool
}

// TimeChart renders series as braille line charts over the time range in opts,
// drawn in order, so a later series is drawn over an earlier one where they
// meet. Series without points are skipped.
func TimeChart(series []ChartSeries, opts TimeChartOptions) string {
	yMax := opts.YMax
	for _, s := range series {
		for _, p := range s.Points {
			yMax = max(yMax, p.Value)
		}
	}

	if yMax <= 0 {
		yMax = 1
	}

	yUnit := opts.YUnit
	if yUnit <= 0 {
		yUnit = 1
	}

	// The plot area is the height less the X axis and its labels. Labels sit
	// on every chartYStep-th row, so a plot area of a multiple of chartYStep
	// rows has labels at even fractions of the range.
	plotHeight := max((opts.Height-2)/chartYStep*chartYStep, chartYStep)
	// Leave some room above the highest value, so a line at the top, such as
	// the capacity, does not merge with the edge of the chart.
	yRange := niceAxisMax(yMax*1.05, yUnit, plotHeight/chartYStep)

	timeFormat := "01/02 15:04"
	if opts.Until.Sub(opts.Since) <= 24*time.Hour {
		timeFormat = "15:04"
	}

	chart := timeserieslinechart.New(opts.Width, plotHeight+2,
		timeserieslinechart.WithTimeRange(opts.Since, opts.Until),
		timeserieslinechart.WithYRange(0, yRange),
		timeserieslinechart.WithXYSteps(4, chartYStep),
		timeserieslinechart.WithXLabelFormatter(func(_ int, v float64) string {
			return time.UnixMilli(int64(v * 1e3)).UTC().Format(timeFormat)
		}),
		timeserieslinechart.WithYLabelFormatter(func(_ int, v float64) string {
			return opts.YLabel(v)
		}),
	)

	lines := make([]string, 0, len(series))
	var fills []ChartSeries
	for _, s := range series {
		if len(s.Points) == 0 {
			continue
		}

		if s.Fill && opts.Color {
			fills = append(fills, s)
			continue
		}

		chart.SetDataSetStyle(s.Name, barStyle(s.Color, opts.Color))
		for _, p := range s.Points {
			chart.PushDataSet(s.Name, timeserieslinechart.TimePoint{Time: p.Time, Value: p.Value})
		}

		lines = append(lines, s.Name)
	}

	if len(lines) > 0 {
		chart.DrawBrailleDataSets(lines)
	} else {
		chart.DrawXYAxisAndLabel()
	}

	drawAreas(&chart, fills, opts.Since, opts.Until, yRange)

	return chart.View()
}

// drawAreas fills the area below each series as columns of block runes, one
// per column of the chart. The outermost series is drawn first, and each one
// inside it over its lower part, so the stack shows every series in its color.
func drawAreas(chart *timeserieslinechart.Model, fills []ChartSeries, since, until time.Time, yRange float64) {
	width, height := chart.GraphWidth(), chart.GraphHeight()
	if len(fills) == 0 || width <= 0 || height <= 0 {
		return
	}

	startX := chart.Origin().X
	if chart.YStep() > 0 {
		startX++
	}

	bottom := chart.Origin().Y - 1
	columnDuration := until.Sub(since) / time.Duration(width)
	for i, fill := range slices.Backward(fills) {
		style := barStyle(fill.Color, true)
		if i+1 < len(fills) {
			// A partially filled top cell shows the series above behind it.
			style = style.Background(barStyle(fills[i+1].Color, true).GetForeground())
		}

		for col := range width {
			from := since.Add(time.Duration(col) * columnDuration)
			v, ok := columnValue(fill.Points, from, from.Add(columnDuration))
			if !ok {
				continue
			}

			rows := math.Min(v/yRange*float64(height), float64(height))
			graph.DrawColumnBottomToTop(&chart.Canvas, canvas.Point{X: startX + col, Y: bottom}, rows, style)
		}
	}
}

// columnValue returns the average of the points within [from, to), or when
// there are none, the value interpolated between the points around the middle
// of the range. It reports false outside of the points.
func columnValue(points []TimePoint, from, to time.Time) (float64, bool) {
	sum, n := 0.0, 0
	for _, p := range points {
		if !p.Time.Before(from) && p.Time.Before(to) {
			sum += p.Value
			n++
		}
	}

	if n > 0 {
		return sum / float64(n), true
	}

	mid := from.Add(to.Sub(from) / 2)
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		if a.Time.After(mid) || b.Time.Before(mid) {
			continue
		}

		span := b.Time.Sub(a.Time)
		if span <= 0 {
			return b.Value, true
		}

		return a.Value + (b.Value-a.Value)*float64(mid.Sub(a.Time))/float64(span), true
	}

	return 0, false
}

// niceAxisMax returns the smallest value of at least v that splits into
// intervals equal steps of 1, 1.5, 2, 2.5, 3, 4, 5, 6 or 8 times a power of ten
// units.
func niceAxisMax(v, unit float64, intervals int) float64 {
	step := v / unit / float64(intervals)
	magnitude := math.Pow(10, math.Floor(math.Log10(step)))
	for _, f := range []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10} {
		if f*magnitude >= step*(1-1e-9) {
			// Drop the floating point noise of the multiplications, such as
			// in 0.6 * 4 = 2.4000000000000004.
			v, _ := strconv.ParseFloat(strconv.FormatFloat(f*magnitude*float64(intervals), 'g', 12, 64), 64)
			return v * unit
		}
	}

	return v
}

// TerminalWidth returns the width of the terminal w writes to, or fallback
// when w is not a terminal.
func TerminalWidth(w io.Writer, fallback int) int {
	f, ok := w.(*os.File)
	if !ok || !term.IsTerminal(f.Fd()) {
		return fallback
	}

	width, _, err := term.GetSize(f.Fd())
	if err != nil || width <= 0 {
		return fallback
	}

	return width
}

// ChartLegend renders the legend entries of a chart on a single line.
func ChartLegend(series []ChartSeries, color bool) string {
	entries := make([]string, 0, len(series))
	for _, s := range series {
		entries = append(entries, BarLegend(s.Name, s.Color, color))
	}

	return strings.Join(entries, "  ")
}
