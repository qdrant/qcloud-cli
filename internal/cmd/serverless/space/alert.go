package space

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newAlertCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Show alerts of serverless spaces",
		Long: `Show alerts of serverless spaces.

Alerts are raised when a space or one of its collections needs attention, for
example when a collection is close to its storage limit, the space is close to
its collection limit, an API key is about to expire, or the space is unhealthy.
An alert is either firing or resolved, and is tied either to a single collection
or to the space as a whole.`,
		Example: `# List all alerts of a space
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# List only the alerts that are currently firing
qcloud serverless space alert list 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 --state firing`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newAlertListCommand(s),
	)
	return cmd
}
