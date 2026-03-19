package cloudprovider

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "cloud-provider" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-provider",
		Short: "Manage cloud providers",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newListCommand(s))
	return cmd
}
