package iam

import (
	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"
	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
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

// userThenRoleCompletion returns a ValidArgsFunction that completes user
// IDs/emails for the first positional argument, and role names/IDs for all
// subsequent arguments.
func userThenRoleCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return listUserCompletions(s, cmd)
		}
		return completion.RoleCompletion(s)(cmd, args, "")
	}
}

// inviteCompletion returns a ValidArgsFunction that completes invite IDs with
// the invited email as description. It only completes the first positional
// argument.
func inviteCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
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

		resp, err := client.Account().ListAccountInvites(ctx, &accountv1.ListAccountInvitesRequest{
			AccountId: accountID,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(resp.GetItems()))
		for _, inv := range resp.GetItems() {
			completions = append(completions, inv.GetId()+"\t"+inv.GetUserEmail())
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
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
