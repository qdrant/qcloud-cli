package context

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
	"github.com/qdrant/qcloud-cli/internal/state/config"
)

type showOutput struct {
	Context       string `json:"context"`
	Endpoint      string `json:"endpoint"`
	AccountID     string `json:"account_id"`
	APIKeyCommand string `json:"api_key_command,omitempty"`
	Auth          string `json:"auth,omitempty"`
}

func newShowCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Example: `# Show the active context
qcloud context show`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "show",
				Short: "Show the active context configuration",
				Args:  cobra.NoArgs,
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			activeCtx := s.Config.ActiveContext()
			endpoint := s.Config.Endpoint()
			accountID := s.Config.AccountID()

			var apiKeyCommand, auth string
			if ctx, ok := s.Config.GetContext(activeCtx); ok {
				apiKeyCommand = ctx.APIKeyCommand
				auth = ctx.Auth
			}

			if s.Config.JSONOutput() {
				return output.PrintJSON(cmd.OutOrStdout(), showOutput{
					Context:       activeCtx,
					Endpoint:      endpoint,
					AccountID:     accountID,
					APIKeyCommand: apiKeyCommand,
					Auth:          auth,
				})
			}

			rows := [][]string{
				{"Context", activeCtx},
				{"Endpoint", endpoint},
				{"Account ID", accountID},
			}
			if apiKeyCommand != "" {
				rows = append(rows, []string{"API Key Command", apiKeyCommand})
			}

			if auth == config.AuthOAuth {
				rows = append(rows, []string{"Credential", "OAuth login (qcloud auth login)"})
			}

			t := output.NewTable[[]string](cmd.OutOrStdout())
			t.AddField("KEY", func(row []string) string { return row[0] })
			t.AddField("VALUE", func(row []string) string { return row[1] })
			t.Write(rows)
			return nil
		},
	}.CobraCommand(s)
}
