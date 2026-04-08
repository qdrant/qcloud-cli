package account

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*accountv1.Account]{
		Use:   "describe [account-id]",
		Short: "Describe an account",
		Long: `Describe an account by its ID.

If no account ID is provided, the current account (from --account-id, the
active context, or the QDRANT_CLOUD_ACCOUNT_ID environment variable) is used.`,
		Example: `# Describe the current account
qcloud account describe

# Describe a specific account
qcloud account describe a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Output as JSON
qcloud account describe --json`,
		Args: cobra.MaximumNArgs(1),
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
		PrintText: func(_ *cobra.Command, w io.Writer, acct *accountv1.Account) error {
			fmt.Fprintf(w, "ID:          %s\n", acct.GetId())
			fmt.Fprintf(w, "Name:        %s\n", acct.GetName())
			fmt.Fprintf(w, "Owner Email: %s\n", acct.GetOwnerEmail())
			if company := acct.GetCompany(); company != nil {
				fmt.Fprintf(w, "Company:     %s\n", company.GetName())
				if company.Domain != nil {
					fmt.Fprintf(w, "Domain:      %s\n", company.GetDomain())
				}
			}
			if privs := acct.GetPrivileges(); len(privs) > 0 {
				fmt.Fprintf(w, "Privileges:  %s\n", strings.Join(privs, ", "))
			}
			if acct.GetCreatedAt() != nil {
				t := acct.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:     %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			if acct.GetLastModifiedAt() != nil {
				t := acct.GetLastModifiedAt().AsTime()
				fmt.Fprintf(w, "Modified:    %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			return nil
		},
		ValidArgsFunction: completion.AccountIDCompletion(s),
	}.CobraCommand(s)
}

// resolveAccountID returns args[0] if present, otherwise falls back to s.AccountID().
func resolveAccountID(s *state.State, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	return s.AccountID()
}
