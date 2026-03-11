package context

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "context" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage named configuration contexts",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newUseCommand(s),
		newShowCommand(s),
		newSetCommand(s),
		newDeleteCommand(s),
	)
	return cmd
}
