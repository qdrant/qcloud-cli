package cluster

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDashboardCommand(s *state.State) *cobra.Command {
	return base.Cmd{
		Example: `# Open a cluster's dashboard in your default browser
qcloud cluster dashboard 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Print the dashboard URL instead of opening a browser (headless/SSH)
qcloud cluster dashboard 7b2ea926-724b-4de2-b73a-8675c42a6ebe --print-url`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "dashboard <cluster-id>",
				Short: "Open a cluster's dashboard in your browser",
				Long: `Open a cluster's dashboard in your default browser.

The command builds the Cloud UI dashboard URL and opens it. The Cloud UI page
handles authentication using your existing browser session and redirects to the
cluster's dashboard.`,
				Args: util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().Bool("print-url", false, "Print the dashboard URL instead of opening a browser")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) error {
			clusterID := args[0]

			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return err
			}

			// Confirm the cluster exists so we give a clear error instead of
			// opening a browser to a URL that will not resolve.
			if _, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
				AccountId: accountID,
				ClusterId: clusterID,
			}); err != nil {
				return fmt.Errorf("failed to find cluster %s: %w", clusterID, err)
			}

			dashURL, err := dashboardURL(s.Config.ConsoleURL(), accountID, clusterID)
			if err != nil {
				return err
			}

			printURL, _ := cmd.Flags().GetBool("print-url")
			if printURL {
				fmt.Fprintln(cmd.OutOrStdout(), dashURL)
				return nil
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "Opening dashboard for cluster %s in your browser...\n", clusterID)
			if err := s.OpenBrowser(dashURL); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Could not open a browser. Open this URL manually:\n%s\n", dashURL)
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}

// dashboardURL builds the Cloud UI dashboard URL for a cluster from the web
// console base URL, account ID, and cluster ID.
func dashboardURL(consoleBase, accountID, clusterID string) (string, error) {
	base := strings.TrimRight(consoleBase, "/")
	if _, err := url.Parse(base); err != nil {
		return "", fmt.Errorf("invalid console URL %q: %w", consoleBase, err)
	}

	return url.JoinPath(base, "accounts", accountID, "clusters", clusterID, "dashboard")
}
