package hybrid

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newClusterCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage clusters in hybrid cloud environments",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newClusterListCommand(s),
		newClusterDescribeCommand(s),
		newClusterCreateCommand(s),
		newClusterUpdateCommand(s),
		newClusterDeleteCommand(s),
	)
	return cmd
}
