package hybrid

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*hybridv1.HybridCloudEnvironment]{
		Example: `# Create a hybrid cloud environment
qcloud hybrid create --name my-hybrid-env

# Create with a custom namespace
qcloud hybrid create --name my-hybrid-env --namespace qdrant-hybrid

# Create with storage classes
qcloud hybrid create --name my-hybrid-env \
  --database-storage-class premium-rwo --snapshot-storage-class standard`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a new hybrid cloud environment",
				Long: `Create a new Hybrid Cloud Environment to deploy and manage Qdrant on your own
Kubernetes clusters (on-premises, cloud, or edge) with enterprise-grade
reliability.

Hybrid Cloud access must be enabled for your account by the Qdrant sales team.
If your account does not have access, you will be prompted to contact us.`,
				Args: cobra.NoArgs,
			}
			cmd.Flags().String("name", "", "Name of the hybrid cloud environment (required)")
			cmd.Flags().String("namespace", "", "Kubernetes namespace where Qdrant components are deployed (read-only after bootstrapping)")
			cmd.Flags().String("database-storage-class", "", "Default database storage class (uses cluster default if omitted)")
			cmd.Flags().String("snapshot-storage-class", "", "Default snapshot storage class (uses cluster default if omitted)")
			cmd.Flags().String("log-level", "", `Log level for deployed components ("debug", "info", "warn", "error")`)
			_ = cmd.MarkFlagRequired("name")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*hybridv1.HybridCloudEnvironment, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			name, _ := cmd.Flags().GetString("name")

			env := &hybridv1.HybridCloudEnvironment{
				AccountId: accountID,
				Name:      name,
			}

			if cmd.Flags().Changed("namespace") || cmd.Flags().Changed("database-storage-class") ||
				cmd.Flags().Changed("snapshot-storage-class") || cmd.Flags().Changed("log-level") {
				env.Configuration = &hybridv1.HybridCloudEnvironmentConfiguration{}

				if cmd.Flags().Changed("namespace") {
					ns, _ := cmd.Flags().GetString("namespace")
					env.Configuration.Namespace = ns
				}
				if cmd.Flags().Changed("database-storage-class") {
					v, _ := cmd.Flags().GetString("database-storage-class")
					env.Configuration.DatabaseStorageClass = &v
				}
				if cmd.Flags().Changed("snapshot-storage-class") {
					v, _ := cmd.Flags().GetString("snapshot-storage-class")
					env.Configuration.SnapshotStorageClass = &v
				}
				if cmd.Flags().Changed("log-level") {
					lvlStr, _ := cmd.Flags().GetString("log-level")
					lvl, err := parseLogLevel(lvlStr)
					if err != nil {
						return nil, err
					}
					env.Configuration.LogLevel = &lvl
				}
			}

			resp, err := client.Hybrid().CreateHybridCloudEnvironment(ctx, &hybridv1.CreateHybridCloudEnvironmentRequest{
				HybridCloudEnvironment: env,
			})
			if err != nil {
				if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
					return nil, fmt.Errorf("your account does not have access to Hybrid Cloud\n\nTo get started, contact us at: https://qdrant.tech/contact-us/")
				}
				return nil, fmt.Errorf("failed to create hybrid cloud environment: %w", err)
			}

			return resp.GetHybridCloudEnvironment(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, env *hybridv1.HybridCloudEnvironment) {
			fmt.Fprintf(out, "Hybrid cloud environment %s (%s) created.\n", env.GetId(), env.GetName())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("log-level", logLevelCompletion())
	return cmd
}
