package base

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// CreateCmd defines a command for creating a resource.
// BaseCobraCommand builds the cobra.Command with Use, Short, Args, and flag definitions.
// Read flags in Run via cmd.Flags().GetString() etc. — do not use bound vars.
type CreateCmd[T any] struct {
	BaseCobraCommand  func() *cobra.Command
	Long              string
	Example           string
	Run               func(s *state.State, cmd *cobra.Command, args []string) (T, error)
	PrintResource     func(cmd *cobra.Command, out io.Writer, resource T)
	ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// CobraCommand builds a cobra.Command from this CreateCmd.
func (cc CreateCmd[T]) CobraCommand(s *state.State) *cobra.Command {
	cmd := cc.BaseCobraCommand()
	if cc.Long != "" {
		cmd.Long = cc.Long
	}
	cmd.Example = cc.Example
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		resource, err := cc.Run(s, cmd, args)
		if err != nil {
			return err
		}
		if s.Config.JSONOutput() {
			return output.PrintJSON(cmd.OutOrStdout(), resource)
		}
		if cc.PrintResource != nil {
			cc.PrintResource(cmd, cmd.OutOrStdout(), resource)
		}
		return nil
	}
	if cc.ValidArgsFunction != nil {
		cmd.ValidArgsFunction = cc.ValidArgsFunction
	}
	return cmd
}
