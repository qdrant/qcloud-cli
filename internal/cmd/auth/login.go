package auth

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	accountv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/account/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

// loginOutput is the result of a login: who logged in, with what access, and the context of each
// granted account.
type loginOutput struct {
	Endpoint     string `json:"endpoint"`
	Email        string `json:"email,omitempty"`
	GrantedScope string `json:"granted_scope,omitempty"`
	// Contexts is empty when no account was granted, or the accounts could not be listed.
	Contexts []loginContext `json:"contexts"`
	// CurrentContext is the current context after the login.
	CurrentContext string `json:"current_context,omitempty"`
	// SelectedContext is the context the login made current, if it did.
	SelectedContext string `json:"selected_context,omitempty"`
	// AccountsListed is false when the accounts could not be listed, so no contexts were written.
	AccountsListed bool `json:"accounts_listed"`
}

type loginContext struct {
	Name        string `json:"name"`
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
}

func newLoginCommand(s *state.State) *cobra.Command {
	return base.CreateCmd[*loginOutput]{
		Long: `Log in to Qdrant Cloud and store a session that refreshes itself, for the current endpoint.

By default the login opens a browser and waits for it on 127.0.0.1; use
--device on a machine without a browser (over SSH, for example). Either way the
login ends on the Qdrant Cloud consent page, where you choose the accounts and
the access (read only or manage) the CLI gets.

After the login, qcloud creates a context for every account you granted, named
after the account and prefixed with the environment (dev-my-account on
development), that uses this login instead of an API key. Unless a context is
already current, the first one becomes current. A later login updates these
contexts, and "qcloud auth logout" removes them. Contexts with an API key are
never changed.

The access and refresh tokens are kept in the OS keychain (macOS Keychain,
Windows Credential Manager, or the Secret Service on Linux), and the rest of the
session in credentials.yaml next to the config file. Without a keychain, the
tokens go to credentials.yaml (mode 0600) too; QDRANT_CLOUD_TOKEN_STORE=keychain
makes that an error, and QDRANT_CLOUD_TOKEN_STORE=file skips the keychain.`,
		Example: `# Log in with the browser
qcloud auth login

# Log in with read-only access
qcloud auth login --scope read-only

# Log in on a machine without a browser
qcloud auth login --device

# Log in from a script and use the context it selected
qcloud auth login --json | jq -r .current_context`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "login",
				Short: "Log in with a browser or device code",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("scope", "manage", "Access to ask for: read-only or manage (the consent page can narrow it)")
			cmd.Flags().Bool("device", false, "Log in with a device code (prints a code to confirm in any browser) instead of opening a browser here")
			cmd.Flags().String("client-id", "", "OAuth client id, for a login server without a built-in one (env: QDRANT_CLOUD_OAUTH_CLIENT_ID)")
			cmd.Flags().String("login-url", "", "Login server URL, instead of the one discovered from the endpoint (for debugging)")
			s.Config.BindPFlag(config.KeyOAuthClientID, cmd.Flags().Lookup("client-id"))
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, _ []string) (*loginOutput, error) {
			scope, _ := cmd.Flags().GetString("scope")
			device, _ := cmd.Flags().GetBool("device")
			loginURL, _ := cmd.Flags().GetString("login-url")

			res, err := s.OAuth().Login(cmd.Context(), oauth.LoginOptions{
				Endpoint: s.Config.Endpoint(),
				ClientID: s.Config.OAuthClientID(),
				Scope:    scope,
				Device:   device,
				LoginURL: loginURL,
				OpenURL:  s.OpenBrowser,
				Output:   cmd.ErrOrStderr(),
			})
			if err != nil {
				return nil, err
			}

			if res.KeychainErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: no OS keychain available (%v); the tokens are stored in %s (0600).\n", res.KeychainErr, res.CredentialsPath)
			}

			out := &loginOutput{
				Endpoint:     s.Config.Endpoint(),
				Email:        res.Token.Email,
				GrantedScope: res.Token.GrantedScope,
				Contexts:     []loginContext{},
			}

			accounts, ok := grantedAccounts(cmd, s)
			if !ok {
				out.CurrentContext = s.Config.CurrentContext()
				return out, nil
			}

			names, selected, err := syncOAuthContexts(s, out.Endpoint, accounts)
			if err != nil {
				return nil, fmt.Errorf("logged in, but could not write the contexts: %w", err)
			}

			out.AccountsListed = true
			out.SelectedContext = selected
			out.CurrentContext = s.Config.CurrentContext()
			for i, name := range names {
				out.Contexts = append(out.Contexts, loginContext{Name: name, AccountID: accounts[i].GetId(), AccountName: accounts[i].GetName()})
			}

			return out, nil
		},
		PrintResource: printLoginResult,
	}.CobraCommand(s)
}

// accountLookupTimeout bounds the account lookup after login; the login itself has already succeeded.
const accountLookupTimeout = 15 * time.Second

// grantedAccounts lists the accounts the login was granted. It never fails the login: on an error it
// warns and returns false. The lookup always uses the new login, whatever API key the current context
// has; an API key from --api-key or QDRANT_CLOUD_API_KEY is only reported, as it overrides the login.
func grantedAccounts(cmd *cobra.Command, s *state.State) ([]*accountv1.Account, bool) {
	if s.Config.APIKeyFromFlagOrEnv() {
		fmt.Fprintln(cmd.ErrOrStderr(), "Warning: an API key is set with --api-key or QDRANT_CLOUD_API_KEY, so commands use it instead of this login.")
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), accountLookupTimeout)
	defer cancel()

	client, err := s.OAuthClient(ctx)
	if err == nil {
		var resp *accountv1.ListAccountsResponse
		if resp, err = client.Account().ListAccounts(ctx, &accountv1.ListAccountsRequest{}); err == nil {
			return resp.GetItems(), true
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not list your accounts (%v); no contexts were created, pass --account-id to commands.\n", err)

	return nil, false
}

// printLoginResult reports the login and the contexts it created, one line per context.
func printLoginResult(_ *cobra.Command, out io.Writer, res *loginOutput) {
	who := "Logged in"
	if res.Email != "" {
		who += " as " + res.Email
	}

	access := ""
	if res.GrantedScope != "" {
		access = ", " + oauth.ScopeLabel(res.GrantedScope)
	}

	switch {
	case !res.AccountsListed:
		fmt.Fprintf(out, "%s%s.\n", who, access)
	case len(res.Contexts) == 0:
		fmt.Fprintf(out, "%s%s; no accounts were granted.\n", who, access)
	case len(res.Contexts) == 1:
		c := res.Contexts[0]
		fmt.Fprintf(out, "%s, account %s%s (context %s).\n", who, c.AccountName, access, c.Name)
		if res.CurrentContext != c.Name {
			fmt.Fprintf(out, "Switch to it with: qcloud context use %s\n", c.Name)
		}
	default:
		fmt.Fprintf(out, "%s%s. Contexts for your %d accounts:\n", who, access, len(res.Contexts))
		for _, c := range res.Contexts {
			marker := " "
			if c.Name == res.CurrentContext {
				marker = "*"
			}

			fmt.Fprintf(out, "  %s %s  (%s)\n", marker, c.Name, c.AccountName)
		}

		fmt.Fprintln(out, "Switch with: qcloud context use <name>")
	}
}
