package cluster

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newAlertCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Show alerts of clusters",
		Long: `Show alerts of clusters.

Alerts are raised when a cluster needs attention, for example when it is close
to running out of disk or memory, is out of disk or memory, has nodes in
recovery mode, is unhealthy or CPU throttled, has too many collections, has a
resource hotspot, runs a version not covered by the SLA, or has a database API
key that is about to expire. An alert is either firing or resolved.`,
		Example: `# List the alerts of a cluster that are currently firing
qcloud cluster alert list 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# List both firing and resolved alerts
qcloud cluster alert list 7b2ea926-724b-4de2-b73a-8675c42a6ebe --show-resolved`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newAlertListCommand(s),
	)
	return cmd
}
