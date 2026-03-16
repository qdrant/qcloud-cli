package backup

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRestoreCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Manage backup restores",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newRestoreListCommand(s),
		newRestoreTriggerCommand(s),
	)
	return cmd
}
