package iam

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	authv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/auth/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newKeyCreateCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*authv1.ManagementKey]{
		Long: `Create a new cloud management key for the account.

Management keys grant access to the Qdrant Cloud API. The full key value is returned
only once at creation time — store it securely, as it cannot be retrieved again. If a
key is lost, delete it and create a new one.`,
		Example: `# Create a new management key
qcloud iam key create

# Create and capture the key value in a script
qcloud iam key create --json | jq -r '.key'`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "create",
				Short: "Create a cloud management key",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*authv1.ManagementKey, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Auth().CreateManagementKey(ctx, &authv1.CreateManagementKeyRequest{
				ManagementKey: &authv1.ManagementKey{
					AccountId: accountID,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create management key: %w", err)
			}

			return resp.GetManagementKey(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, key *authv1.ManagementKey) {
			fmt.Fprintf(out, "Management key %s created.\n", key.GetId())
			if k := key.GetKey(); k != "" {
				fmt.Fprintln(out, "")
				fmt.Fprintln(out, "Save this key now — it will not be shown again:")
				fmt.Fprintf(out, "  %s\n", k)
			}
		},
	}.CobraCommand(s)
}
