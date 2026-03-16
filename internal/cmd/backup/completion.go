package backup

import (
	"github.com/spf13/cobra"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// backupIDCompletion returns a ValidArgsFunction that completes backup IDs.
func backupIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		req := &backupv1.ListBackupsRequest{AccountId: accountID}
		if cmd.Flags().Changed("cluster-id") {
			clusterID, _ := cmd.Flags().GetString("cluster-id")
			req.ClusterId = &clusterID
		}

		resp, err := client.Backup().ListBackups(ctx, req)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, b := range resp.GetItems() {
			completions = append(completions, b.GetId()+"\t"+b.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
