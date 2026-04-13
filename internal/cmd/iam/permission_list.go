package iam

import (
	"io"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newPermissionListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*iamv1.ListPermissionsResponse]{
		Use:   "list",
		Short: "List all available permissions",
		Long: `List all permissions known in the system for the account.

Permissions are the individual access rights that can be assigned to roles.
Each permission has a value (e.g. "read:clusters") and a category
(e.g. "Cluster").`,
		Example: `# List all available permissions
qcloud iam permission list

# Output as JSON
qcloud iam permission list --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*iamv1.ListPermissionsResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			return client.IAM().ListPermissions(ctx, &iamv1.ListPermissionsRequest{
				AccountId: accountID,
			})
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *iamv1.ListPermissionsResponse) (output.TableRenderer, error) {
			t := output.NewTable[*iamv1.Permission](w)
			t.AddField("PERMISSION", func(v *iamv1.Permission) string {
				return v.GetValue()
			})
			t.AddField("CATEGORY", func(v *iamv1.Permission) string {
				return v.GetCategory()
			})
			t.SetItems(resp.GetPermissions())
			return t, nil
		},
	}.CobraCommand(s)
}
