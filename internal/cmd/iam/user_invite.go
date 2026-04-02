package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUserInviteCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "invite",
				Short: "Invite a user to the account",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("email", "", "Email address of the user to invite (required)")
			_ = cmd.MarkFlagRequired("email")
			cmd.Flags().StringArray("role", nil, "Role ID or name to assign to the invited user (repeatable)")
			return cmd
		},
		Long: `Invite a user to the account.

Sends an account invite to the specified email address. The invited user will
receive an invitation they can accept or reject.

Use --role to pre-assign roles to the invited user upon acceptance. Each
--role flag accepts either a role UUID or a role name.`,
		Example: `# Invite a user with no roles
qcloud iam user invite --email user@example.com

# Invite a user and assign a role by name
qcloud iam user invite --email user@example.com --role admin

# Invite a user and assign multiple roles
qcloud iam user invite --email user@example.com --role viewer --role admin`,
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			roleNames, _ := cmd.Flags().GetStringArray("role")

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}
			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			roleIDs, err := resolveRoleIDs(ctx, client, accountID, roleNames)
			if err != nil {
				return fmt.Errorf("--role: %w", err)
			}

			resp, err := client.Account().CreateAccountInvite(ctx, &accountv1.CreateAccountInviteRequest{
				AccountInvite: &accountv1.AccountInvite{
					AccountId:   accountID,
					UserEmail:   email,
					UserRoleIds: roleIDs,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create invite: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Invite %s sent to %s.\n",
				resp.GetAccountInvite().GetId(), email)
			return nil
		},
	}.CobraCommand(s)
}
