package cluster

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newMetricsCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show resource usage and request metrics of a cluster",
		Long: `Show resource usage and request metrics of a cluster.

The summary shows the current CPU, RAM and disk usage of each node. The history
commands draw how the request rate and latency of the whole cluster, or the
resource usage of a single node, changed over a time range.`,
		Example: `# Show the current resource usage of every node of a cluster
qcloud cluster metrics summary 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Show the request rate and latency of a cluster over the last 7 days
qcloud cluster metrics history 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Show the resource usage of node 0 over the last 24 hours
qcloud cluster metrics node-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0 --since 24h`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newMetricsSummaryCommand(s),
		newMetricsHistoryCommand(s),
		newMetricsNodeHistoryCommand(s),
	)
	return cmd
}
