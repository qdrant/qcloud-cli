package cluster

import (
	"errors"
	"fmt"
	"io"
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
// before --until, and --until to now.
func readMetricsHistoryRange(cmd *cobra.Command, now time.Time) (time.Time, time.Time, error) {
	sinceTS, untilTS, err := util.ReadTimeRange(cmd, now)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	until := now
	if untilTS != nil {
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

func historyPoints(metrics []*monitoringv1.Metric) []historyPoint {
	points := make([]historyPoint, 0, len(metrics))
	for _, m := range metrics {
		points = append(points, historyPoint{Timestamp: m.GetTimestamp().AsTime().UTC(), Value: m.GetValue()})
	}

	return points
}

func pointsByTime(points []historyPoint) map[int64]float64 {
	byTime := make(map[int64]float64, len(points))
	for _, p := range points {
		byTime[p.Timestamp.Unix()] = p.Value
	}

	return byTime
}

// stack returns the points of base with the value at the same time in each of
// layers, and offset, added on top.
func stack(base []historyPoint, offset float64, layers ...[]historyPoint) []output.TimePoint {
	lookups := make([]map[int64]float64, len(layers))
	for i, l := range layers {
		lookups[i] = pointsByTime(l)
	}

	points := make([]output.TimePoint, 0, len(base))
	for _, p := range base {
		v := offset + p.Value
		for _, l := range lookups {
			v += l[p.Timestamp.Unix()]
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
	title  string
	base   []historyPoint
	total  float64
	series []output.ChartSeries
	yLabel func(float64) string
	yUnit  float64
}

// metricsChartWidth returns the width of the charts written to w: the width
// of the terminal, within limits.
func metricsChartWidth(w io.Writer) int {
	return min(output.TerminalWidth(w, metricsChartDefaultWidth), metricsChartMaxWidth)
}

// printMetricsChart prints the title, legend and chart of c, with a capacity
// line when c has a total.
func printMetricsChart(w io.Writer, c metricsChart, since, until time.Time, width int, color bool) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, c.title)
	if len(c.base) == 0 {
		fmt.Fprintln(w, "no data")
		return
	}

	series := make([]output.ChartSeries, 0, len(c.series)+1)
	if c.total > 0 {
		series = append(series, output.ChartSeries{
			Name:   "capacity " + c.yLabel(c.total),
			Color:  output.ChartColorCapacity,
			Points: constant(c.base, c.total),
		})
	}

	series = append(series, c.series...)

	fmt.Fprintln(w, output.ChartLegend(series, color))
	fmt.Fprintln(w, output.TimeChart(series, output.TimeChartOptions{
		Since:  since,
		Until:  until,
		YMax:   c.total,
		YUnit:  c.yUnit,
		Width:  width,
		Height: metricsChartHeight,
		YLabel: c.yLabel,
		Color:  color,
	}))
}
