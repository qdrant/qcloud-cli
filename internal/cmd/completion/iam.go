package completion

import (
	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// RoleCompletion returns a completion function that completes IAM role names
// with their ID as description.
func RoleCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.IAM().ListRoles(ctx, &iamv1.ListRolesRequest{AccountId: accountID})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, r := range resp.GetItems() {
			completions = append(completions, r.GetName()+"\t"+r.GetId())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// RoleIDCompletion returns a ValidArgsFunction that completes role IDs.
func RoleIDCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.IAM().ListRoles(ctx, &iamv1.ListRolesRequest{
			AccountId: accountID,
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

// PermissionCompletion returns a completion function for the --permission flag.
func PermissionCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.IAM().ListPermissions(ctx, &iamv1.ListPermissionsRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetPermissions()))
		for _, p := range resp.GetPermissions() {
			completions = append(completions, p.GetValue()+"\t"+p.GetCategory())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
