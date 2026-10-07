package space

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	spacev1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateFromBackupCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*spacev1.Space]{
		Long: `Create a new serverless space seeded with the data from an existing backup.

The new space is provisioned in the same region and with the same configuration
as the original space at the time the backup was taken. The backup must belong
to the current account. To restore a backup into its original space instead, use
"qcloud serverless space backup restore trigger".`,
		Example: `# Create a space from a backup
qcloud serverless space create-from-backup --backup-id 9d8c7b6a-5e4f-4a3b-8c2d-1e0f9a8b7c6d --name my-restored-space

# Create a space from a backup and wait until it is ready
qcloud serverless space create-from-backup --backup-id 9d8c7b6a-5e4f-4a3b-8c2d-1e0f9a8b7c6d --name my-restored-space --wait`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create-from-backup",
				Short: "Create a new space from a backup",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("backup-id", "", "ID of the backup to restore from (required)")
			cmd.Flags().String("name", "", "Name for the new space (required)")
			util.AddWaitFlags(cmd, "the space to become ready", 10*time.Minute, 5*time.Second)
			_ = cmd.MarkFlagRequired("backup-id")
			_ = cmd.MarkFlagRequired("name")
			_ = cmd.RegisterFlagCompletionFunc("backup-id", completion.SpaceBackupIDCompletion(s))
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*spacev1.Space, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			backupID, _ := cmd.Flags().GetString("backup-id")
			name, _ := cmd.Flags().GetString("name")

			resp, err := client.ServerlessSpace().CreateSpaceFromBackup(ctx, &spacev1.CreateSpaceFromBackupRequest{
				AccountId: accountID,
				BackupId:  backupID,
				SpaceName: name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create space from backup: %w", err)
			}

			created := resp.GetSpace()

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return created, nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")
			fmt.Fprintf(cmd.ErrOrStderr(), "Space %s created, waiting for it to become ready...\n", created.GetId())
			ready, err := waitForSpaceReady(ctx, client.ServerlessSpace(), cmd.ErrOrStderr(), accountID, created.GetId(), waitTimeout, pollInterval)
			if err != nil {
				if s.Config.JSONOutput() {
					_ = output.PrintJSON(cmd.OutOrStdout(), created)
				} else {
					fmt.Fprint(cmd.OutOrStdout(), spaceResultMessage(created, "created from backup"))
				}

				return nil, fmt.Errorf("space %s was created but did not become ready: %w", created.GetId(), err)
			}

			return ready, nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *spacev1.Space) {
			fmt.Fprint(out, spaceResultMessage(created, "created from backup"))
		},
	}.CobraCommand(s)
}
