package iam

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newInviteDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*accountv1.AccountInvite]{
		Use:               "describe <invite-id>",
		Short:             "Describe an account invite",
		ValidArgsFunction: inviteCompletion(s),
		Long: `Describe an account invite.

Displays the full details of a specific account invite, including the invited
email address, assigned roles, and current status. Requires the read:invites
permission.`,
		Example: `# Describe an invite by ID
qcloud iam invite describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output as JSON
qcloud iam invite describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json`,
		Args: util.ExactArgs(1, "an invite ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*accountv1.AccountInvite, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}
			resp, err := client.Account().GetAccountInvite(ctx, &accountv1.GetAccountInviteRequest{
				AccountId: accountID,
				InviteId:  args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get invite: %w", err)
			}
			return resp.GetAccountInvite(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, inv *accountv1.AccountInvite) error {
			fmt.Fprintf(w, "ID:      %s\n", inv.GetId())
			fmt.Fprintf(w, "Email:   %s\n", inv.GetUserEmail())
			fmt.Fprintf(w, "Status:  %s\n", output.AccountInviteStatus(inv.GetStatus()))
			if len(inv.GetUserRoleIds()) > 0 {
				fmt.Fprintf(w, "Roles:   %s\n", strings.Join(inv.GetUserRoleIds(), ", "))
			}
			if inv.GetCreatedAt() != nil {
				t := inv.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created: %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			return nil
		},
	}.CobraCommand(s)
}
