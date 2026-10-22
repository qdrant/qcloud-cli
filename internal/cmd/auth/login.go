package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newLoginCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Log in and store a refreshable access token for the current endpoint.

The default flow is authorization code + PKCE with a loopback listener on
127.0.0.1 (RFC 8252), on the first free port of 48271, 48272 and 48273. The
browser callback checks state and, when Auth0 sends it, iss.

Use --device on SSH or when no browser is available. The CLI only prints the
verification URL and polls. Sign-in and the Qdrant consent page happen in the
browser; a declined consent ends the poll with access denied.

The login URL comes from the gateway's unauthenticated protected-resource
metadata when OAuth is enabled, otherwise from rewriting grpc.cloud.qdrant.io
to login.cloud.qdrant.io. Override with --login-url only for debugging.

The Qdrant Cloud CLI client id is built in per login URL. --client-id or
QDRANT_CLOUD_OAUTH_CLIENT_ID override it, for a tenant without a built-in id.
Tokens are stored in credentials.yaml next to the config file (mode 0600).

When the login is granted exactly one account, that account is stored with the
session and used by commands when no --account-id, QDRANT_CLOUD_ACCOUNT_ID or
context sets one ("qcloud auth status" shows it).`,
		Example: `# Browser login (default production endpoint)
qcloud auth login

# Read-only session
qcloud auth login --scope read-only

# Device code
qcloud auth login --device`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "login",
				Short: "Log in with a browser or device code",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("scope", "manage", "OAuth scope: read-only or manage")
			cmd.Flags().Bool("device", false, "Use the device authorization grant instead of a loopback browser flow")
			cmd.Flags().String("client-id", os.Getenv("QDRANT_CLOUD_OAUTH_CLIENT_ID"), "Override the built-in OAuth client id (env: QDRANT_CLOUD_OAUTH_CLIENT_ID)")
			cmd.Flags().String("login-url", "", "Override the login issuer URL (defaults to discovery / endpoint inference)")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			scope, _ := cmd.Flags().GetString("scope")
			device, _ := cmd.Flags().GetBool("device")
			clientID, _ := cmd.Flags().GetString("client-id")
			loginURL, _ := cmd.Flags().GetString("login-url")

			opts := oauth.LoginOptions{
				Endpoint:    s.Config.Endpoint(),
				ClientID:    clientID,
				Scope:       scope,
				Device:      device,
				LoginURL:    loginURL,
				WaitCode:    s.WaitAuthCode(),
				RedirectURL: s.OAuthRedirectURL(),
				OpenURL:     s.OpenBrowser,
				Output:      cmd.ErrOrStderr(),
			}

			tok, _, err := s.OAuthDiscoverer().Login(cmd.Context(), opts)
			if err != nil {
				return err
			}

			path := oauth.CredentialsPath(s.Config.ConfigFilePath())
			if err := oauth.SaveToken(path, tok); err != nil {
				return err
			}

			account := rememberGrantedAccount(cmd, s, path, tok)

			who := "Logged in"
			if email := oauth.TokenEmail(tok.AccessToken); email != "" {
				who += " as " + email
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s%s.\n", who, account)

			return nil
		},
	}.CobraCommand(s)
}

// accountLookupTimeout bounds the account lookup after login; the login itself has already succeeded.
const accountLookupTimeout = 15 * time.Second

// rememberGrantedAccount stores the account the login was granted when it was granted exactly one, so
// commands need no --account-id, and returns the text the login line ends with. It never fails the
// login: without it, commands ask for an account as before. An API key takes precedence over the OAuth
// session, so then there is nothing to look up.
func rememberGrantedAccount(cmd *cobra.Command, s *state.State, path string, tok oauth.Token) string {
	if s.Config.APIKey() != "" {
		return ""
	}

	stderr := cmd.ErrOrStderr()
	ctx, cancel := context.WithTimeout(cmd.Context(), accountLookupTimeout)
	defer cancel()

	client, err := s.Client(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "Could not list your accounts (%v); pass --account-id to commands.\n", err)
		return ""
	}

	resp, err := client.Account().ListAccounts(ctx, &accountv1.ListAccountsRequest{})
	if err != nil {
		fmt.Fprintf(stderr, "Could not list your accounts (%v); pass --account-id to commands.\n", err)
		return ""
	}

	accounts := resp.GetItems()
	if len(accounts) != 1 {
		return fmt.Sprintf("; %d accounts available, pass --account-id", len(accounts))
	}

	tok.AccountID = accounts[0].GetId()
	if err := oauth.SaveToken(path, tok); err != nil {
		fmt.Fprintf(stderr, "Could not store the account: %v\n", err)
		return ""
	}

	return ", account " + accounts[0].GetName()
}
