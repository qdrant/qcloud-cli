package iam

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the iam command group.
func NewCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage IAM resources in Qdrant Cloud",
		Long:  `Manage IAM resources for the Qdrant Cloud account.`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(
		newKeyCommand(s),
		newUserCommand(s),
		newInviteCommand(s),
	)
	return cmd
}
