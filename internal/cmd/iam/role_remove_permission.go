package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRoleRemovePermissionCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		ValidArgsFunction: completion.RoleIDCompletion(s),
		Long: `Remove permissions from a custom role.

Fetches the role's current permissions, removes the specified ones, and updates
the role. A role must retain at least one permission.`,
		Example: `# Remove a single permission
qcloud iam role remove-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe --permission read:clusters

# Remove multiple permissions
qcloud iam role remove-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --permission read:clusters --permission read:backups`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "remove-permission <role-id>",
				Short: "Remove permissions from a role",
				Args:  util.ExactArgs(1, "a role ID"),
			}
			cmd.Flags().StringSlice("permission", nil, "Permission to remove (repeatable)")
			_ = cmd.MarkFlagRequired("permission")
			_ = cmd.RegisterFlagCompletionFunc("permission", completion.PermissionCompletion(s))
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			resp, err := client.IAM().GetRole(ctx, &iamv1.GetRoleRequest{
				AccountId: accountID,
				RoleId:    args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to get role: %w", err)
			}

			role := resp.GetRole()
			toRemove, _ := cmd.Flags().GetStringSlice("permission")

			removeSet := make(map[string]bool, len(toRemove))
			for _, v := range toRemove {
				removeSet[v] = true
			}

			var kept []*iamv1.Permission
			removed := 0
			for _, p := range role.GetPermissions() {
				if removeSet[p.GetValue()] {
					removed++
				} else {
					kept = append(kept, p)
				}
			}

			if removed == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No matching permissions to remove.")
				return nil
			}

			if len(kept) == 0 {
				return fmt.Errorf("cannot remove all permissions: a role must have at least one permission")
			}

			role.Permissions = kept

			_, err = client.IAM().UpdateRole(ctx, &iamv1.UpdateRoleRequest{
				Role: role,
			})
			if err != nil {
				return fmt.Errorf("failed to update role: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed %d permission(s) from role %s.\n", removed, args[0])
			return nil
		},
	}.CobraCommand(s)
}
