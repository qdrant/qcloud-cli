package cluster

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage API keys for a cluster",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newKeyListCommand(s),
		newKeyCreateCommand(s),
		newKeyDeleteCommand(s),
	)
	return cmd
}
