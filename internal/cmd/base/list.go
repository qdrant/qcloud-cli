package base

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// ListCmd defines a command for fetching and displaying a list response.
// T is the full response proto message (e.g. *clusterv1.ListClustersResponse).
type ListCmd[T any] struct {
	Use               string
	Short             string
	Example           string
	Fetch             func(s *state.State, cmd *cobra.Command) (T, error)
	PrintText         func(cmd *cobra.Command, out io.Writer, resp T) error
	ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// CobraCommand builds a cobra.Command from this ListCmd.
func (lc ListCmd[T]) CobraCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:     lc.Use,
		Short:   lc.Short,
		Example: lc.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := lc.Fetch(s, cmd)
			if err != nil {
				return err
			}
			if s.Config.JSONOutput() {
				return output.PrintJSON(cmd.OutOrStdout(), resp)
			}
			return lc.PrintText(cmd, cmd.OutOrStdout(), resp)
		},
	}
	if lc.ValidArgsFunction != nil {
		cmd.ValidArgsFunction = lc.ValidArgsFunction
	}
	return cmd
}
