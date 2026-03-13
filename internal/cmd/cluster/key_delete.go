package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	clusterauthv2 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/auth/v2"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <cluster-id> <key-id>",
				Short: "Delete an API key from a cluster",
				Args:  util.ExactArgs(2, "a cluster ID and a key ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			clusterID := args[0]
			keyID := args[1]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, fmt.Sprintf("Are you sure you want to delete API key %s?", keyID)) {
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

			_, err = client.DatabaseApiKey().DeleteDatabaseApiKey(ctx, &clusterauthv2.DeleteDatabaseApiKeyRequest{
				AccountId:        accountID,
				ClusterId:        clusterID,
				DatabaseApiKeyId: keyID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete API key: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "API key %s deleted.\n", keyID)
			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
