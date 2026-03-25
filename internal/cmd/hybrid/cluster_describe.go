package hybrid

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newClusterDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*clusterv1.Cluster]{
		Use:               "describe <cluster-id>",
		Short:             "Describe a cluster in a hybrid cloud environment",
		Args:              util.ExactArgs(1, "a cluster ID"),
		ValidArgsFunction: hybridClusterIDCompletion(s),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			resp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
				AccountId: accountID,
				ClusterId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster: %w", err)
			}

			return resp.GetCluster(), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, cluster *clusterv1.Cluster) error {
			clusterPhase := strings.TrimPrefix(cluster.GetState().GetPhase().String(), "CLUSTER_PHASE_")

			fmt.Fprintf(w, "ID:          %s\n", cluster.GetId())
			fmt.Fprintf(w, "Name:        %s\n", cluster.GetName())
			fmt.Fprintf(w, "Status:      %s\n", clusterPhase)
			fmt.Fprintf(w, "Environment: %s\n", cluster.GetCloudProviderRegionId())

			if cfg := cluster.GetConfiguration(); cfg != nil {
				fmt.Fprintf(w, "Version:     %s\n", cfg.GetVersion())
				fmt.Fprintf(w, "Nodes:       %d\n", cfg.GetNumberOfNodes())
			}

			if cluster.GetCreatedAt() != nil {
				t := cluster.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:     %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}

			if labels := cluster.GetLabels(); len(labels) > 0 {
				fmt.Fprintf(w, "Labels:      ")
				for i, kv := range labels {
					if i > 0 {
						fmt.Fprintf(w, "             ")
					}
					fmt.Fprintf(w, "%s=%s\n", kv.GetKey(), kv.GetValue())
				}
			}

			if cfg := cluster.GetConfiguration(); cfg != nil {
				notSet := "(not set)"

				// Hybrid-specific configuration
				hasHybridCfg := cfg.ServiceType != nil || len(cfg.GetNodeSelector()) > 0 ||
					len(cfg.GetTolerations()) > 0 || len(cfg.GetAnnotations()) > 0 ||
					len(cfg.GetPodLabels()) > 0 || len(cfg.GetServiceAnnotations()) > 0 ||
					cfg.ReservedCpuPercentage != nil || cfg.ReservedMemoryPercentage != nil

				if hasHybridCfg {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Hybrid Configuration:")
					if cfg.ServiceType != nil {
						st := strings.TrimPrefix(cfg.GetServiceType().String(), "CLUSTER_SERVICE_TYPE_")
						fmt.Fprintf(w, "  Service Type:              %s\n", st)
					}
					if cfg.ReservedCpuPercentage != nil {
						fmt.Fprintf(w, "  Reserved CPU %%:            %d\n", cfg.GetReservedCpuPercentage())
					}
					if cfg.ReservedMemoryPercentage != nil {
						fmt.Fprintf(w, "  Reserved Memory %%:         %d\n", cfg.GetReservedMemoryPercentage())
					}
					if ns := cfg.GetNodeSelector(); len(ns) > 0 {
						fmt.Fprintf(w, "  Node Selectors:\n")
						for _, kv := range ns {
							fmt.Fprintf(w, "    %s=%s\n", kv.GetKey(), kv.GetValue())
						}
					}
					if tols := cfg.GetTolerations(); len(tols) > 0 {
						fmt.Fprintf(w, "  Tolerations:\n")
						for _, t := range tols {
							op := strings.TrimPrefix(t.GetOperator().String(), "TOLERATION_OPERATOR_")
							effect := strings.TrimPrefix(t.GetEffect().String(), "TOLERATION_EFFECT_")
							if t.GetOperator() == clusterv1.TolerationOperator_TOLERATION_OPERATOR_EXISTS {
								fmt.Fprintf(w, "    key=%s operator=%s effect=%s\n", t.GetKey(), op, effect)
							} else {
								fmt.Fprintf(w, "    key=%s value=%s operator=%s effect=%s\n", t.GetKey(), t.GetValue(), op, effect)
							}
						}
					}
					if anns := cfg.GetAnnotations(); len(anns) > 0 {
						fmt.Fprintf(w, "  Annotations:\n")
						for _, kv := range anns {
							fmt.Fprintf(w, "    %s=%s\n", kv.GetKey(), kv.GetValue())
						}
					}
					if pl := cfg.GetPodLabels(); len(pl) > 0 {
						fmt.Fprintf(w, "  Pod Labels:\n")
						for _, kv := range pl {
							fmt.Fprintf(w, "    %s=%s\n", kv.GetKey(), kv.GetValue())
						}
					}
					if sa := cfg.GetServiceAnnotations(); len(sa) > 0 {
						fmt.Fprintf(w, "  Service Annotations:\n")
						for _, kv := range sa {
							fmt.Fprintf(w, "    %s=%s\n", kv.GetKey(), kv.GetValue())
						}
					}
				}

				if dbCfg := cfg.GetDatabaseConfiguration(); dbCfg != nil {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Database Configuration:")
					col := dbCfg.GetCollection()
					fmt.Fprintln(w, "  Collection Defaults:")
					if col != nil {
						fmt.Fprintf(w, "    Replication Factor:        %s\n", output.OptionalValue(col.ReplicationFactor, notSet))
						fmt.Fprintf(w, "    Write Consistency Factor:  %s\n", output.OptionalValue(col.WriteConsistencyFactor, notSet))
					} else {
						fmt.Fprintf(w, "    Replication Factor:        %s\n", notSet)
						fmt.Fprintf(w, "    Write Consistency Factor:  %s\n", notSet)
					}
					perf := dbCfg.GetStorage().GetPerformance()
					fmt.Fprintln(w, "  Advanced Optimizations:")
					if perf != nil {
						fmt.Fprintf(w, "    Optimizer CPU Budget:      %s\n", output.OptionalValue(perf.OptimizerCpuBudget, notSet))
						fmt.Fprintf(w, "    Async Scorer:              %s\n", output.OptionalValue(perf.AsyncScorer, notSet))
					} else {
						fmt.Fprintf(w, "    Optimizer CPU Budget:      %s\n", notSet)
						fmt.Fprintf(w, "    Async Scorer:              %s\n", notSet)
					}
				}

				if ips := cfg.GetAllowedIpSourceRanges(); len(ips) > 0 {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Cluster Configuration:")
					fmt.Fprintf(w, "  Allowed IPs:  %s\n", strings.Join(ips, ", "))
				}
			}

			if st := cluster.GetState(); st != nil {
				if ep := st.GetEndpoint(); ep != nil {
					fmt.Fprintln(w)
					fmt.Fprintf(w, "Endpoint:   %s\n", ep.GetUrl())
					fmt.Fprintf(w, "REST Port:  %d\n", ep.GetRestPort())
					fmt.Fprintf(w, "gRPC Port:  %d\n", ep.GetGrpcPort())
				}
			}

			return nil
		},
	}.CobraCommand(s)
}
