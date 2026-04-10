package completion

import (
	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"
	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	platformv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/platform/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// CloudProviderCompletion returns a completion function for the --cloud-provider flag. It skips the 'hybrid' cloud,
// as this flag is meant to be used for cloud clusters, not hybrid ones.
func CloudProviderCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		ctx := cmd.Context()
		client, err := s.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		accountID, err := s.AccountID()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		resp, err := client.Platform().ListCloudProviders(ctx, &platformv1.ListCloudProvidersRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, p := range resp.GetItems() {
			completions = append(completions, p.GetId()+"\t"+p.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CloudRegionCompletion returns a completion function for the --cloud-region flag.
func CloudRegionCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		provider, _ := cmd.Flags().GetString("cloud-provider")
		if provider == "" {
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

		resp, err := client.Platform().ListCloudProviderRegions(ctx, &platformv1.ListCloudProviderRegionsRequest{
			AccountId:       accountID,
			CloudProviderId: provider,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, r := range resp.GetItems() {
			completions = append(completions, r.GetId()+"\t"+r.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// ManagementKeyIDCompletion returns a ValidArgsFunction that completes management key IDs.
func ManagementKeyIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.Auth().ListManagementKeys(ctx, &authv1.ListManagementKeysRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, k := range resp.GetItems() {
			completions = append(completions, k.GetId()+"\t"+k.GetPrefix())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// BackupIDCompletion returns a ValidArgsFunction that completes backup IDs.
func BackupIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

// ClusterIDCompletion returns a ValidArgsFunction that completes cluster IDs.
func ClusterIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		clusters, err := client.Cluster().ListAllClusters(ctx, accountID)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(clusters))
		for _, c := range clusters {
			completions = append(completions, c.GetId()+"\t"+c.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// VersionCompletion returns a completion function for the --version flag.
func VersionCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		ctx := cmd.Context()
		client, err := s.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		accountID, err := s.AccountID()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		resp, err := client.Cluster().ListQdrantReleases(ctx, &clusterv1.ListQdrantReleasesRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, r := range resp.GetItems() {
			if r.GetUnavailable() {
				continue
			}
			desc := ""
			if r.GetDefault() {
				desc += "(default)"
			}
			if r.GetEndOfLife() {
				if desc != "" {
					desc += " "
				}
				desc += "(end of life)"
			}
			if remarks := r.GetRemarks(); remarks != "" {
				if desc != "" {
					desc += " "
				}
				desc += remarks
			}
			entry := r.GetVersion()
			if desc != "" {
				entry += "\t" + desc
			}
			completions = append(completions, entry)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// AccountIDCompletion returns a ValidArgsFunction that completes account IDs.
func AccountIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx := cmd.Context()
		client, err := s.Client(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		resp, err := client.Account().ListAccounts(ctx, &accountv1.ListAccountsRequest{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, a := range resp.GetItems() {
			completions = append(completions, a.GetId()+"\t"+a.GetName())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// AccountMemberIDCompletion returns a ValidArgsFunction that completes account member user IDs.
func AccountMemberIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.Account().ListAccountMembers(ctx, &accountv1.ListAccountMembersRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, m := range resp.GetItems() {
			completions = append(completions, m.GetAccountMember().GetId()+"\t"+m.GetAccountMember().GetEmail())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
