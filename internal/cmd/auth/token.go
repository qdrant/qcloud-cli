package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newTokenCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Print a fresh OAuth access token for the current endpoint.

The token is sensitive. Use this for local Terraform (auth = "cli") or scripts
that need a Bearer token. Expired access tokens are refreshed when a refresh
token is stored.`,
		Example: `qcloud auth token
qcloud auth token --json`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "token",
				Short: "Print a fresh OAuth access token (sensitive)",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			path := oauth.CredentialsPath(s.Config.ConfigFilePath())
			tok, ok, err := oauth.LoadToken(path, s.Config.Endpoint())
			if err != nil {
				return err
			}

			if !ok {
				return fmt.Errorf("not logged in — run \"qcloud auth login\"")
			}

			if tok.Expired(time.Now()) {
				refreshed, err := s.OAuthDiscoverer().Refresh(cmd.Context(), tok, time.Now())
				if err != nil {
					return fmt.Errorf("refresh OAuth token: %w", err)
				}

				if err := oauth.SaveToken(path, refreshed); err != nil {
					return err
				}

				tok = refreshed
			}

			if s.Config.JSONOutput() {
				fmt.Fprintf(cmd.OutOrStdout(), `{"access_token":%q,"token_type":%q,"scope":%q}`+"\n", tok.AccessToken, tok.TokenType, tok.Scope)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), tok.AccessToken)
			return nil
		},
	}.CobraCommand(s)
}
