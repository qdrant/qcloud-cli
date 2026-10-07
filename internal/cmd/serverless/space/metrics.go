package space

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newMetricsCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show metrics of a space's collections",
		Long: `Show metrics of the collections in a serverless space.

Metrics are reported per collection: request rates, search latency, point
counts, storage usage and the number of search workers. The summary command
gives a current overview, while the usage command reports time series over a
chosen period. Both also report the space's quota so that usage can be compared
against the limits.`,
		Example: `# Show a current overview of all collections in a space
qcloud serverless space metrics summary 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Show usage over the last 24 hours
qcloud serverless space metrics usage 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --since 24h`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newMetricsSummaryCommand(s),
		newMetricsUsageCommand(s),
	)
	return cmd
}
