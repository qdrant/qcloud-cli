package auth

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/oauth"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// tokenOutput is a fresh access token. The field names are what the Terraform provider reads.
type tokenOutput struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	// Scope is the requested scope.
	Scope        string    `json:"scope,omitempty"`
	GrantedScope string    `json:"granted_scope,omitempty"`
	Expiry       time.Time `json:"expiry,omitzero"`
}

func newTokenCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*tokenOutput]{
		Use:   "token",
		Short: "Print a fresh OAuth access token (sensitive)",
		Long: `Print a fresh OAuth access token for the current endpoint.

An access token that has expired, or is about to, is refreshed first. The token
is a credential: it is meant for tools that need a Bearer token, such as the
Terraform provider with auth = "cli", and should not be logged or shared.`,
		Example: `# Print the access token
qcloud auth token

# Keep the token in a variable for a script
TOKEN=$(qcloud auth token)

# Print the token with its type, scope, and expiry
qcloud auth token --json`,
		Args: cobra.NoArgs,
		Fetch: func(s *state.State, cmd *cobra.Command, _ []string) (*tokenOutput, error) {
			tok, err := s.OAuth().Token(cmd.Context(), s.Config.Endpoint())
			if errors.Is(err, oauth.ErrNotLoggedIn) {
				return nil, fmt.Errorf("not logged in — run \"qcloud auth login\"")
			}

			if err != nil {
				return nil, err
			}

			return &tokenOutput{
				AccessToken:  tok.AccessToken,
				TokenType:    tok.TokenType,
				Scope:        tok.Scope,
				GrantedScope: tok.GrantedScope,
				Expiry:       tok.Expiry,
			}, nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, tok *tokenOutput) error {
			_, err := fmt.Fprintln(w, tok.AccessToken)
			return err
		},
	}.CobraCommand(s)
}
