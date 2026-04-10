package account

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the account command group.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage Qdrant Cloud accounts",
		Long: `Manage Qdrant Cloud accounts and their members.

Use these commands to list, inspect, and update accounts that the current
management key has access to. Account member commands show who belongs to the
current account and whether they are the owner.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newListCommand(s),
		newDescribeCommand(s),
		newUpdateCommand(s),
		newMemberCommand(s),
	)
	return cmd
}
