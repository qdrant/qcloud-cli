package cloudregion

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "cloud-region" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-region",
		Short: "Manage cloud regions",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newListCommand(s))
	return cmd
}
