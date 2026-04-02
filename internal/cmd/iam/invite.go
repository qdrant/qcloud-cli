package iam

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newInviteCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Manage account invites",
		Long: `Manage account invites in Qdrant Cloud.

Provides commands to list, view, and delete account invites.
To send a new invite, use the 'iam user invite' command.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newInviteListCommand(s),
		newInviteDescribeCommand(s),
		newInviteDeleteCommand(s),
	)
	return cmd
}
