package backup

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "backup" parent command and registers all subcommands.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage Qdrant Cloud backups",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newDescribeCommand(s),
		newCreateCommand(s),
		newDeleteCommand(s),
		newRestoreCommand(s),
		newScheduleCommand(s),
	)
	return cmd
}
