package cluster

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newDescribeCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*clusterv1.Cluster]{
		Use:   "describe <cluster-id>",
		Short: "Describe a cluster",
		Example: `# Describe a cluster
qcloud cluster describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output as JSON
qcloud cluster describe 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json`,
		Args: util.ExactArgs(1, "a cluster ID"),
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
			fmt.Fprintf(w, "ID:       %s\n", cluster.GetId())
			fmt.Fprintf(w, "Name:     %s\n", cluster.GetName())
			if cluster.GetState() != nil {
				fmt.Fprintf(w, "Status:   %s\n", phaseString(cluster.GetState().GetPhase()))
			}
			if cluster.GetConfiguration() != nil {
				cfg := cluster.GetConfiguration()
				fmt.Fprintf(w, "Version:  %s\n", cfg.GetVersion())
				fmt.Fprintf(w, "Nodes:    %d\n", cfg.GetNumberOfNodes())
				fmt.Fprintf(w, "Package:  %s\n", cfg.GetPackageId())
			}
			fmt.Fprintf(w, "Cloud:    %s\n", cluster.GetCloudProviderId())
			fmt.Fprintf(w, "Region:   %s\n", cluster.GetCloudProviderRegionId())
			if cluster.GetCreatedAt() != nil {
				t := cluster.GetCreatedAt().AsTime()
				fmt.Fprintf(w, "Created:  %s  (%s)\n", output.HumanTime(t), output.FullDateTime(t))
			}
			if labels := cluster.GetLabels(); len(labels) > 0 {
				fmt.Fprintf(w, "Labels:   ")
				for i, kv := range labels {
					if i > 0 {
						fmt.Fprintf(w, "          ")
					}
					fmt.Fprintf(w, "%s=%s\n", kv.GetKey(), kv.GetValue())
				}
			}

			if cfg := cluster.GetConfiguration(); cfg != nil {
				notSet := "(not set)"

				if dbCfg := cfg.GetDatabaseConfiguration(); dbCfg != nil {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Database Configuration:")

					col := dbCfg.GetCollection()
					fmt.Fprintln(w, "  Collection Defaults:")
					if col != nil {
						fmt.Fprintf(w, "    Replication Factor:        %s\n", output.OptionalValue(col.ReplicationFactor, notSet))
						fmt.Fprintf(w, "    Write Consistency Factor:  %s\n", output.OptionalValue(col.WriteConsistencyFactor, notSet))
						if vec := col.GetVectors(); vec != nil {
							fmt.Fprintf(w, "    On Disk:                   %s\n", output.OptionalValue(vec.OnDisk, notSet))
						} else {
							fmt.Fprintf(w, "    On Disk:                   %s\n", notSet)
						}
					} else {
						fmt.Fprintf(w, "    Replication Factor:        %s\n", notSet)
						fmt.Fprintf(w, "    Write Consistency Factor:  %s\n", notSet)
						fmt.Fprintf(w, "    On Disk:                   %s\n", notSet)
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

				ips := cfg.GetAllowedIpSourceRanges()
				restartMode := restartPolicyString(cfg.GetRestartPolicy())
				rebalance := rebalanceStrategyString(cfg.GetRebalanceStrategy())

				if len(ips) > 0 || restartMode != "" || rebalance != "" {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Cluster Configuration:")
					if len(ips) > 0 {
						fmt.Fprintf(w, "  Allowed IPs:          %s\n", strings.Join(ips, ", "))
					}
					if restartMode != "" {
						fmt.Fprintf(w, "  Restart Mode:         %s\n", restartMode)
					}
					if rebalance != "" {
						fmt.Fprintf(w, "  Rebalance Strategy:   %s\n", rebalance)
					}
				}
			}

			if st := cluster.GetState(); st != nil {
				if ep := st.GetEndpoint(); ep != nil {
					fmt.Fprintln(w)
					fmt.Fprintf(w, "Endpoint:   %s\n", ep.GetUrl())
					fmt.Fprintf(w, "REST Port:  %d\n", ep.GetRestPort())
					fmt.Fprintf(w, "gRPC Port:  %d\n", ep.GetGrpcPort())
				}

				if res := st.GetResources(); res != nil {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Resources (per node):")
					if disk := res.GetDisk(); disk != nil {
						tier := storageTierString(cluster.GetConfiguration().GetClusterStorageConfiguration().GetStorageTierType())
						tierSuffix := "\n"
						if tier != "" {
							tierSuffix = fmt.Sprintf(" (tier: %s)\n", tier)
						}
						fmt.Fprintf(w, "  Disk:  %s base, %s available"+tierSuffix,
							formatGiB(disk.GetBase()), formatGiB(disk.GetAvailable()))
					}
					if ram := res.GetRam(); ram != nil {
						fmt.Fprintf(w, "  RAM:   %s base, %s reserved, %s available\n",
							formatGiB(ram.GetBase()), formatGiB(ram.GetReserved()), formatGiB(ram.GetAvailable()))
					}
					if cpu := res.GetCpu(); cpu != nil {
						fmt.Fprintf(w, "  CPU:   %s base, %s reserved, %s available\n",
							formatMillicores(cpu.GetBase()), formatMillicores(cpu.GetReserved()), formatMillicores(cpu.GetAvailable()))
					}
					if gpu := res.GetGpu(); gpu != nil {
						fmt.Fprintf(w, "  GPU:   %s base, %s reserved, %s available\n",
							formatMillicores(gpu.GetBase()), formatMillicores(gpu.GetReserved()), formatMillicores(gpu.GetAvailable()))
					}
				}

				if nodes := st.GetNodes(); len(nodes) > 0 {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Nodes:")
					for _, n := range nodes {
						started := ""
						if n.GetStartedAt() != nil {
							started = "started " + output.HumanTime(n.GetStartedAt().AsTime())
						}
						fmt.Fprintf(w, "  %-40s  %-12s  %-10s  %s\n",
							n.GetName(), nodeStateString(n.GetState()), n.GetVersion(), started)
					}
				}
			}

			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
