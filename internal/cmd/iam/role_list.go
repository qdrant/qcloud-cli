package iam

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRoleListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*iamv1.ListRolesResponse]{
		Use:   "list",
		Short: "List all roles",
		Long: `List all roles for the account, including both system and custom roles.

System roles are managed by Qdrant and cannot be modified. Custom roles are
created and managed by the account administrator.`,
		Example: `# List all roles
qcloud iam role list

# Output as JSON
qcloud iam role list --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*iamv1.ListRolesResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			return client.IAM().ListRoles(ctx, &iamv1.ListRolesRequest{
				AccountId: accountID,
			})
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *iamv1.ListRolesResponse) (output.TableRenderer, error) {
			t := output.NewTable[*iamv1.Role](w)
			t.AddField("ID", func(v *iamv1.Role) string {
				return v.GetId()
			})
			t.AddField("NAME", func(v *iamv1.Role) string {
				return v.GetName()
			})
			t.AddField("TYPE", func(v *iamv1.Role) string {
				return output.RoleType(v.GetRoleType())
			})
			t.AddField("PERMISSIONS", func(v *iamv1.Role) string {
				return fmt.Sprintf("%d", len(v.GetPermissions()))
			})
			t.AddField("CREATED", func(v *iamv1.Role) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)
}
