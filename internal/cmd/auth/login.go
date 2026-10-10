package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newLoginCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Log in and store a refreshable access token for the current endpoint.

The default flow is authorization code + PKCE with a loopback listener on
127.0.0.1 (RFC 8252). Use --device on SSH or when no browser is available.

The login URL comes from the gateway's unauthenticated protected-resource
metadata when OAuth is enabled, otherwise from rewriting grpc.cloud.qdrant.io
to login.cloud.qdrant.io. Override with --login-url only for debugging.

The native application client id is not hard-coded yet.
Pass --client-id or QDRANT_CLOUD_OAUTH_CLIENT_ID after that app exists.
Tokens are stored in credentials.yaml next to the config file (mode 0600).`,
		Example: `# Browser login (default production endpoint)
qcloud auth login --client-id <id>

# Read-only session
qcloud auth login --scope read-only --client-id <id>

# Device code
qcloud auth login --device --client-id <id>`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "login",
				Short: "Log in with a browser or device code",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("scope", "manage", "OAuth scope: read-only or manage")
			cmd.Flags().Bool("device", false, "Use the device authorization grant instead of a loopback browser flow")
			cmd.Flags().String("client-id", os.Getenv("QDRANT_CLOUD_OAUTH_CLIENT_ID"), "OAuth native app client id (env: QDRANT_CLOUD_OAUTH_CLIENT_ID)")
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

			tok, disc, err := s.OAuthDiscoverer().Login(cmd.Context(), opts)
			if err != nil {
				return err
			}

			path := oauth.CredentialsPath(s.Config.ConfigFilePath())
			if err := oauth.SaveToken(path, tok); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Logged in.\nLogin URL: %s (%s)\nResource:  %s\nScope:     %s\n", disc.Issuer, disc.Source, disc.Resource, tok.Scope)
			return nil
		},
	}.CobraCommand(s)
}
