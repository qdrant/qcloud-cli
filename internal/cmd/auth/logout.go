package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newLogoutCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Long: `Remove the stored OAuth token for the current endpoint.

If the authorization server advertises a revocation endpoint, the refresh
token is revoked first. Management API keys in the config file are not
changed.`,
		Example: `qcloud auth logout`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "logout",
				Short: "Log out and drop stored OAuth tokens",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			path := oauth.CredentialsPath(s.Config.ConfigFilePath())
			endpoint := s.Config.Endpoint()
			tok, ok, err := oauth.LoadToken(path, endpoint)
			if err != nil {
				return err
			}

			if ok {
				if err := s.OAuthDiscoverer().RevokeRefreshToken(cmd.Context(), tok); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not revoke refresh token: %v\n", err)
				}
			}

			if err := oauth.DeleteToken(path, endpoint); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
			return nil
		},
	}.CobraCommand(s)
}
