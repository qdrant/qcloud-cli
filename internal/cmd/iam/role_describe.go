package iam

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	iamv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/iam/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newRoleDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*iamv1.Role]{
		Use:   "describe <role-id>",
		Short: "Describe a role",
		Long: `Display detailed information about a role, including its name, type,
description, and the full list of assigned permissions.`,
		Example: `# Describe a role
qcloud iam role describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output as JSON
qcloud iam role describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json`,
		Args:              util.ExactArgs(1, "a role ID"),
		ValidArgsFunction: completion.RoleIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*iamv1.Role, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.IAM().GetRole(ctx, &iamv1.GetRoleRequest{
				AccountId: accountID,
				RoleId:    args[0],
			})
			if err != nil {
				return nil, err
			}

			return resp.GetRole(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, role *iamv1.Role) error {
			fmt.Fprintf(w, "ID:            %s\n", role.GetId())
			fmt.Fprintf(w, "Name:          %s\n", role.GetName())
			fmt.Fprintf(w, "Description:   %s\n", role.GetDescription())
			fmt.Fprintf(w, "Type:          %s\n", output.RoleType(role.GetRoleType()))
			if role.GetCreatedAt() != nil {
				fmt.Fprintf(w, "Created:       %s\n", output.FullDateTime(role.GetCreatedAt().AsTime()))
			}
			if role.GetLastModifiedAt() != nil {
				fmt.Fprintf(w, "Last Modified: %s\n", output.FullDateTime(role.GetLastModifiedAt().AsTime()))
			}
			fmt.Fprintf(w, "\nPermissions:\n")
			for _, p := range role.GetPermissions() {
				if cat := p.GetCategory(); cat != "" {
					fmt.Fprintf(w, "  %-30s (%s)\n", p.GetValue(), cat)
				} else {
					fmt.Fprintf(w, "  %s\n", p.GetValue())
				}
			}
			return nil
		},
	}.CobraCommand(s)
}
