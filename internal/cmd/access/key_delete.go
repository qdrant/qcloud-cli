package access

import (
	"fmt"

	"github.com/spf13/cobra"

	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Delete a cloud management key from the account.

Deleting a key immediately revokes its access to the Qdrant Cloud API. Any client
using the deleted key will receive authentication errors. This action cannot be undone.

A confirmation prompt is shown unless --force is passed.`,
		Example: `# Delete a management key (with confirmation prompt)
qcloud access key delete a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Delete without confirmation
qcloud access key delete a1b2c3d4-e5f6-7890-abcd-ef1234567890 --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <key-id>",
				Short: "Delete a cloud management key",
				Args:  util.ExactArgs(1, "a management key ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		ValidArgsFunction: completion.ManagementKeyIDCompletion(s),
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			keyID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to delete management key %s?", keyID)) {
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

			_, err = client.Auth().DeleteManagementKey(ctx, &authv1.DeleteManagementKeyRequest{
				AccountId:       accountID,
				ManagementKeyId: keyID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete management key: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Management key %s deleted.\n", keyID)
			return nil
		},
	}.CobraCommand(s)
}
