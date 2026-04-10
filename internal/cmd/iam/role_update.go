package iam

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRoleUpdateCommand(s *state.State) *cobra.Command {
	return base.UpdateCmd[*iamv1.Role]{
		Long: `Update the name or description of a custom role.

Only custom roles can be updated. System roles are managed by Qdrant and cannot
be modified. To change a role's permissions, use the assign-permission and
remove-permission subcommands.`,
		Example: `# Rename a role
qcloud iam role update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --name "New Name"

# Update the description
qcloud iam role update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --description "Updated description"`,
		ValidArgsFunction: completion.RoleIDCompletion(s),
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <role-id>",
				Short: "Update a custom role",
				Args:  util.ExactArgs(1, "a role ID"),
			}
			cmd.Flags().String("name", "", "New name for the role")
			cmd.Flags().String("description", "", "New description for the role")
			return cmd
		},
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*iamv1.Role, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.IAM().GetRole(ctx, &iamv1.GetRoleRequest{
				AccountId: accountID,
				RoleId:    args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get role: %w", err)
			}

			return resp.GetRole(), nil
		},
		Update: func(s *state.State, cmd *cobra.Command, role *iamv1.Role) (*iamv1.Role, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				role.Name = name
			}
			if cmd.Flags().Changed("description") {
				description, _ := cmd.Flags().GetString("description")
				role.Description = description
			}

			resp, err := client.IAM().UpdateRole(ctx, &iamv1.UpdateRoleRequest{
				Role: role,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update role: %w", err)
			}

			return resp.GetRole(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, role *iamv1.Role) {
			if role == nil {
				return
			}
			fmt.Fprintf(out, "Role %s (%s) updated.\n", role.GetId(), role.GetName())
		},
	}.CobraCommand(s)
}
