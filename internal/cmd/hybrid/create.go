package hybrid

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "create",
				Short: "Create a new hybrid cloud environment",
				Long: `Hybrid Cloud Environments let you deploy and manage Qdrant on your own
Kubernetes clusters (on-premises, cloud, or edge) with enterprise-grade
reliability.

Creating a hybrid cloud environment requires contacting the Qdrant sales team.`,
				Args: cobra.NoArgs,
			}
		},
		Run: func(_ *state.State, cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Hybrid Cloud Environments let you deploy and manage Qdrant on your own")
			fmt.Fprintln(cmd.OutOrStdout(), "Kubernetes clusters (on-premises, cloud, or edge) with enterprise-grade")
			fmt.Fprintln(cmd.OutOrStdout(), "reliability.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "To get started, contact us at: https://qdrant.tech/contact-us/")
			return nil
		},
	}.CobraCommand(s)
}
