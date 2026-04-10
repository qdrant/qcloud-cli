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

func newUserListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*iamv1.ListUsersResponse]{
		Use:   "list",
		Short: "List users in the account",
		Long: `List users in the account.

Lists all users who are members of the current account. Requires the read:users
permission.`,
		Example: `# List all users in the account
qcloud iam user list

# Output as JSON
qcloud iam user list --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*iamv1.ListUsersResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}
			resp, err := client.IAM().ListUsers(ctx, &iamv1.ListUsersRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list users: %w", err)
			}
			return resp, nil
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *iamv1.ListUsersResponse) (output.TableRenderer, error) {
			t := output.NewTable[*iamv1.User](w)
			t.AddField("ID", func(v *iamv1.User) string { return v.GetId() })
			t.AddField("EMAIL", func(v *iamv1.User) string { return v.GetEmail() })
			t.AddField("STATUS", func(v *iamv1.User) string { return output.UserStatus(v.GetStatus()) })
			t.AddField("CREATED", func(v *iamv1.User) string {
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
