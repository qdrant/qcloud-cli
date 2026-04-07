package iam

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUserCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users in Qdrant Cloud",
		Long: `Manage users in the Qdrant Cloud account.

Provides commands to list users, view user details and assigned roles, and
manage role assignments.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newUserListCommand(s),
		newUserDescribeCommand(s),
		newUserAssignRoleCommand(s),
		newUserRemoveRoleCommand(s),
	)
	return cmd
}
