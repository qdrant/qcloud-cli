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
The access and refresh tokens are stored in the OS keychain (macOS Keychain,
Windows Credential Manager, or the Secret Service on Linux); the rest of the
session is in credentials.yaml next to the config file. Without a keychain, the
tokens go to credentials.yaml (mode 0600) too. Set QDRANT_CLOUD_TOKEN_STORE to
"keychain" to require the keychain, or to "file" to skip it.

After the login, qcloud creates a context for every account you granted on the
consent page (named after the account, prefixed with the environment, e.g.
dev-my-account), marked as OAuth so it needs no API key. Unless you already have
a current context, the first one becomes current; switch with
"qcloud context use <name>". A later login updates
these contexts, and "qcloud auth logout" removes them. Contexts with an API key
are never changed.`,
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

			claims := oauth.ReadTokenClaims(tok.AccessToken)
			tok.GrantedScope = claims.GrantedScope

			path := oauth.CredentialsPath(s.Config.ConfigFilePath())
			saved, err := oauth.SaveToken(path, tok)
			if err != nil {
				return err
			}

			if saved.KeychainErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "No OS keychain available (%v); the tokens are stored in %s (0600).\n", saved.KeychainErr, path)
			}

			who := "Logged in"
			if claims.Email != "" {
				who += " as " + claims.Email
			}

			access := ""
			if tok.GrantedScope != "" {
				access = ", " + oauth.ScopeLabel(tok.GrantedScope)
			}

			accounts, ok := grantedAccounts(cmd, s)
			if !ok {
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s.\n", who, access)
				return nil
			}

			endpoint := s.Config.Endpoint()
			names, selected, err := syncOAuthContexts(s, endpoint, accounts)
			if err != nil {
				return fmt.Errorf("logged in, but could not write the contexts: %w", err)
			}

			printLoginResult(cmd, who, access, accounts, names, selected, s.Config.CurrentContext())

			return nil
		},
	}.CobraCommand(s)
}

// accountLookupTimeout bounds the account lookup after login; the login itself has already succeeded.
const accountLookupTimeout = 15 * time.Second

// grantedAccounts lists the accounts the login was granted. It never fails the login: on an error it
// warns and returns false. An API key takes precedence over the OAuth login, so then there is nothing to
// look up.
func grantedAccounts(cmd *cobra.Command, s *state.State) ([]*accountv1.Account, bool) {
	if s.Config.APIKey() != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "An API key is set, so commands use it instead of this login; no contexts were created.")
		return nil, false
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), accountLookupTimeout)
	defer cancel()

	client, err := s.Client(ctx)
	if err == nil {
		var resp *accountv1.ListAccountsResponse
		if resp, err = client.Account().ListAccounts(ctx, &accountv1.ListAccountsRequest{}); err == nil {
			return resp.GetItems(), true
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Could not list your accounts (%v); no contexts were created, pass --account-id to commands.\n", err)

	return nil, false
}

// printLoginResult reports the login and the contexts it created, one line per context.
func printLoginResult(cmd *cobra.Command, who, access string, accounts []*accountv1.Account, names []string, selected, current string) {
	out := cmd.OutOrStdout()
	switch len(names) {
	case 0:
		fmt.Fprintf(out, "%s%s; no accounts were granted.\n", who, access)
	case 1:
		fmt.Fprintf(out, "%s, account %s%s (context %s).\n", who, accounts[0].GetName(), access, names[0])
		if selected == "" {
			fmt.Fprintf(out, "Switch to it with: qcloud context use %s\n", names[0])
		}
	default:
		fmt.Fprintf(out, "%s%s. Contexts for your %d accounts:\n", who, access, len(names))
		for i, name := range names {
			marker := " "
			if name == current {
				marker = "*"
			}

			fmt.Fprintf(out, "  %s %s  (%s)\n", marker, name, accounts[i].GetName())
		}

		fmt.Fprintln(out, "Switch with: qcloud context use <name>")
	}
}
