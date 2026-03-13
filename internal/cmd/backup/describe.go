package backup

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*backupv1.Backup]{
		Use:               "describe <backup-id>",
		Short:             "Describe a backup",
		Args:              util.ExactArgs(1, "a backup ID"),
		ValidArgsFunction: backupIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*backupv1.Backup, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Backup().GetBackup(ctx, &backupv1.GetBackupRequest{
				AccountId: accountID,
				BackupId:  args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get backup: %w", err)
			}
			return resp.GetBackup(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, b *backupv1.Backup) error {
			fmt.Fprintf(w, "ID:        %s\n", b.GetId())
			fmt.Fprintf(w, "Name:      %s\n", b.GetName())
			fmt.Fprintf(w, "Cluster:   %s\n", b.GetClusterId())
			fmt.Fprintf(w, "Status:    %s\n", backupStatusString(b.GetStatus()))
			if b.GetCreatedAt() != nil {
				t := b.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:   %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			if b.GetBackupScheduleId() != "" {
				fmt.Fprintf(w, "Schedule:  %s\n", b.GetBackupScheduleId())
			}
			if b.GetRetentionPeriod() != nil {
				days := int64(b.GetRetentionPeriod().AsDuration().Hours()) / 24
				fmt.Fprintf(w, "Retention: %d days\n", days)
			}
			if b.GetClusterInfo() != nil {
				ci := b.GetClusterInfo()
				fmt.Fprintf(w, "Cloud:     %s\n", ci.GetCloudProviderId())
				fmt.Fprintf(w, "Region:    %s\n", ci.GetCloudProviderRegionId())
			}
			return nil
		},
	}.CobraCommand(s)
}
