package base

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// DescribeCmd defines a command for fetching and displaying a single resource.
type DescribeCmd[T any] struct {
	Use               string
	Short             string
	Args              cobra.PositionalArgs
	Fetch             func(s *state.State, cmd *cobra.Command, args []string) (T, error)
	PrintText         func(cmd *cobra.Command, out io.Writer, resource T) error
	ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// CobraCommand builds a cobra.Command from this DescribeCmd.
func (dc DescribeCmd[T]) CobraCommand(s *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   dc.Use,
		Short: dc.Short,
		Args:  dc.Args,
		RunE: func(cmd *cobra.Command, args []string) error {
			resource, err := dc.Fetch(s, cmd, args)
			if err != nil {
				return err
			}
			if s.Config.JSONOutput() {
				return output.PrintJSON(cmd.OutOrStdout(), resource)
			}
			return dc.PrintText(cmd, cmd.OutOrStdout(), resource)
		},
	}
	if dc.ValidArgsFunction != nil {
		cmd.ValidArgsFunction = dc.ValidArgsFunction
	}
	return cmd
}
