package account

import (
	"io"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*accountv1.ListAccountsResponse]{
		Use:   "list",
		Short: "List accounts",
		Long: `List all accounts associated with the authenticated management key.

Returns every account the current API key has access to. No account ID is
required because the server resolves accounts from the caller's credentials.`,
		Example: `# List all accessible accounts
qcloud account list

# Output as JSON
qcloud account list --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*accountv1.ListAccountsResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			return client.Account().ListAccounts(ctx, &accountv1.ListAccountsRequest{})
		},
		PrintText: func(_ *cobra.Command, w io.Writer, resp *accountv1.ListAccountsResponse) error {
			t := output.NewTable[*accountv1.Account](w)
			t.AddField("ID", func(v *accountv1.Account) string { return v.GetId() })
			t.AddField("NAME", func(v *accountv1.Account) string { return v.GetName() })
			t.AddField("OWNER EMAIL", func(v *accountv1.Account) string { return v.GetOwnerEmail() })
			t.AddField("CREATED", func(v *accountv1.Account) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.Write(resp.GetItems())
			return nil
		},
	}.CobraCommand(s)
}
