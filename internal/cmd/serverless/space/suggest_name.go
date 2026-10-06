package space

import (
	"fmt"

	"github.com/spf13/cobra"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newSuggestNameCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Suggest a unique, human-friendly name for a new space.

The suggested name is not reserved: it is only guaranteed to be unused in the
current account at the time of the call. "qcloud serverless space create" uses
the same suggestion automatically when --name is omitted.`,
		Example: `# Print a suggested space name
qcloud serverless space suggest-name

# Use the suggestion in a script
qcloud serverless space create --cloud-region eu-central-1 --name "$(qcloud serverless space suggest-name)"`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "suggest-name",
				Short: "Suggest a name for a new space",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			resp, err := client.ServerlessSpace().SuggestSpaceName(ctx, &spacev1.SuggestSpaceNameRequest{
				AccountId: accountID,
			})
			if err != nil {
				return fmt.Errorf("failed to suggest a space name: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), resp.GetName())
			return nil
		},
	}.CobraCommand(s)
}
