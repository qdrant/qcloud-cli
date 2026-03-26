package hybrid

import (
	"fmt"

	"github.com/spf13/cobra"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newBootstrapCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		ValidArgsFunction: envIDCompletion(s),
		Example: `# Generate bootstrap commands for an environment
qcloud hybrid bootstrap 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Pipe directly to a shell
qcloud hybrid bootstrap 7b2ea926-724b-4de2-b73a-8675c42a6ebe | bash`,
		BaseCobraCommand: func() *cobra.Command {
			return &cobra.Command{
				Use:   "bootstrap <env-id>",
				Short: "Generate bootstrap commands for a hybrid cloud environment",
				Long: `Generate the commands needed to bootstrap a Kubernetes cluster into a hybrid cloud environment.

Each command in the output is ready to copy-paste or pipe to a shell. The credentials
printed to stderr are sensitive and should be treated as secrets.

Note: each invocation creates new Qdrant Cloud access tokens and registry credentials.
Only run this if the Kubernetes cluster is not yet registered to the environment.`,
				Args: util.ExactArgs(1, "a hybrid cloud environment ID"),
			}
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			resp, err := client.Hybrid().GenerateBootstrapCommands(ctx, &hybridv1.GenerateBootstrapCommandsRequest{
				AccountId:                accountID,
				HybridCloudEnvironmentId: args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to generate bootstrap commands: %w", err)
			}

			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			for _, command := range resp.GetCommands() {
				fmt.Fprintln(out, command)
			}

			fmt.Fprintln(errOut, "")
			fmt.Fprintln(errOut, "WARNING: The following credentials are sensitive. Handle them as secrets.")
			fmt.Fprintf(errOut, "Access Key:        %s\n", resp.GetAccessKey())
			fmt.Fprintf(errOut, "Registry Username: %s\n", resp.GetRegistryUsername())
			fmt.Fprintf(errOut, "Registry Password: %s\n", resp.GetRegistryPassword())

			return nil
		},
	}.CobraCommand(s)
}
