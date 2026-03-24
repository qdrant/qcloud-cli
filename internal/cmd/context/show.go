package context

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

type showOutput struct {
	Context   string `json:"context"`
	Endpoint  string `json:"endpoint"`
	AccountID string `json:"account_id"`
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

			if s.Config.JSONOutput() {
				return output.PrintJSON(cmd.OutOrStdout(), showOutput{Context: activeCtx, Endpoint: endpoint, AccountID: accountID})
			}

			t := output.NewTable[[]string](cmd.OutOrStdout())
			t.AddField("KEY", func(row []string) string { return row[0] })
			t.AddField("VALUE", func(row []string) string { return row[1] })
			t.Write([][]string{
				{"Context", activeCtx},
				{"Endpoint", endpoint},
				{"Account ID", accountID},
			})
			return nil
		},
	}.CobraCommand(s)
}
