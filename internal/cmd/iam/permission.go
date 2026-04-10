package iam

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newPermissionCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permission",
		Short: "Manage permissions in Qdrant Cloud",
		Long: `Manage permissions for the Qdrant Cloud account.

Permissions represent individual access rights that can be assigned to roles.
Use these commands to discover which permissions are available in the system.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newPermissionListCommand(s))
	return cmd
}
