package hybrid

import (
	"fmt"

	"github.com/spf13/cobra"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDeleteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		ValidArgsFunction: envIDCompletion(s),
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "delete <env-id>",
				Short: "Delete a hybrid cloud environment",
				Args:  util.ExactArgs(1, "a hybrid cloud environment ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			envID := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !util.ConfirmAction(force, cmd.ErrOrStderr(), fmt.Sprintf("Are you sure you want to delete hybrid cloud environment %s?", envID)) {
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

			_, err = client.Hybrid().DeleteHybridCloudEnvironment(ctx, &hybridv1.DeleteHybridCloudEnvironmentRequest{
				AccountId:                accountID,
				HybridCloudEnvironmentId: envID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete hybrid cloud environment: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Hybrid cloud environment %s deleted.\n", envID)
			return nil
		},
	}.CobraCommand(s)
}
