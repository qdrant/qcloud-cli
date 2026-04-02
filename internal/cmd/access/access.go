package access

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the access command group.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "Manage access to Qdrant Cloud",
		Long:  `Manage access settings for the Qdrant Cloud account.`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newKeyCommand(s))
	return cmd
}
