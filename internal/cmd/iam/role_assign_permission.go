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

func newRoleAssignPermissionCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		ValidArgsFunction: completion.RoleIDCompletion(s),
		Long: `Add permissions to a custom role.

Fetches the role's current permissions, merges the new ones (deduplicating),
and updates the role. Use "qcloud iam permission list" to see available
permissions.`,
		Example: `# Add a single permission
qcloud iam role assign-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe --permission read:clusters

# Add multiple permissions
qcloud iam role assign-permission 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --permission read:clusters --permission read:backups`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "assign-permission <role-id>",
				Short: "Add permissions to a role",
				Args:  util.ExactArgs(1, "a role ID"),
			}
			cmd.Flags().StringSlice("permission", nil, "Permission to add (repeatable)")
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
			newPerms, _ := cmd.Flags().GetStringSlice("permission")

			// Build a set of existing permission values for dedup.
			existing := make(map[string]bool, len(role.GetPermissions()))
			for _, p := range role.GetPermissions() {
				existing[p.GetValue()] = true
			}

			added := 0
			for _, v := range newPerms {
				if !existing[v] {
					role.Permissions = append(role.Permissions, &iamv1.Permission{Value: v})
					existing[v] = true
					added++
				}
			}

			if added == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No new permissions to add.")
				return nil
			}

			_, err = client.IAM().UpdateRole(ctx, &iamv1.UpdateRoleRequest{
				Role: role,
			})
			if err != nil {
				return fmt.Errorf("failed to update role: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Added %d permission(s) to role %s.\n", added, args[0])
			return nil
		},
	}.CobraCommand(s)
}
