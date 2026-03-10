package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/state"
)

// NewCommand creates the "version" command.
func NewCommand(s *state.State) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the qcloud CLI version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "qcloud %s\n", s.Version)
		},
	}
}
