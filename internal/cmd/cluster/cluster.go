package cluster

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "cluster" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Qdrant Cloud clusters",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newDescribeCommand(s),
		newCreateCommand(s),
		newUpdateCommand(s),
		newDeleteCommand(s),
		newWaitCommand(s),
		newPackageCommand(s),
		newCloudProviderCommand(s),
		newCloudRegionCommand(s),
		newKeyCommand(s),
	)
	return cmd
}
