package account

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUpdateCommand(s *state.State) *cobra.Command {
	return base.UpdateCmd[*accountv1.Account]{
		Long: `Update an account's name or company information.

If no account ID is provided, the current account (from --account-id, the
active context, or the QDRANT_CLOUD_ACCOUNT_ID environment variable) is used.

Only flags that are explicitly set are applied. Unset flags leave the existing
values unchanged.`,
		Example: `# Rename the current account
qcloud account update --name "Production Account"

# Update company information on a specific account
qcloud account update a1b2c3d4-e5f6-7890-abcd-ef1234567890 --company-name "Acme Corp" --company-domain acme.com

# Output as JSON
qcloud account update --name "New Name" --json`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update [account-id]",
				Short: "Update an account",
				Args:  cobra.MaximumNArgs(1),
			}
			cmd.Flags().String("name", "", "New account name")
			cmd.Flags().String("company-name", "", "Company name")
			cmd.Flags().String("company-domain", "", "Company domain")
			return cmd
		},
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*accountv1.Account, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := resolveAccountID(s, args)
			if err != nil {
				return nil, err
			}

			resp, err := client.Account().GetAccount(ctx, &accountv1.GetAccountRequest{
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get account: %w", err)
			}

			return resp.GetAccount(), nil
		},
		Update: func(s *state.State, cmd *cobra.Command, acct *accountv1.Account) (*accountv1.Account, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			updated := proto.CloneOf(acct)

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				updated.Name = name
			}

			if cmd.Flags().Changed("company-name") || cmd.Flags().Changed("company-domain") {
				if updated.Company == nil {
					updated.Company = &accountv1.Company{}
				}
				if cmd.Flags().Changed("company-name") {
					name, _ := cmd.Flags().GetString("company-name")
					updated.Company.Name = name
				}
				if cmd.Flags().Changed("company-domain") {
					domain, _ := cmd.Flags().GetString("company-domain")
					updated.Company.Domain = &domain
				}
			}

			resp, err := client.Account().UpdateAccount(ctx, &accountv1.UpdateAccountRequest{
				Account: updated,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update account: %w", err)
			}

			return resp.GetAccount(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, acct *accountv1.Account) {
			if acct == nil {
				return
			}
			fmt.Fprintf(out, "Account %s (%s) updated successfully.\n", acct.GetId(), acct.GetName())
		},
	}.CobraCommand(s)
}
