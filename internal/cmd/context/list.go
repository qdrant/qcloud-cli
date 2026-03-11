package context

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

type listOutput struct {
	Contexts []string `json:"contexts"`
	Current  string   `json:"current"`
}

func newListCommand(s *state.State) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all contexts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			names := s.Config.ContextNames()
			current := s.Config.CurrentContext()

			if s.Config.JSONOutput() {
				return output.PrintJSON(cmd.OutOrStdout(), listOutput{Contexts: names, Current: current})
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
			t.Write(names)
			return nil
		},
	}
}
