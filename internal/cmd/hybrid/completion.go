package hybrid

import (
	"github.com/spf13/cobra"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// envIDCompletion returns a ValidArgsFunction that completes hybrid cloud environment IDs.
func envIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.Hybrid().ListHybridCloudEnvironments(ctx, &hybridv1.ListHybridCloudEnvironmentsRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, env := range resp.GetItems() {
			completions = append(completions, env.GetId()+"\t"+env.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// logLevelCompletion returns a static completion function for the --log-level flag.
func logLevelCompletion() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{logLevelDebug, logLevelInfo, logLevelWarn, logLevelError}, cobra.ShellCompDirectiveNoFileComp
	}
}
