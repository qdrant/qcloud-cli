package hybrid

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*hybridv1.HybridCloudEnvironment]{
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a new hybrid cloud environment",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("name", "", "Name of the hybrid cloud environment (required)")
			cmd.Flags().String("namespace", "", "Kubernetes namespace where Qdrant components are deployed")
			cmd.Flags().String("database-storage-class", "", "Default database storage class")
			cmd.Flags().String("snapshot-storage-class", "", "Default snapshot storage class")
			cmd.Flags().String("log-level", "", `Log level ("debug", "info", "warn", "error")`)
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
