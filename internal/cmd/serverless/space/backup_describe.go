package space

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBackupDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*spacebackupv1.Backup]{
		Use:   "describe <backup-id>",
		Short: "Describe a backup of a serverless space",
		Long: `Describe a backup of a serverless space.

Shows the backup status and statistics together with a snapshot of the space
(name, region and configuration) taken when the backup was created. That snapshot
is used when a new space is created from the backup.`,
		Example: `# Describe a backup
qcloud serverless space backup describe 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d

# Output as JSON
qcloud serverless space backup describe 9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d --json`,
		Args:              util.ExactArgs(1, "a backup ID"),
		ValidArgsFunction: completion.SpaceBackupIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*spacebackupv1.Backup, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.ServerlessBackup().GetBackup(ctx, &spacebackupv1.GetBackupRequest{
				AccountId: accountID,
				BackupId:  args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get backup: %w", err)
			}

			return resp.GetBackup(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, b *spacebackupv1.Backup) error {
			fmt.Fprintf(w, "ID:          %s\n", b.GetId())
			fmt.Fprintf(w, "Name:        %s\n", b.GetName())
			fmt.Fprintf(w, "Space:       %s\n", b.GetSpaceId())
			fmt.Fprintf(w, "Collection:  %s\n", formatCollection(b.CollectionName))
			fmt.Fprintf(w, "Status:      %s\n", output.SpaceBackupStatus(b.GetStatus()))
			if b.GetCreatedAt() != nil {
				t := b.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:     %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}

			if b.BackupScheduleId != nil {
				fmt.Fprintf(w, "Schedule:    %s\n", b.GetBackupScheduleId())
			}

			fmt.Fprintf(w, "Retention:   %s\n", formatRetention(b.GetRetentionPeriod()))

			if st := b.GetStats(); st != nil {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "Statistics:")
				if st.CollectionCount != nil {
					fmt.Fprintf(w, "  Collections: %d\n", st.GetCollectionCount())
				}

				if st.SizeBytes != nil {
					fmt.Fprintf(w, "  Size:        %s\n", formatOptionalBytes(st.SizeBytes))
				}

				if st.TotalPoints != nil {
					fmt.Fprintf(w, "  Points:      %d\n", st.GetTotalPoints())
				}

				if st.Duration != nil {
					fmt.Fprintf(w, "  Duration:    %s\n", formatOptionalDuration(st.GetDuration()))
				}

				if st.Progress != nil {
					fmt.Fprintf(w, "  Progress:    %s\n", st.GetProgress())
				}
			}

			if info := b.GetSpaceInfo(); info != nil {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "Space Snapshot:")
				fmt.Fprintf(w, "  Name:   %s\n", info.GetName())
				fmt.Fprintf(w, "  Region: %s\n", info.GetCloudRegionId())
				if cfg := info.GetConfiguration(); cfg != nil {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Space Snapshot Configuration:")
					printSpaceConfiguration(w, cfg)
				}
			}

			return nil
		},
	}.CobraCommand(s)
}
