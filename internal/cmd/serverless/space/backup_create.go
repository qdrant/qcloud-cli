package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*spacebackupv1.Backup]{
		Long: `Create an on-demand backup of a serverless space.

By default the whole space is backed up; use --collection to back up a single
collection. Without --retention-days the backup is kept until it is deleted.
The backup runs asynchronously; use "qcloud serverless space backup describe" to
follow its status and progress.`,
		Example: `# Back up a whole space
qcloud serverless space backup create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60

# Back up a single collection and keep the backup for 7 days
qcloud serverless space backup create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --collection products --retention-days 7`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a backup of a serverless space",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("space-id", "", "ID of the space to back up (required)")
			cmd.Flags().String("collection", "", "Only back up this collection (default: the whole space)")
			cmd.Flags().Uint32("retention-days", 0, "Retention period in days (1-365) (default: keep indefinitely)")
			_ = cmd.MarkFlagRequired("space-id")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, _ []string) (*spacebackupv1.Backup, error) {
			spaceID, _ := cmd.Flags().GetString("space-id")

			b := &spacebackupv1.Backup{SpaceId: spaceID}

			if cmd.Flags().Changed("collection") {
				collection, _ := cmd.Flags().GetString("collection")
				b.CollectionName = &collection
			}

			if cmd.Flags().Changed("retention-days") {
				days, _ := cmd.Flags().GetUint32("retention-days")
				retention, err := retentionFromDays(days)
				if err != nil {
					return nil, err
				}

				b.RetentionPeriod = retention
			}

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			b.AccountId = accountID

			resp, err := client.ServerlessBackup().CreateBackup(ctx, &spacebackupv1.CreateBackupRequest{Backup: b})
			if err != nil {
				return nil, fmt.Errorf("failed to create backup: %w", err)
			}

			return resp.GetBackup(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, b *spacebackupv1.Backup) {
			fmt.Fprintf(out, "Backup %s created for space %s.\n", b.GetId(), b.GetSpaceId())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("space-id", completion.SpaceIDFlagCompletion(s))
	return cmd
}
