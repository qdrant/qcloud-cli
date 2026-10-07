package space

import (
	"github.com/spf13/cobra"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// backupScheduleIDCompletion returns a ValidArgsFunction that completes serverless
// backup schedule IDs, filtered by --space-id when it is set.
func backupScheduleIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx := cmd.Context()
		client, err := s.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		accountID, err := s.AccountID()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		req := &spacebackupv1.ListBackupSchedulesRequest{AccountId: accountID}
		if f := cmd.Flags().Lookup("space-id"); f != nil && f.Changed {
			spaceID := f.Value.String()
			req.SpaceId = &spaceID
		}

		resp, err := client.ServerlessBackup().ListBackupSchedules(ctx, req)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, sched := range resp.GetItems() {
			completions = append(completions, sched.GetId()+"\t"+sched.GetName()+" | "+sched.GetSchedule())
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
