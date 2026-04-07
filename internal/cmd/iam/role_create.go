package iam

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRoleCreateCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*iamv1.Role]{
		Long: `Create a new custom role for the account.

Custom roles allow fine-grained access control by combining specific permissions.
Use "qcloud iam permission list" to see available permissions.`,
		Example: `# Create a role with specific permissions
qcloud iam role create --name "Cluster Viewer" --permission read:clusters --permission read:cluster-endpoints

# Create a role with a description
qcloud iam role create --name "Backup Manager" --description "Can manage backups" \
  --permission read:clusters --permission read:backups --permission write:backups`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a custom role",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("name", "", "Name of the role (4-64 characters)")
			cmd.Flags().String("description", "", "Description of the role")
			cmd.Flags().StringSlice("permission", nil, "Permission to assign (repeatable)")
			_ = cmd.MarkFlagRequired("name")
			_ = cmd.MarkFlagRequired("permission")
			_ = cmd.RegisterFlagCompletionFunc("permission", completion.PermissionCompletion(s))
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*iamv1.Role, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")
			permValues, _ := cmd.Flags().GetStringSlice("permission")

			permissions := make([]*iamv1.Permission, len(permValues))
			for i, v := range permValues {
				permissions[i] = &iamv1.Permission{Value: v}
			}

			resp, err := client.IAM().CreateRole(ctx, &iamv1.CreateRoleRequest{
				Role: &iamv1.Role{
					AccountId:   accountID,
					Name:        name,
					Description: description,
					RoleType:    iamv1.RoleType_ROLE_TYPE_CUSTOM,
					Permissions: permissions,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create role: %w", err)
			}

			return resp.GetRole(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, role *iamv1.Role) {
			fmt.Fprintf(out, "Role %s (%s) created.\n", role.GetId(), role.GetName())
		},
	}.CobraCommand(s)
}
