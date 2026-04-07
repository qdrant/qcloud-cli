package iam

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRoleCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role",
		Short: "Manage roles in Qdrant Cloud",
		Long: `Manage roles for the Qdrant Cloud account.

Roles define sets of permissions that control access to resources. There are two
types of roles: system roles (immutable, managed by Qdrant) and custom roles
(created and managed by the account). Use these commands to list, inspect, create,
update, and delete custom roles, as well as manage their permissions.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newRoleListCommand(s))
	cmd.AddCommand(newRoleDescribeCommand(s))
	cmd.AddCommand(newRoleCreateCommand(s))
	cmd.AddCommand(newRoleUpdateCommand(s))
	cmd.AddCommand(newRoleDeleteCommand(s))
	cmd.AddCommand(newRoleAssignPermissionCommand(s))
	cmd.AddCommand(newRoleRemovePermissionCommand(s))
	return cmd
}
