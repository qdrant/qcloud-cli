package cluster

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	bookingv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/booking/v1"
	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/clusterutil"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*clusterv1.Cluster]{
		Example: `# Create a free-tier cluster
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --package free

# Create a cluster with specific resources
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --cpu 0.5 --ram 4Gi

# Create a cluster and wait for it to become healthy
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --cpu 2 --ram 8Gi --wait

# Create with labels and extra disk
qcloud cluster create --cloud-provider aws --cloud-region eu-central-1 --cpu 4 --ram 32Gi \
  --disk 200Gi --label env=production --label team=search`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a new cluster",
				Args:  cobra.NoArgs,
			}
			cmd.Flags().String("name", "", "Cluster name (auto-generated if not provided)")
			cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required, see 'cluster cloud-provider list)")
			cmd.Flags().String("cloud-region", "", "Cloud provider region ID (required, see 'cluster cloud-region list --cloud-provider <provider_id>)")
			cmd.Flags().String("version", "", "Qdrant version (default latest)")
			cmd.Flags().Uint32("nodes", 1, "Number of nodes (default 1)")
			cmd.Flags().String("package", "", "Booking package name or ID (see 'cluster package list')")
			cmd.Flags().Var(new(resource.Millicores), "cpu", "CPU to select a package (e.g. \"1\", \"0.5\", or \"1000m\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "ram", "RAM to select a package (e.g. \"8\", \"8G\", \"8Gi\", or \"8GiB\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "disk", "Total disk size (e.g. \"200GiB\"); if larger than the package's included disk, the difference is provisioned as additional storage")
			cmd.Flags().Var(new(resource.Millicores), "gpu", "Number of GPUs to select a package (e.g. \"1\", \"2\", or \"1000m\")")
			cmd.Flags().Bool("multi-az", false, "Require a multi-AZ package")
			cmd.Flags().StringArray("label", nil, "Label to apply to the cluster ('key=value'), can be specified multiple times")
			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
			cmd.Flags().String("disk-performance", "", `Disk performance tier ("balanced", "cost-optimised", "performance")`)
			cmd.Flags().Uint32("replication-factor", 0, "Default replication factor for new collections")
			cmd.Flags().Int32("write-consistency-factor", 0, "Default write consistency factor for new collections")
			cmd.Flags().Bool("async-scorer", false, "Enable async scorer (uses io_uring on Linux)")
			cmd.Flags().Int32("optimizer-cpu-budget", 0, `CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)`)
			cmd.Flags().StringArray("allowed-ip", nil, "Allowed client IP CIDR ranges; max 20")
			cmd.Flags().String("restart-mode", "", `Restart policy ("rolling", "parallel", "automatic")`)
			cmd.Flags().String("rebalance-strategy", "", `Shard rebalance strategy ("by-count", "by-size", "by-count-and-size")`)
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
			_ = cmd.MarkFlagRequired("cloud-provider")
			_ = cmd.MarkFlagRequired("cloud-region")
			return cmd
		},
		Run: func(s *state.State, cmd *cobra.Command, args []string) (*clusterv1.Cluster, error) {
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
			if name == "" {
				suggested, err := client.Cluster().SuggestClusterName(ctx, &clusterv1.SuggestClusterNameRequest{
					AccountId: accountID,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to suggest cluster name: %w", err)
				}
				name = suggested.GetName()
			}
			cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
			cloudRegion, _ := cmd.Flags().GetString("cloud-region")
			version, _ := cmd.Flags().GetString("version")
			nodes, _ := cmd.Flags().GetUint32("nodes")
			packageValue, _ := cmd.Flags().GetString("package")
			multiAz, _ := cmd.Flags().GetBool("multi-az")
			rawLabels, _ := cmd.Flags().GetStringArray("label")

			cpuChanged := cmd.Flags().Changed("cpu")
			ramChanged := cmd.Flags().Changed("ram")

			var cpu resource.Millicores
			var ram resource.ByteQuantity
			var gpu resource.Millicores
			if cpuChanged {
				cpu = *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
			}
			if ramChanged {
				ram = *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
			}
			if cmd.Flags().Changed("gpu") {
				gpu = *cmd.Flags().Lookup("gpu").Value.(*resource.Millicores)
			}

			if packageValue == "" && !cpuChanged && !ramChanged {
				return nil, fmt.Errorf("either --package or --cpu and --ram are required")
			}

			var pkg *bookingv1.Package
			var packageID string

			if packageValue != "" {
				if util.IsUUID(packageValue) {
					packageID = packageValue
					if cmd.Flags().Changed("disk") {
						pkg, err = clusterutil.ResolvePackageByID(ctx, client.Booking(), accountID, cloudProvider, &cloudRegion, packageValue)
						if err != nil {
							return nil, err
						}
					}
				} else {
					pkg, err = clusterutil.ResolvePackageByName(ctx, client.Booking(), accountID, cloudProvider, &cloudRegion, packageValue)
					if err != nil {
						return nil, err
					}
					packageID = pkg.GetId()
				}
			} else {
				pkg, err = clusterutil.ResolvePackageByResources(ctx, client.Booking(), clusterutil.PackageResourceQuery{
					AccountID:     accountID,
					CloudProvider: cloudProvider,
					CloudRegion:   &cloudRegion,
					CPU:           cpu,
					GPU:           gpu,
					RAM:           ram,
					MultiAz:       multiAz,
				})
				if err != nil {
					return nil, err
				}
				packageID = pkg.GetId()
			}

			cluster := &clusterv1.Cluster{
				AccountId:             accountID,
				Name:                  name,
				CloudProviderId:       cloudProvider,
				CloudProviderRegionId: cloudRegion,
				Configuration: &clusterv1.ClusterConfiguration{
					NumberOfNodes: nodes,
				},
			}
			if version != "" {
				cluster.Configuration.Version = &version
			}
			if packageID != "" {
				cluster.Configuration.PackageId = packageID
			}
			labelChanges, err := util.ParseLabels(rawLabels)
			if err != nil {
				return nil, err
			}
			for k, v := range labelChanges.Set {
				cluster.Labels = append(cluster.Labels, &commonv1.KeyValue{Key: k, Value: v})
			}

			if cmd.Flags().Changed("disk-performance") {
				perfStr, _ := cmd.Flags().GetString("disk-performance")
				tierType, err := parseDiskPerformance(perfStr)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.ClusterStorageConfiguration = &clusterv1.ClusterStorageConfiguration{
					StorageTierType: tierType,
				}
			}

			// Database configuration flags
			if util.AnyFlagChanged(cmd, []string{"replication-factor", "write-consistency-factor", "async-scorer", "optimizer-cpu-budget"}) {
				if cluster.Configuration.DatabaseConfiguration == nil {
					cluster.Configuration.DatabaseConfiguration = &clusterv1.DatabaseConfiguration{}
				}
				dbCfg := cluster.Configuration.DatabaseConfiguration

				if util.AnyFlagChanged(cmd, []string{"replication-factor", "write-consistency-factor"}) {
					if dbCfg.Collection == nil {
						dbCfg.Collection = &clusterv1.DatabaseConfigurationCollection{}
					}
					if cmd.Flags().Changed("replication-factor") {
						v, _ := cmd.Flags().GetUint32("replication-factor")
						dbCfg.Collection.ReplicationFactor = &v
					}
					if cmd.Flags().Changed("write-consistency-factor") {
						v, _ := cmd.Flags().GetInt32("write-consistency-factor")
						dbCfg.Collection.WriteConsistencyFactor = &v
					}
				}

				if util.AnyFlagChanged(cmd, []string{"async-scorer", "optimizer-cpu-budget"}) {
					if dbCfg.Storage == nil {
						dbCfg.Storage = &clusterv1.DatabaseConfigurationStorage{}
					}
					if dbCfg.Storage.Performance == nil {
						dbCfg.Storage.Performance = &clusterv1.DatabaseConfigurationStoragePerformance{}
					}
					if cmd.Flags().Changed("async-scorer") {
						v, _ := cmd.Flags().GetBool("async-scorer")
						dbCfg.Storage.Performance.AsyncScorer = &v
					}
					if cmd.Flags().Changed("optimizer-cpu-budget") {
						v, _ := cmd.Flags().GetInt32("optimizer-cpu-budget")
						dbCfg.Storage.Performance.OptimizerCpuBudget = &v
					}
				}
			}

			// Cluster configuration flags
			if cmd.Flags().Changed("allowed-ip") {
				raw, _ := cmd.Flags().GetStringArray("allowed-ip")
				changes, err := util.ParseIPs(raw)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.AllowedIpSourceRanges = changes.Add
			}

			if cmd.Flags().Changed("restart-mode") {
				modeStr, _ := cmd.Flags().GetString("restart-mode")
				mode, err := parseRestartMode(modeStr)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.RestartPolicy = mode.Enum()
			}

			if cmd.Flags().Changed("rebalance-strategy") {
				stratStr, _ := cmd.Flags().GetString("rebalance-strategy")
				strat, err := parseRebalanceStrategy(stratStr)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.RebalanceStrategy = strat.Enum()
			}

			if cmd.Flags().Changed("disk") && pkg != nil {
				requestedDisk := *cmd.Flags().Lookup("disk").Value.(*resource.ByteQuantity)
				additionalDisk, err := clusterutil.CalculateAdditionalDisk(requestedDisk, pkg)
				if err != nil {
					return nil, err
				}
				if additionalDisk > 0 {
					cluster.Configuration.AdditionalResources = &clusterv1.AdditionalResources{
						Disk: additionalDisk,
					}
				}
			}

			resp, err := client.Cluster().CreateCluster(ctx, &clusterv1.CreateClusterRequest{
				Cluster: cluster,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create cluster: %w", err)
			}

			created := resp.GetCluster()

			wait, _ := cmd.Flags().GetBool("wait")
			if !wait {
				return created, nil
			}

			waitTimeout, _ := cmd.Flags().GetDuration("wait-timeout")
			pollInterval, _ := cmd.Flags().GetDuration("wait-poll-interval")
			fmt.Fprintf(cmd.ErrOrStderr(), "Cluster %s created, waiting for it to become healthy...\n", created.GetId())
			return clusterutil.WaitForClusterHealthy(ctx, client.Cluster(), cmd.ErrOrStderr(), accountID, created.GetId(), waitTimeout, pollInterval)
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *clusterv1.Cluster) {
			if ep := created.GetState().GetEndpoint(); ep != nil && ep.GetUrl() != "" {
				fmt.Fprintf(out, "Cluster %s (%s) is ready. Endpoint: %s\n", created.GetId(), created.GetName(), ep.GetUrl())
			} else {
				fmt.Fprintf(out, "Cluster %s (%s) created successfully.\n", created.GetId(), created.GetName())
			}
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("cloud-provider", completion.CloudProviderCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cloud-region", completion.CloudRegionCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("package", packageCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("version", completion.VersionCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cpu", cpuCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("ram", ramCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk", diskCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("gpu", gpuCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk-performance", diskPerformanceCompletion())
	_ = cmd.RegisterFlagCompletionFunc("restart-mode", restartModeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("rebalance-strategy", rebalanceStrategyCompletion())
	return cmd
}
