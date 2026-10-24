package cluster

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

type clusterMetricsHistory struct {
	ClusterID         string         `json:"clusterId"`
	Since             time.Time      `json:"since"`
	Until             time.Time      `json:"until"`
	RequestsPerSecond []historyPoint `json:"requestsPerSecond"`
	LatencyMs         []historyPoint `json:"latencyMs"`
}

func newClusterMetricsHistoryCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*clusterMetricsHistory]{
		Use:   "cluster-metrics-history <cluster-id>",
		Short: "Show the request rate and latency of a cluster over time",
		Long: `Show the request rate and latency of a cluster over time.

The requests per second and the request latency in milliseconds across all
nodes of the cluster are drawn as separate line charts. Request metrics are
measured at the ingress of Qdrant Cloud, so they are not available for Hybrid
Cloud clusters.

By default, the last 7 days are shown. --since and --until accept an RFC3339
timestamp, a YYYY-MM-DD date in UTC, or a duration ago such as 6h or 7d. When
only --until is given, the range covers the 7 days before it. Times on the
charts are in UTC.`,
		Example: `# Show the request rate and latency over the last 7 days
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Show the request rate and latency over the last 6 hours
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe --since 6h

# Show the request rate and latency in a specific time range
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe --since 2026-10-01 --until 2026-10-07

# Output the time series as JSON
qcloud cluster cluster-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json`,
		Args: util.ExactArgs(1, "a cluster ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*clusterMetricsHistory, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			since, until, err := readMetricsHistoryRange(cmd, time.Now().Truncate(time.Second))
			if err != nil {
				return nil, err
			}

			clusterResp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
				AccountId: accountID,
				ClusterId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster: %w", err)
			}

			if clusterResp.GetCluster().GetCloudProviderId() == qcloudapi.HybridCloudProviderID {
				return nil, errors.New("request metrics are not available for Hybrid Cloud clusters")
			}

			metricsResp, err := client.Monitoring().GetClusterUsageMetrics(ctx, &monitoringv1.GetClusterUsageMetricsRequest{
				AccountId: accountID,
				ClusterId: args[0],
				Since:     timestamppb.New(since),
				Until:     timestamppb.New(until),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster metrics: %w", err)
			}

			return &clusterMetricsHistory{
				ClusterID:         args[0],
				Since:             since.UTC(),
				Until:             until.UTC(),
				RequestsPerSecond: historyPoints(metricsResp.GetRps()),
				LatencyMs:         historyPoints(metricsResp.GetLatency()),
			}, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, h *clusterMetricsHistory) error {
			printClusterMetricsHistory(w, h, metricsChartWidth(w), output.ColorEnabled(w))
			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)

	addMetricsHistoryRangeFlags(cmd)

	return cmd
}

func printClusterMetricsHistory(w io.Writer, h *clusterMetricsHistory, width int, color bool) {
	fmt.Fprintf(w, "Cluster %s  %s - %s\n", h.ClusterID, output.FullDateTime(h.Since), output.FullDateTime(h.Until))

	// The API fills the request series with zeros where there were no
	// requests, so all zeros means there is nothing to draw. The latency of
	// no requests is no measurement either.
	requests, latency, empty := h.RequestsPerSecond, h.LatencyMs, ""
	if !slices.ContainsFunc(requests, func(p historyPoint) bool { return p.Value != 0 }) {
		requests, latency, empty = nil, nil, "no requests in this time range"
	}

	charts := []metricsChart{
		{
			title:  "Requests",
			base:   requests,
			empty:  empty,
			series: []output.ChartSeries{{Name: "requests per second", Color: output.ChartColorRequests, Points: stack(requests, 0)}},
			yLabel: func(v float64) string { return output.Significant(v, 3) + "/s" },
		},
		{
			title:  "Latency",
			base:   latency,
			empty:  empty,
			series: []output.ChartSeries{{Name: "latency", Color: output.ChartColorLatency, Points: stack(latency, 0)}},
			yLabel: func(v float64) string { return output.Significant(v, 3) + "ms" },
		},
	}
	printMetricsCharts(w, charts, h.Since, h.Until, width, color)
}
