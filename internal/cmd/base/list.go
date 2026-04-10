package base

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

// ListCmd defines a command for fetching and displaying a list response.
// T is the full response proto message (e.g. *clusterv1.ListClustersResponse).
//
// OutputTable must be set. When set, --no-headers is automatically registered
// and handled.
type ListCmd[T any] struct {
	Use               string
	Short             string
	Long              string
	Example           string
	Args              cobra.PositionalArgs // optional; defaults to cobra.NoArgs
	Fetch             func(s *state.State, cmd *cobra.Command) (T, error)
	OutputTable       func(cmd *cobra.Command, out io.Writer, resp T) (output.TableRenderer, error)
	ValidArgsFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// CobraCommand builds a cobra.Command from this ListCmd.
func (lc ListCmd[T]) CobraCommand(s *state.State) *cobra.Command {
	posArgs := lc.Args
	if posArgs == nil {
		posArgs = cobra.NoArgs
	}
	cmd := &cobra.Command{
		Use:     lc.Use,
		Short:   lc.Short,
		Long:    lc.Long,
		Example: lc.Example,
		Args:    posArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := lc.Fetch(s, cmd)
			if err != nil {
				return err
			}
			if s.Config.JSONOutput() {
				return output.PrintJSON(cmd.OutOrStdout(), resp)
			}
			r, err := lc.OutputTable(cmd, cmd.OutOrStdout(), resp)
			if err != nil {
				return err
			}
			noHeaders, _ := cmd.Flags().GetBool("no-headers")
			r.SetNoHeaders(noHeaders)
			r.Render()
			return nil
		},
	}
	cmd.Flags().Bool("no-headers", false, "Do not print column headers")
	if lc.ValidArgsFunction != nil {
		cmd.ValidArgsFunction = lc.ValidArgsFunction
	}
	return cmd
}
