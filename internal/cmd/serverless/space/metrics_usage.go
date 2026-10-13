package space

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	serverlessmonitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newMetricsUsageCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*serverlessmonitoringv1.GetSpaceUsageMetricsResponse]{
		Use:   "usage <space-id>",
		Short: "Show usage metrics of a space's collections over time",
		Long: `Show usage metrics of the collections in a serverless space over a period.

The server returns a time series per collection for the search rate, write
rate, search latency, point count, storage usage and number of search workers.
The table condenses each series: point count, storage and workers show the
latest value, while request rates and latency show the average and maximum over
the period. Use --json to get the full time series.

When --since is omitted, the period covers the last hour.

By default, all collections are fetched automatically across multiple pages.
Use --page-size and --page-token for manual pagination.`,
		Example: `# Show usage over the last hour
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Show usage over the last 24 hours
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 24h

# Show usage of a single collection on a specific day
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --collection products --since 2026-10-01 --until 2026-10-01

# Export the full time series as JSON
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 7d --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*serverlessmonitoringv1.GetSpaceUsageMetricsResponse, error) {
			since, until, err := util.ReadTimeRange(cmd, time.Now())
			if err != nil {
				return nil, err
			}

			filter, err := readCollectionFilter(cmd)
			if err != nil {
				return nil, err
			}

			var aggregator *serverlessmonitoringv1.Aggregator
			if cmd.Flags().Changed("aggregator") {
				v, _ := cmd.Flags().GetString("aggregator")
				a, err := parseAggregator(v)
				if err != nil {
					return nil, err
				}

				aggregator = &a
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			var quota *serverlessmonitoringv1.SpaceQuotaSnapshot
			items, next, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*serverlessmonitoringv1.SpaceCollectionUsageMetrics, string, error) {
				req := &serverlessmonitoringv1.GetSpaceUsageMetricsRequest{
					AccountId:  accountID,
					SpaceId:    args[0],
					Since:      since,
					Until:      until,
					Aggregator: aggregator,
					PageSize:   pageSize,
					PageToken:  pageToken,
				}
				switch {
				case filter.name != nil:
					req.CollectionFilter = &serverlessmonitoringv1.GetSpaceUsageMetricsRequest_CollectionName{CollectionName: *filter.name}
				case filter.contains != nil:
					req.CollectionFilter = &serverlessmonitoringv1.GetSpaceUsageMetricsRequest_CollectionNameContains{CollectionNameContains: *filter.contains}
				}

				resp, err := client.ServerlessMonitoring().GetSpaceUsageMetrics(ctx, req)
				if err != nil {
					return nil, "", fmt.Errorf("failed to get space usage metrics: %w", err)
				}

				if quota == nil {
					quota = resp.GetQuota()
				}

				return resp.GetItems(), resp.GetNextPageToken(), nil
			})
			if err != nil {
				return nil, err
			}

			return &serverlessmonitoringv1.GetSpaceUsageMetricsResponse{Items: items, Quota: quota, NextPageToken: next}, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *serverlessmonitoringv1.GetSpaceUsageMetricsResponse) error {
			printQuotaSnapshot(w, resp.GetQuota())

			if first, last, ok := usagePeriod(resp.GetItems()); ok {
				fmt.Fprintf(w, "Period: %s to %s\n\n", output.FullDateTime(first), output.FullDateTime(last))
			}

			type row struct {
				name                      string
				points, storage, workers  seriesRollup
				searches, writes, latency seriesRollup
			}

			rows := make([]row, 0, len(resp.GetItems()))
			for _, item := range resp.GetItems() {
				rows = append(rows, row{
					name:     item.GetCollectionName(),
					points:   rollup(item.GetPointCount()),
					storage:  rollup(item.GetUsedStorageBytes()),
					workers:  rollup(item.GetCurrentSearcherWorkers()),
					searches: rollup(item.GetSearchRequests()),
					writes:   rollup(item.GetWriteRequests()),
					latency:  rollup(item.GetSearchLatency()),
				})
			}

			value := func(r seriesRollup, v float64, format func(float64) string) string {
				if !r.ok {
					return "-"
				}

				return format(v)
			}
			count := func(v float64) string { return strconv.FormatUint(uint64(v), 10) }
			bytes := func(v float64) string { return output.Bytes(uint64(v)) }

			t := output.NewTable[row](w)
			t.AddField("COLLECTION", func(r row) string { return r.name })
			t.AddField("POINTS", func(r row) string { return value(r.points, r.points.latest, count) })
			t.AddField("STORAGE", func(r row) string { return value(r.storage, r.storage.latest, bytes) })
			t.AddField("WORKERS", func(r row) string { return value(r.workers, r.workers.latest, count) })
			t.AddField("SEARCHES AVG", func(r row) string { return value(r.searches, r.searches.avg, output.Rate) })
			t.AddField("SEARCHES MAX", func(r row) string { return value(r.searches, r.searches.max, output.Rate) })
			t.AddField("WRITES AVG", func(r row) string { return value(r.writes, r.writes.avg, output.Rate) })
			t.AddField("WRITES MAX", func(r row) string { return value(r.writes, r.writes.max, output.Rate) })
			t.AddField("LATENCY AVG", func(r row) string { return value(r.latency, r.latency.avg, output.Milliseconds) })
			t.AddField("LATENCY MAX", func(r row) string { return value(r.latency, r.latency.max, output.Milliseconds) })
			t.SetItems(rows)
			t.Render()

			return nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)

	addTimeRangeFlags(cmd, "1h ago")
	cmd.Flags().String("aggregator", "", "Aggregation function applied to the time series (sum, avg, max, min; default sum)")
	_ = cmd.RegisterFlagCompletionFunc("aggregator", cobra.FixedCompletions([]string{"sum", "avg", "max", "min"}, cobra.ShellCompDirectiveNoFileComp))
	addCollectionFilterFlags(cmd, false)
	util.AddPaginationFlags(cmd, "collections")

	return cmd
}

// usagePeriod returns the earliest and latest timestamps across all series.
func usagePeriod(items []*serverlessmonitoringv1.SpaceCollectionUsageMetrics) (first, last time.Time, ok bool) {
	for _, item := range items {
		for _, series := range [][]*serverlessmonitoringv1.Metric{
			item.GetSearchRequests(), item.GetWriteRequests(), item.GetSearchLatency(),
			item.GetPointCount(), item.GetUsedStorageBytes(), item.GetCurrentSearcherWorkers(),
		} {
			for _, m := range series {
				if m.GetTimestamp() == nil {
					continue
				}

				ts := m.GetTimestamp().AsTime()
				if !ok || ts.Before(first) {
					first = ts
				}

				if !ok || ts.After(last) {
					last = ts
				}

				ok = true
			}
		}
	}

	return first, last, ok
}
