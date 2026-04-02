package access

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage cloud management keys",
		Long: `Manage cloud management keys for the account.

Management keys authenticate requests to the Qdrant Cloud API. Use them to authorize
the CLI, automation scripts, or any other tooling that calls the Qdrant Cloud API.`,
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(
		newKeyListCommand(s),
		newKeyCreateCommand(s),
		newKeyDeleteCommand(s),
	)
	return cmd
}
