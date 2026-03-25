package hybrid

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	hybridv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/hybrid/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*hybridv1.HybridCloudEnvironment]{
		Use:   "describe <env-id>",
		Short: "Describe a hybrid cloud environment",
		Args:  util.ExactArgs(1, "a hybrid cloud environment ID"),
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
		PrintText: func(_ *cobra.Command, w io.Writer, env *hybridv1.HybridCloudEnvironment) error {
			fmt.Fprintf(w, "ID:                  %s\n", env.GetId())
			fmt.Fprintf(w, "Name:                %s\n", env.GetName())
			if env.GetStatus() != nil {
				fmt.Fprintf(w, "Status:              %s\n", phaseString(env.GetStatus().GetPhase()))
			}
			if env.GetCreatedAt() != nil {
				t := env.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:             %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			fmt.Fprintf(w, "Bootstrap Generated: %s\n", boolToYesNo(env.GetBootstrapCommandsGenerated()))

			if cfg := env.GetConfiguration(); cfg != nil {
				fmt.Fprintf(w, "\nConfiguration:\n")
				fmt.Fprintf(w, "  Namespace:              %s\n", cfg.GetNamespace())
				if cfg.DatabaseStorageClass != nil {
					fmt.Fprintf(w, "  Database Storage Class: %s\n", cfg.GetDatabaseStorageClass())
				}
				if cfg.SnapshotStorageClass != nil {
					fmt.Fprintf(w, "  Snapshot Storage Class: %s\n", cfg.GetSnapshotStorageClass())
				}
				if cfg.LogLevel != nil {
					fmt.Fprintf(w, "  Log Level:              %s\n", logLevelString(cfg.GetLogLevel()))
				}
				if cfg.HttpProxyUrl != nil {
					fmt.Fprintf(w, "  HTTP Proxy:             %s\n", cfg.GetHttpProxyUrl())
				}
				if cfg.HttpsProxyUrl != nil {
					fmt.Fprintf(w, "  HTTPS Proxy:            %s\n", cfg.GetHttpsProxyUrl())
				}
			}

			if st := env.GetStatus(); st != nil {
				fmt.Fprintf(w, "\nStatus:\n")
				if st.KubernetesVersion != "" {
					fmt.Fprintf(w, "  Kubernetes Version:   %s\n", st.GetKubernetesVersion())
				}
				if st.KubernetesDistribution != nil {
					dist := st.GetKubernetesDistribution().String()
					fmt.Fprintf(w, "  Distribution:         %s\n", dist)
				}
				fmt.Fprintf(w, "  Node Count:           %d\n", st.GetNumberOfNodes())
				fmt.Fprintf(w, "  Cluster Creation:     %s\n", clusterCreationStatusString(st.GetClusterCreationReadiness()))
				if len(st.GetComponentStatuses()) > 0 {
					fmt.Fprintf(w, "  Components:\n")
					for _, cs := range st.GetComponentStatuses() {
						fmt.Fprintf(w, "    %-30s %s\n", cs.GetName(), componentPhaseString(cs.GetPhase()))
					}
				}
			}

			return nil
		},
	}.CobraCommand(s)
}

func boolToYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
