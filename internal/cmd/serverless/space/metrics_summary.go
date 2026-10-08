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

func newMetricsSummaryCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse]{
		Use:   "summary <space-id>",
		Short: "Show a metrics overview of a space's collections",
		Long: `Show a metrics overview of the collections in a serverless space.

For every collection, the overview reports the current point count, storage
usage and number of search workers, together with the search rate, write rate
and average search latency. The server averages the rates and latency over
several intervals; the table shows the shortest one, which is named above the
table. Use --json to get the averages for all intervals.

The space's quota is shown above the table so that storage and worker usage
can be compared against the per-collection limits.

By default, all collections are fetched automatically across multiple pages.
Use --page-size and --page-token for manual pagination.`,
		Example: `# Show an overview of all collections in a space
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Only show a single collection
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --collection products

# Only show collections whose name contains "staging"
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --collection-contains staging

# Output all interval averages as JSON
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --json`,
		Args: util.ExactArgs(1, "a space ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*serverlessmonitoringv1.GetSpaceSummaryMetricsResponse, error) {
			filter, err := readCollectionFilter(cmd)
			if err != nil {
				return nil, err
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
			items, next, err := util.FetchPages(cmd, func(pageSize *int32, pageToken *string) ([]*serverlessmonitoringv1.SpaceCollectionMetrics, string, error) {
				req := &serverlessmonitoringv1.GetSpaceSummaryMetricsRequest{
					AccountId: accountID,
					SpaceId:   args[0],
					PageSize:  pageSize,
					PageToken: pageToken,
				}
				switch {
				case filter.name != nil:
					req.CollectionFilter = &serverlessmonitoringv1.GetSpaceSummaryMetricsRequest_CollectionName{CollectionName: *filter.name}
				case filter.contains != nil:
					req.CollectionFilter = &serverlessmonitoringv1.GetSpaceSummaryMetricsRequest_CollectionNameContains{CollectionNameContains: *filter.contains}
				}

				resp, err := client.ServerlessMonitoring().GetSpaceSummaryMetrics(ctx, req)
				if err != nil {
					return nil, "", fmt.Errorf("failed to get space summary metrics: %w", err)
				}

				if quota == nil {
					quota = resp.GetQuota()
				}

				return resp.GetItems(), resp.GetNextPageToken(), nil
			})
			if err != nil {
				return nil, err
			}

			return &serverlessmonitoringv1.GetSpaceSummaryMetricsResponse{Items: items, Quota: quota, NextPageToken: next}, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *serverlessmonitoringv1.GetSpaceSummaryMetricsResponse) error {
			printQuotaSnapshot(w, resp.GetQuota())

			interval, hasInterval := shortestOverviewInterval(resp.GetItems())
			if hasInterval {
				fmt.Fprintf(w, "Rates and latency averaged over the last %s.\n\n", output.CompactDuration(interval))
			}

			overview := func(o *serverlessmonitoringv1.SpaceMetricOverview, format func(float64) string) string {
				if !hasInterval {
					return "-"
				}

				for _, a := range o.GetAvg() {
					if a.GetInterval() != nil && a.GetInterval().AsDuration() == interval {
						return format(a.GetValue())
					}
				}

				return "-"
			}

			t := output.NewTable[*serverlessmonitoringv1.SpaceCollectionMetrics](w)
			t.AddField("COLLECTION", func(v *serverlessmonitoringv1.SpaceCollectionMetrics) string {
				return v.GetCollectionName()
			})
			t.AddField("POINTS", func(v *serverlessmonitoringv1.SpaceCollectionMetrics) string {
				return strconv.FormatUint(v.GetPointCount(), 10)
			})
			t.AddField("STORAGE", func(v *serverlessmonitoringv1.SpaceCollectionMetrics) string {
				return output.Bytes(v.GetUsedStorageBytes())
			})
			t.AddField("WORKERS", func(v *serverlessmonitoringv1.SpaceCollectionMetrics) string {
				return strconv.FormatUint(v.GetCurrentSearcherWorkers(), 10)
			})
			t.AddField("SEARCHES", func(v *serverlessmonitoringv1.SpaceCollectionMetrics) string {
				return overview(v.GetSearchRequests(), output.Rate)
			})
			t.AddField("WRITES", func(v *serverlessmonitoringv1.SpaceCollectionMetrics) string {
				return overview(v.GetWriteRequests(), output.Rate)
			})
			t.AddField("LATENCY", func(v *serverlessmonitoringv1.SpaceCollectionMetrics) string {
				return overview(v.GetSearchLatency(), output.Milliseconds)
			})
			t.SetItems(resp.GetItems())
			t.Render()

			return nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)

	addCollectionFilterFlags(cmd, false)
	util.AddPaginationFlags(cmd, "collections")

	return cmd
}

// shortestOverviewInterval returns the shortest averaging interval reported
// for any collection. Averages without an interval are ignored.
func shortestOverviewInterval(items []*serverlessmonitoringv1.SpaceCollectionMetrics) (time.Duration, bool) {
	var shortest time.Duration
	found := false
	for _, item := range items {
		for _, o := range []*serverlessmonitoringv1.SpaceMetricOverview{item.GetSearchRequests(), item.GetWriteRequests(), item.GetSearchLatency()} {
			for _, a := range o.GetAvg() {
				if a.GetInterval() == nil {
					continue
				}

				d := a.GetInterval().AsDuration()
				if !found || d < shortest {
					shortest = d
					found = true
				}
			}
		}
	}

	return shortest, found
}
