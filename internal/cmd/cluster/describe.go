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
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
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
			isHybrid := cluster.GetCloudProviderId() == qcloudapi.HybridCloudProviderID

			fmt.Fprintf(w, "ID:       %s\n", cluster.GetId())
			fmt.Fprintf(w, "Name:     %s\n", cluster.GetName())
			if cluster.GetState() != nil {
				fmt.Fprintf(w, "Status:   %s\n", output.ClusterPhase(cluster.GetState().GetPhase()))
			}
			if cluster.GetConfiguration() != nil {
				cfg := cluster.GetConfiguration()
				fmt.Fprintf(w, "Version:  %s\n", cfg.GetVersion())
				fmt.Fprintf(w, "Nodes:    %d\n", cfg.GetNumberOfNodes())
				fmt.Fprintf(w, "Package:  %s\n", cfg.GetPackageId())
			}
			fmt.Fprintf(w, "Cloud:    %s\n", cluster.GetCloudProviderId())
			if isHybrid {
				fmt.Fprintf(w, "Environment: %s\n", cluster.GetCloudProviderRegionId())
			} else {
				fmt.Fprintf(w, "Region:   %s\n", cluster.GetCloudProviderRegionId())
			}
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

				// Hybrid Configuration section (hybrid only)
				if isHybrid {
					hasHybridCfg := cfg.ServiceType != nil || len(cfg.GetNodeSelector()) > 0 ||
						len(cfg.GetTolerations()) > 0 || len(cfg.GetAnnotations()) > 0 ||
						len(cfg.GetPodLabels()) > 0 || len(cfg.GetServiceAnnotations()) > 0 ||
						cfg.ReservedCpuPercentage != nil || cfg.ReservedMemoryPercentage != nil ||
						cfg.GpuType != nil || len(cfg.GetTopologySpreadConstraints()) > 0 ||
						cfg.AdditionalResources != nil

					if hasHybridCfg {
						fmt.Fprintln(w)
						fmt.Fprintln(w, "Hybrid Configuration:")
						if cfg.ServiceType != nil {
							fmt.Fprintf(w, "  Service Type:              %s\n", serviceTypeString(cfg.GetServiceType()))
						}
						if cfg.GpuType != nil {
							fmt.Fprintf(w, "  GPU Type:                  %s\n", gpuTypeString(cfg.GetGpuType()))
						}
						if cfg.ReservedCpuPercentage != nil {
							fmt.Fprintf(w, "  Reserved CPU %%:            %d\n", cfg.GetReservedCpuPercentage())
						}
						if cfg.ReservedMemoryPercentage != nil {
							fmt.Fprintf(w, "  Reserved Memory %%:         %d\n", cfg.GetReservedMemoryPercentage())
						}
						if cfg.AdditionalResources != nil {
							fmt.Fprintf(w, "  Additional Disk (GiB):     %d\n", cfg.GetAdditionalResources().GetDisk())
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
								op := output.TolerationOperator(t.GetOperator())
								effect := output.TolerationEffect(t.GetEffect())
								if t.GetOperator() == clusterv1.TolerationOperator_TOLERATION_OPERATOR_EXISTS {
									fmt.Fprintf(w, "    key=%s operator=%s effect=%s\n", t.GetKey(), op, effect)
								} else {
									fmt.Fprintf(w, "    key=%s value=%s operator=%s effect=%s\n", t.GetKey(), t.GetValue(), op, effect)
								}
							}
						}
						if tscs := cfg.GetTopologySpreadConstraints(); len(tscs) > 0 {
							fmt.Fprintf(w, "  Topology Spread Constraints:\n")
							for _, tsc := range tscs {
								fmt.Fprintf(w, "    topologyKey=%s", tsc.GetTopologyKey())
								if tsc.MaxSkew != nil {
									fmt.Fprintf(w, " maxSkew=%d", tsc.GetMaxSkew())
								}
								if tsc.WhenUnsatisfiable != nil {
									fmt.Fprintf(w, " whenUnsatisfiable=%s", strings.TrimPrefix(tsc.GetWhenUnsatisfiable().String(), "TOPOLOGY_SPREAD_CONSTRAINT_WHEN_UNSATISFIABLE_"))
								}
								fmt.Fprintln(w)
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

					// Storage Configuration section (hybrid only)
					if sc := cfg.GetClusterStorageConfiguration(); sc != nil &&
						(sc.DatabaseStorageClass != nil || sc.SnapshotStorageClass != nil ||
							sc.VolumeSnapshotClass != nil || sc.VolumeAttributesClass != nil) {
						fmt.Fprintln(w)
						fmt.Fprintln(w, "Storage Configuration:")
						if sc.DatabaseStorageClass != nil {
							fmt.Fprintf(w, "  Database Storage Class:    %s\n", sc.GetDatabaseStorageClass())
						}
						if sc.SnapshotStorageClass != nil {
							fmt.Fprintf(w, "  Snapshot Storage Class:    %s\n", sc.GetSnapshotStorageClass())
						}
						if sc.VolumeSnapshotClass != nil {
							fmt.Fprintf(w, "  Volume Snapshot Class:     %s\n", sc.GetVolumeSnapshotClass())
						}
						if sc.VolumeAttributesClass != nil {
							fmt.Fprintf(w, "  Volume Attributes Class:   %s\n", sc.GetVolumeAttributesClass())
						}
					}
				}

				// Database Configuration section (all clusters)
				if dbCfg := cfg.GetDatabaseConfiguration(); dbCfg != nil {
					fmt.Fprintln(w)
					fmt.Fprintln(w, "Database Configuration:")

					if isHybrid && dbCfg.LogLevel != nil {
						fmt.Fprintf(w, "  Log Level:                   %s\n", dbLogLevelString(dbCfg.GetLogLevel()))
					}

					col := dbCfg.GetCollection()
					fmt.Fprintln(w, "  Collection Defaults:")
					if col != nil {
						fmt.Fprintf(w, "    Replication Factor:        %s\n", output.OptionalValue(col.ReplicationFactor, notSet))
						fmt.Fprintf(w, "    Write Consistency Factor:  %s\n", output.OptionalValue(col.WriteConsistencyFactor, notSet))
						if vec := col.GetVectors(); vec != nil {
							if isHybrid && vec.OnDisk != nil {
								fmt.Fprintf(w, "    Vectors on Disk:           %s\n", output.BoolYesNo(vec.GetOnDisk()))
							} else if !isHybrid {
								fmt.Fprintf(w, "    On Disk:                   %s\n", output.OptionalValue(vec.OnDisk, notSet))
							}
						} else if !isHybrid {
							fmt.Fprintf(w, "    On Disk:                   %s\n", notSet)
						}
					} else {
						fmt.Fprintf(w, "    Replication Factor:        %s\n", notSet)
						fmt.Fprintf(w, "    Write Consistency Factor:  %s\n", notSet)
						if !isHybrid {
							fmt.Fprintf(w, "    On Disk:                   %s\n", notSet)
						}
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

					// Hybrid-only DB config sections
					if isHybrid {
						if svc := dbCfg.GetService(); svc != nil &&
							(svc.EnableTls != nil || svc.ApiKey != nil || svc.ReadOnlyApiKey != nil) {
							fmt.Fprintln(w, "  Service:")
							if svc.EnableTls != nil {
								fmt.Fprintf(w, "    Enable TLS:                %s\n", output.BoolYesNo(svc.GetEnableTls()))
							}
							if svc.ApiKey != nil {
								fmt.Fprintf(w, "    API Key Secret:            %s\n", secretKeyRefString(svc.GetApiKey()))
							}
							if svc.ReadOnlyApiKey != nil {
								fmt.Fprintf(w, "    Read-Only API Key Secret:  %s\n", secretKeyRefString(svc.GetReadOnlyApiKey()))
							}
						}
						if tls := dbCfg.GetTls(); tls != nil && (tls.Cert != nil || tls.Key != nil) {
							fmt.Fprintln(w, "  TLS:")
							if tls.Cert != nil {
								fmt.Fprintf(w, "    Cert Secret:               %s\n", secretKeyRefString(tls.GetCert()))
							}
							if tls.Key != nil {
								fmt.Fprintf(w, "    Key Secret:                %s\n", secretKeyRefString(tls.GetKey()))
							}
						}
						if al := dbCfg.GetAuditLogging(); al != nil {
							fmt.Fprintln(w, "  Audit Logging:")
							fmt.Fprintf(w, "    Enabled:                   %s\n", output.BoolYesNo(al.GetEnabled()))
							if al.Rotation != nil {
								fmt.Fprintf(w, "    Rotation:                  %s\n", auditLogRotationString(al.GetRotation()))
							}
							if al.MaxLogFiles != nil {
								fmt.Fprintf(w, "    Max Log Files:             %d\n", al.GetMaxLogFiles())
							}
							if al.TrustForwardedHeaders != nil {
								fmt.Fprintf(w, "    Trust Forwarded Headers:   %s\n", output.BoolYesNo(al.GetTrustForwardedHeaders()))
							}
						}
					}
				}

				// Cluster Configuration section (all clusters)
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

				// Cost Allocation Label (hybrid only)
				if isHybrid && cluster.CostAllocationLabel != nil {
					fmt.Fprintln(w)
					fmt.Fprintf(w, "Cost Allocation Label:  %s\n", cluster.GetCostAllocationLabel())
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
							n.GetName(), output.ClusterNodeState(n.GetState()), n.GetVersion(), started)
					}
				}
			}

			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}
