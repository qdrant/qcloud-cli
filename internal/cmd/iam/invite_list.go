package iam

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newInviteListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*accountv1.ListAccountInvitesResponse]{
		Use:   "list",
		Short: "List account invites",
		Long: `List account invites.

Lists all invites for the current account. By default, invites of all statuses
are returned. Requires the read:invites permission.`,
		Example: `# List all invites
qcloud iam invite list

# Output as JSON
qcloud iam invite list --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*accountv1.ListAccountInvitesResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}
			resp, err := client.Account().ListAccountInvites(ctx, &accountv1.ListAccountInvitesRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list invites: %w", err)
			}
			return resp, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *accountv1.ListAccountInvitesResponse) error {
			t := output.NewTable[*accountv1.AccountInvite](w)
			t.AddField("ID", func(v *accountv1.AccountInvite) string { return v.GetId() })
			t.AddField("EMAIL", func(v *accountv1.AccountInvite) string { return v.GetUserEmail() })
			t.AddField("STATUS", func(v *accountv1.AccountInvite) string {
				return output.AccountInviteStatus(v.GetStatus())
			})
			t.AddField("CREATED", func(v *accountv1.AccountInvite) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.Write(resp.GetItems())
			return nil
		},
	}.CobraCommand(s)
}
