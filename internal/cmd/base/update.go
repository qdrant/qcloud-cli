package base

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// UpdateCmd defines a command for updating a resource.
// BaseCobraCommand builds the cobra.Command with Use, Short, Args, and flag definitions.
// Fetch retrieves the existing resource; Update applies changes and returns the updated one.
// Read flags in Update via cmd.Flags().GetString() etc. — do not use bound vars.
type UpdateCmd[T any] struct {
	BaseCobraCommand  func() *cobra.Command
	Example           string
	Fetch             func(s *state.State, cmd *cobra.Command, args []string) (T, error)
	Update            func(s *state.State, cmd *cobra.Command, resource T) (T, error)
	PrintResource     func(cmd *cobra.Command, out io.Writer, resource T)
	ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// CobraCommand builds a cobra.Command from this UpdateCmd.
func (uc UpdateCmd[T]) CobraCommand(s *state.State) *cobra.Command {
	cmd := uc.BaseCobraCommand()
	cmd.Example = uc.Example
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		resource, err := uc.Fetch(s, cmd, args)
		if err != nil {
			return err
		}
		updated, err := uc.Update(s, cmd, resource)
		if err != nil {
			return err
		}
		if s.Config.JSONOutput() {
			return output.PrintJSON(cmd.OutOrStdout(), updated)
		}
		if uc.PrintResource != nil {
			uc.PrintResource(cmd, cmd.OutOrStdout(), updated)
		}
		return nil
	}
	if uc.ValidArgsFunction != nil {
		cmd.ValidArgsFunction = uc.ValidArgsFunction
	}
	return cmd
}
