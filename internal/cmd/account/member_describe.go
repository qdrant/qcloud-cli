package account

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newMemberDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*accountv1.AccountMember]{
		Use:   "describe <user-id>",
		Short: "Describe an account member",
		Long: `Describe a member of the current account by their user ID.

Shows the member's user details and whether they are the account owner.`,
		Example: `# Describe a member
qcloud account member describe a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Output as JSON
qcloud account member describe a1b2c3d4-e5f6-7890-abcd-ef1234567890 --json`,
		Args: util.ExactArgs(1, "a user ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*accountv1.AccountMember, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Account().GetAccountMember(ctx, &accountv1.GetAccountMemberRequest{
				AccountId: accountID,
				UserId:    args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get account member: %w", err)
			}

			return resp.GetAccountMember(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, member *accountv1.AccountMember) error {
			user := member.GetAccountMember()
			fmt.Fprintf(w, "ID:      %s\n", user.GetId())
			fmt.Fprintf(w, "Email:   %s\n", user.GetEmail())
			fmt.Fprintf(w, "Owner:   %s\n", output.BoolYesNo(member.GetIsOwner()))
			if user.GetCreatedAt() != nil {
				t := user.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created: %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			return nil
		},
		ValidArgsFunction: completion.AccountMemberIDCompletion(s),
	}.CobraCommand(s)
}
