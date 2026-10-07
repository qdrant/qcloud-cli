package space

import (
	"fmt"

	"github.com/spf13/cobra"

	spaceauthv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Delete an API key from a serverless space.

Requests authenticated with the key are rejected once the deletion has been
propagated to the space. Deletion cannot be undone.`,
		Example: `# Delete an API key (prompts for confirmation)
qcloud serverless space key delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Delete an API key without confirmation
qcloud serverless space key delete 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 a1b2c3d4-e5f6-7890-abcd-ef1234567890 --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <space-id> <key-id>",
				Short: "Delete an API key from a space",
				Args:  util.ExactArgs(2, "a space ID and a key ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			spaceID := args[0]
			keyID := args[1]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to delete API key %s?", keyID)) {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			_, err = client.ServerlessSpaceApiKey().DeleteSpaceApiKey(ctx, &spaceauthv1.DeleteSpaceApiKeyRequest{
				AccountId:     accountID,
				SpaceId:       spaceID,
				SpaceApiKeyId: keyID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete API key: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "API key %s deleted.\n", keyID)
			return nil
		},
		ValidArgsFunction: completion.SpaceIDCompletion(s),
	}.CobraCommand(s)
}
