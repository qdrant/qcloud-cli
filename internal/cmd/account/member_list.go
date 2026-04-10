package account

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newMemberListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*accountv1.ListAccountMembersResponse]{
		Use:   "list",
		Short: "List account members",
		Long: `List all members of the current account.

Each member has an associated user record and an ownership flag. Use
"qcloud iam user list" to see users with their status, or this command to
see who is in the account and who owns it.`,
		Example: `# List all members
qcloud account member list

# Output as JSON
qcloud account member list --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*accountv1.ListAccountMembersResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Account().ListAccountMembers(ctx, &accountv1.ListAccountMembersRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list account members: %w", err)
			}

			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *accountv1.ListAccountMembersResponse) error {
			t := output.NewTable[*accountv1.AccountMember](w)
			t.AddField("ID", func(v *accountv1.AccountMember) string {
				return v.GetAccountMember().GetId()
			})
			t.AddField("EMAIL", func(v *accountv1.AccountMember) string {
				return v.GetAccountMember().GetEmail()
			})
			t.AddField("OWNER", func(v *accountv1.AccountMember) string {
				return output.BoolMark(v.GetIsOwner())
			})
			t.AddField("CREATED", func(v *accountv1.AccountMember) string {
				if v.GetAccountMember().GetCreatedAt() != nil {
					return output.HumanTime(v.GetAccountMember().GetCreatedAt().AsTime())
				}
				return ""
			})
			t.Write(resp.GetItems())
			return nil
		},
	}.CobraCommand(s)
}
