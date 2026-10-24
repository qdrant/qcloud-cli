package cluster

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/spf13/cobra"

	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

const (
	defaultMetricsHistoryRange = 7 * 24 * time.Hour
	metricsChartHeight         = 14
	metricsChartDefaultWidth   = 100
	metricsChartMinWidth       = 40
	metricsChartMaxWidth       = 200
)

type historyPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

func addMetricsHistoryRangeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("since", "s", "", "Start of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 24h or 7d; default: 7 days before --until)")
	cmd.Flags().StringP("until", "u", "", "End of the time range (RFC3339, YYYY-MM-DD, or a duration ago such as 1h; default: now)")
}

// readMetricsHistoryRange reads --since and --until, defaulting to the 7 days
// before --until, and --until to now. An --until in the future is cut to now:
// the API fills some series with zeros up to any time, which would read as
// measured values.
func readMetricsHistoryRange(cmd *cobra.Command, now time.Time) (time.Time, time.Time, error) {
	sinceTS, untilTS, err := util.ReadTimeRange(cmd, now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if sinceTS != nil && !sinceTS.AsTime().Before(now) {
		return time.Time{}, time.Time{}, errors.New("--since must be in the past")
	}

	until := now
	if untilTS != nil && untilTS.AsTime().Before(now) {
		until = untilTS.AsTime()
	}

	since := until.Add(-defaultMetricsHistoryRange)
	if sinceTS != nil {
		since = sinceTS.AsTime()
	}

	if !since.Before(until) {
		return time.Time{}, time.Time{}, errors.New("--since must be before --until")
	}

	return since, until, nil
}

// historyPoints converts metrics to points in chronological order.
func historyPoints(metrics []*monitoringv1.Metric) []historyPoint {
	points := make([]historyPoint, 0, len(metrics))
	for _, m := range metrics {
		points = append(points, historyPoint{Timestamp: m.GetTimestamp().AsTime().UTC(), Value: m.GetValue()})
	}

	slices.SortStableFunc(points, func(a, b historyPoint) int { return a.Timestamp.Compare(b.Timestamp) })

	return points
}

// valueAt returns the value of the point closest to t in points, which are in
// chronological order, or 0 when there are none. Series of the same query
// share timestamps, but taking the closest point keeps a stack intact when
// they do not.
func valueAt(points []historyPoint, t time.Time) float64 {
	i, _ := slices.BinarySearchFunc(points, t, func(p historyPoint, t time.Time) int { return p.Timestamp.Compare(t) })
	switch {
	case len(points) == 0:
		return 0
	case i == 0:
		return points[0].Value
	case i == len(points):
		return points[i-1].Value
	case t.Sub(points[i-1].Timestamp) <= points[i].Timestamp.Sub(t):
		return points[i-1].Value
	default:
		return points[i].Value
	}
}

// stack returns the points of base with offset, and the value at the same
// time in each of layers, added on top.
func stack(base []historyPoint, offset float64, layers ...[]historyPoint) []output.TimePoint {
	points := make([]output.TimePoint, 0, len(base))
	for _, p := range base {
		v := offset + p.Value
		for _, l := range layers {
			v += valueAt(l, p.Timestamp)
		}

		points = append(points, output.TimePoint{Time: p.Timestamp, Value: v})
	}

	return points
}

// constant returns a point with value v at the time of every point in base.
func constant(base []historyPoint, v float64) []output.TimePoint {
	points := make([]output.TimePoint, 0, len(base))
	for _, p := range base {
		points = append(points, output.TimePoint{Time: p.Timestamp, Value: v})
	}

	return points
}

type metricsChart struct {
	title string
	// base is the series the chart is drawn for. The chart shows empty, or
	// "no data" when empty is not set, when base has no points.
	base  []historyPoint
	empty string
	// total is the capacity, drawn as a line when set.
	total  float64
	series []output.ChartSeries
	// yTop is the least top of the Y axis, such as 100 for a percentage.
	yTop   float64
	yLabel func(float64) string
	yUnit  float64
}

// metricsChartWidth returns the width of the charts written to w: the width
// of the terminal, within limits.
func metricsChartWidth(w io.Writer) int {
	return min(max(output.TerminalWidth(w, metricsChartDefaultWidth), metricsChartMinWidth), metricsChartMaxWidth)
}

// printMetricsCharts prints the title, legend and chart of each of charts, with
// the labels on their Y axes padded to the same width, so they line up.
func printMetricsCharts(w io.Writer, charts []metricsChart, since, until time.Time, width int, color bool) {
	labelWidth := 0
	for _, c := range charts {
		if len(c.base) > 0 {
			labelWidth = max(labelWidth, output.TimeChartYLabelWidth(c.chartSeries(since, until), c.chartOptions(since, until, width, color)))
		}
	}

	for _, c := range charts {
		fmt.Fprintln(w)
		fmt.Fprintln(w, c.title)
		if len(c.base) == 0 {
			fmt.Fprintln(w, cmp.Or(c.empty, "no data"))
			continue
		}

		series := c.chartSeries(since, until)
		opts := c.chartOptions(since, until, width, color)
		opts.YLabelWidth = labelWidth

		fmt.Fprintln(w, output.ChartLegend(series, color))
		fmt.Fprintln(w, output.TimeChart(series, opts))
	}
}

// chartSeries returns the series of c, with a capacity line across the whole
// time range when c has a total.
func (c metricsChart) chartSeries(since, until time.Time) []output.ChartSeries {
	series := make([]output.ChartSeries, 0, len(c.series)+1)
	if c.total > 0 {
		series = append(series, output.ChartSeries{
			Name:   "capacity " + c.yLabel(c.total),
			Color:  output.ChartColorCapacity,
			Points: []output.TimePoint{{Time: since, Value: c.total}, {Time: until, Value: c.total}},
		})
	}

	return append(series, c.series...)
}

func (c metricsChart) chartOptions(since, until time.Time, width int, color bool) output.TimeChartOptions {
	return output.TimeChartOptions{
		Since:  since,
		Until:  until,
		YMax:   c.total,
		YTop:   c.yTop,
		YUnit:  c.yUnit,
		Width:  width,
		Height: metricsChartHeight,
		YLabel: c.yLabel,
		Color:  color,
	}
}
