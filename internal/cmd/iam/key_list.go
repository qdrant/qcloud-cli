package iam

import (
	"io"

	"github.com/spf13/cobra"

	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyListCommand(s *state.State) *cobra.Command {
	return base.ListCmd[*authv1.ListManagementKeysResponse]{
		Use:   "list",
		Short: "List cloud management keys",
		Long: `List all cloud management keys for the account.

Management keys grant access to the Qdrant Cloud API and are used to authenticate CLI
and API requests. Each key is identified by its ID and a prefix — the prefix represents
the first bytes of the key value and is safe to display.`,
		Example: `# List all management keys for the account
qcloud iam key list

# Output as JSON
qcloud iam key list --json`,
		Fetch: func(s *state.State, cmd *cobra.Command) (*authv1.ListManagementKeysResponse, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			return client.Auth().ListManagementKeys(ctx, &authv1.ListManagementKeysRequest{
				AccountId: accountID,
			})
		},
		OutputTable: func(_ *cobra.Command, w io.Writer, resp *authv1.ListManagementKeysResponse) (output.TableRenderer, error) {
			t := output.NewTable[*authv1.ManagementKey](w)
			t.AddField("ID", func(v *authv1.ManagementKey) string {
				return v.GetId()
			})
			t.AddField("PREFIX", func(v *authv1.ManagementKey) string {
				return v.GetPrefix()
			})
			t.AddField("CREATED", func(v *authv1.ManagementKey) string {
				if v.GetCreatedAt() != nil {
					return output.HumanTime(v.GetCreatedAt().AsTime())
				}
				return ""
			})
			t.SetItems(resp.GetItems())
			return t, nil
		},
	}.CobraCommand(s)
}
