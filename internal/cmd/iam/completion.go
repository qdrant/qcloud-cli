package iam

import (
	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// userCompletion returns a ValidArgsFunction that completes user IDs with
// their email as description. It only completes the first positional argument.
func userCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return listUserCompletions(s, cmd)
	}
}

func listUserCompletions(s *state.State, cmd *cobra.Command) ([]string, cobra.ShellCompDirective) {
	ctx := cmd.Context()
	client, err := s.Client(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	accountID, err := s.AccountID()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := client.IAM().ListUsers(ctx, &iamv1.ListUsersRequest{AccountId: accountID})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := make([]string, 0, len(resp.GetItems()))
	for _, u := range resp.GetItems() {
		completions = append(completions, u.GetId()+"\t"+u.GetEmail())
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
