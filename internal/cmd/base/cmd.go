package base

import (
	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// Cmd is a generic wrapper for imperative (action) commands that don't
// return a resource. Use for delete, wait, use, set, and similar operations.
type Cmd struct {
	BaseCobraCommand  func() *cobra.Command
	Run               func(s *state.State, cmd *cobra.Command, args []string) error
	ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// CobraCommand builds a cobra.Command from this Cmd.
func (gc Cmd) CobraCommand(s *state.State) *cobra.Command {
	cmd := gc.BaseCobraCommand()
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return gc.Run(s, cmd, args)
	}
	if gc.ValidArgsFunction != nil {
		cmd.ValidArgsFunction = gc.ValidArgsFunction
	}
	return cmd
}
