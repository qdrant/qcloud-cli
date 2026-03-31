package hybrid

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "hybrid" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hybrid",
		Short: "Manage hybrid cloud environments",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newDescribeCommand(s),
		newCreateCommand(s),
		newUpdateCommand(s),
		newDeleteCommand(s),
		newBootstrapCommand(s),
	)
	return cmd
}
