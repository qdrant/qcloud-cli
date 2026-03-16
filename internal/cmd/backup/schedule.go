package backup

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newScheduleCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage backup schedules",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newScheduleListCommand(s),
		newScheduleDescribeCommand(s),
		newScheduleCreateCommand(s),
		newScheduleUpdateCommand(s),
		newScheduleDeleteCommand(s),
	)
	return cmd
}
