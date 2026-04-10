package account

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newMemberCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage account members",
		Long: `Manage members of the current Qdrant Cloud account.

Members are users who have been added to the account. Each member has an
associated user record and an ownership flag indicating whether they are the
account owner.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newMemberListCommand(s),
		newMemberDescribeCommand(s),
	)
	return cmd
}
