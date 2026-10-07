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

func newBackupScheduleCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*spacebackupv1.BackupSchedule]{
		Long: `Create a backup schedule for a serverless space.

The --schedule flag takes a standard cron expression evaluated in UTC (for example
"0 2 * * *" for every day at 02:00), or a descriptor such as "@daily". By default
the whole space is backed up; use --collection to back up a single collection.
Without --retention-days, created backups are kept until they are deleted.`,
		Example: `# Back up a space every night at 02:00 UTC
qcloud serverless space backup schedule create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name nightly --schedule "0 2 * * *"

# Back up a single collection every hour and keep backups for 7 days
qcloud serverless space backup schedule create --space-id 0e7a3c1d-5f2b-4c8e-9a6d-1b2c3d4e5f60 \
  --name products-hourly --schedule @hourly --collection products --retention-days 7`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a backup schedule for a serverless space",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("space-id", "", "ID of the space to back up (required)")
			cmd.Flags().String("name", "", "Name of the schedule (required)")
			cmd.Flags().String("schedule", "", "Cron schedule expression in UTC, e.g. '0 2 * * *' (required)")
			cmd.Flags().String("collection", "", "Only back up this collection (default: the whole space)")
			cmd.Flags().Uint32("retention-days", 0, "Retention period of created backups in days (1-365) (default: keep indefinitely)")
			_ = cmd.MarkFlagRequired("space-id")
			_ = cmd.MarkFlagRequired("name")
			_ = cmd.MarkFlagRequired("schedule")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, _ []string) (*spacebackupv1.BackupSchedule, error) {
			spaceID, _ := cmd.Flags().GetString("space-id")
			name, _ := cmd.Flags().GetString("name")
			schedule, _ := cmd.Flags().GetString("schedule")

			sched := &spacebackupv1.BackupSchedule{
				SpaceId:  spaceID,
				Name:     name,
				Schedule: schedule,
			}

			if cmd.Flags().Changed("collection") {
				collection, _ := cmd.Flags().GetString("collection")
				sched.CollectionName = &collection
			}

			if cmd.Flags().Changed("retention-days") {
				days, _ := cmd.Flags().GetUint32("retention-days")
				retention, err := retentionFromDays(days)
				if err != nil {
					return nil, err
				}

				sched.RetentionPeriod = retention
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

			sched.AccountId = accountID

			resp, err := client.ServerlessBackup().CreateBackupSchedule(ctx, &spacebackupv1.CreateBackupScheduleRequest{
				BackupSchedule: sched,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create backup schedule: %w", err)
			}

			return resp.GetBackupSchedule(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, sched *spacebackupv1.BackupSchedule) {
			fmt.Fprintf(out, "Backup schedule %s (%s) created for space %s.\n", sched.GetId(), sched.GetName(), sched.GetSpaceId())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("space-id", completion.SpaceIDFlagCompletion(s))
	return cmd
}
