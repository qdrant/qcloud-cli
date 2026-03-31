package hybrid

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUpdateCommand(s *state.State) *cobra.Command {
	return base.UpdateCmd[*hybridv1.HybridCloudEnvironment]{
		ValidArgsFunction: envIDCompletion(s),
		Example: `# Rename a hybrid cloud environment
qcloud hybrid update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --name new-name

# Update the default storage classes
qcloud hybrid update 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --database-storage-class premium-rwo --snapshot-storage-class standard

# Change the log level
qcloud hybrid update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --log-level debug`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <env-id>",
				Short: "Update a hybrid cloud environment",
				Args:  util.ExactArgs(1, "a hybrid cloud environment ID"),
			}
			cmd.Flags().String("name", "", "New name for the hybrid cloud environment")
			cmd.Flags().String("namespace", "", "Kubernetes namespace where Qdrant components are deployed (read-only after bootstrapping)")
			cmd.Flags().String("database-storage-class", "", "Default database storage class (uses cluster default if omitted)")
			cmd.Flags().String("snapshot-storage-class", "", "Default snapshot storage class (uses cluster default if omitted)")
			cmd.Flags().String("log-level", "", `Log level for deployed components ("debug", "info", "warn", "error")`)
			cmd.RegisterFlagCompletionFunc("log-level", logLevelCompletion())
			return cmd
		},
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*hybridv1.HybridCloudEnvironment, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Hybrid().GetHybridCloudEnvironment(ctx, &hybridv1.GetHybridCloudEnvironmentRequest{
				AccountId:                accountID,
				HybridCloudEnvironmentId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get hybrid cloud environment: %w", err)
			}

			return resp.GetHybridCloudEnvironment(), nil
		},
		Update: func(s *state.State, cmd *cobra.Command, env *hybridv1.HybridCloudEnvironment) (*hybridv1.HybridCloudEnvironment, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			updated := proto.CloneOf(env)

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				updated.Name = name
			}

			if cmd.Flags().Changed("namespace") || cmd.Flags().Changed("database-storage-class") ||
				cmd.Flags().Changed("snapshot-storage-class") || cmd.Flags().Changed("log-level") {
				if updated.Configuration == nil {
					updated.Configuration = &hybridv1.HybridCloudEnvironmentConfiguration{}
				}
				cfg := updated.Configuration

				if cmd.Flags().Changed("namespace") {
					ns, _ := cmd.Flags().GetString("namespace")
					cfg.Namespace = ns
				}
				if cmd.Flags().Changed("database-storage-class") {
					v, _ := cmd.Flags().GetString("database-storage-class")
					cfg.DatabaseStorageClass = &v
				}
				if cmd.Flags().Changed("snapshot-storage-class") {
					v, _ := cmd.Flags().GetString("snapshot-storage-class")
					cfg.SnapshotStorageClass = &v
				}
				if cmd.Flags().Changed("log-level") {
					lvlStr, _ := cmd.Flags().GetString("log-level")
					lvl, err := parseLogLevel(lvlStr)
					if err != nil {
						return nil, err
					}
					cfg.LogLevel = &lvl
				}
			}

			resp, err := client.Hybrid().UpdateHybridCloudEnvironment(ctx, &hybridv1.UpdateHybridCloudEnvironmentRequest{
				HybridCloudEnvironment: updated,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update hybrid cloud environment: %w", err)
			}

			return resp.GetHybridCloudEnvironment(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, env *hybridv1.HybridCloudEnvironment) {
			fmt.Fprintf(out, "Hybrid cloud environment %s (%s) updated.\n", env.GetId(), env.GetName())
		},
	}.CobraCommand(s)
}
