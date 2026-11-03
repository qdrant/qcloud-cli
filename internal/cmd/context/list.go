package context

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

type listOutput struct {
	Contexts []string `json:"contexts"`
	Current  string   `json:"current"`
	// Auth is the credential of each context: "oauth" or "api key".
	Auth map[string]string `json:"auth"`
}

func newListCommand(s *state.State) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all contexts",
		Example: `# List all contexts
qcloud context list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			names := s.Config.ContextNames()
			current := s.Config.CurrentContext()

			credential := func(name string) string {
				if ctx, ok := s.Config.GetContext(name); ok && ctx.IsOAuth() {
					return "oauth"
				}

				return "api key"
			}

			if s.Config.JSONOutput() {
				auth := make(map[string]string, len(names))
				for _, name := range names {
					auth[name] = credential(name)
				}

				return output.PrintJSON(cmd.OutOrStdout(), listOutput{Contexts: names, Current: current, Auth: auth})
			}

			t := output.NewTable[string](cmd.OutOrStdout())
			t.AddField("CURRENT", func(name string) string {
				if name == current {
					return "*"
				}

				return ""
			})
			t.AddField("NAME", func(name string) string {
				return name
			})
			t.AddField("CREDENTIAL", credential)
			t.Write(names)
			return nil
		},
	}
}
