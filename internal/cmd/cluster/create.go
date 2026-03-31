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
			cmd.Flags().String("cloud-provider", "", "Cloud provider ID (required, see 'cloud-provider list)")
			cmd.Flags().String("cloud-region", "", "Cloud provider region ID (required, see 'cloud-region list --cloud-provider <provider_id>)")
			cmd.Flags().String("version", "", "Qdrant version (default latest)")
			cmd.Flags().Uint32("nodes", 1, "Number of nodes (default 1)")
			cmd.Flags().String("package", "", "Booking package name or ID (see 'cluster package list')")
			cmd.Flags().Var(new(resource.Millicores), "cpu", "CPU to select a package (e.g. \"1\", \"0.5\", or \"1000m\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "ram", "RAM to select a package (e.g. \"8\", \"8G\", \"8Gi\", or \"8GiB\")")
			cmd.Flags().Var(new(resource.ByteQuantity), "disk", "Total disk size (e.g. \"200GiB\"); if larger than the package's included disk, the difference is provisioned as additional storage")
			cmd.Flags().Var(new(resource.Millicores), "gpu", "Number of GPUs to select a package (e.g. \"1\", \"2\", or \"1000m\")")
			cmd.Flags().Bool("multi-az", false, "Require a multi-AZ package")
			cmd.Flags().StringArray("label", nil, "Label to apply to the cluster ('key=value'), can be specified multiple times")
			cmd.Flags().String("disk-performance", "", `Disk performance tier ("balanced", "cost-optimised", "performance")`)
			cmd.Flags().Uint32("replication-factor", 0, "Default replication factor for new collections")
			cmd.Flags().Int32("write-consistency-factor", 0, "Default write consistency factor for new collections")
			cmd.Flags().Bool("async-scorer", false, "Enable async scorer (uses io_uring on Linux)")
			cmd.Flags().Int32("optimizer-cpu-budget", 0, `CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)`)
			cmd.Flags().StringArray("allowed-ip", nil, "Allowed client IP CIDR ranges; max 20")
			cmd.Flags().String("restart-mode", "", `Restart policy ("rolling", "parallel", "automatic")`)
			cmd.Flags().String("rebalance-strategy", "", `Shard rebalance strategy ("by-count", "by-size", "by-count-and-size")`)
			cmd.Flags().Bool("vectors-on-disk", false, "Store vectors in memmap storage")

			// Audit logging
			cmd.Flags().Bool("audit-logging", false, "Enable audit logging")
			cmd.Flags().String("audit-log-rotation", "", `Audit log rotation ("daily", "hourly")`)
			cmd.Flags().Uint32("audit-log-max-files", 0, "Maximum number of audit log files (1-1000)")
			cmd.Flags().Bool("audit-log-trust-forwarded-headers", false, "Trust forwarded headers in audit logs")

			cmd.Flags().String("cost-allocation-label", "", "Label for billing reports")

			// Hybrid Cluster flags
			cmd.Flags().String("service-type", "", `(cloud-provider: hybrid) Kubernetes service type ("cluster-ip", "node-port", "load-balancer")`)
			cmd.Flags().StringArray("node-selector", nil, "Node selector label ('key=value'), can be specified multiple times")
			cmd.Flags().StringArray("toleration", nil, "Toleration ('key=value:Effect' or 'key:Exists:Effect'), can be specified multiple times")
			cmd.Flags().StringArray("topology-spread-constraint", nil, "Topology spread constraint ('topologyKey[:maxSkew[:whenUnsatisfiable]]'), can be specified multiple times")
			cmd.Flags().StringArray("annotation", nil, "Annotation ('key=value'), can be specified multiple times")
			cmd.Flags().StringArray("pod-label", nil, "Pod label ('key=value'), can be specified multiple times")
			cmd.Flags().StringArray("service-annotation", nil, "Service annotation ('key=value'), can be specified multiple times")
			cmd.Flags().Uint32("reserved-cpu-percentage", 0, "Percentage of CPU reserved for system components, 1-80 (default 20)")
			cmd.Flags().Uint32("reserved-memory-percentage", 0, "Percentage of memory reserved for system components, 1-80 (default 20)")
			cmd.Flags().String("database-storage-class", "", "Kubernetes storage class for database volumes")
			cmd.Flags().String("snapshot-storage-class", "", "Kubernetes storage class for snapshot volumes")
			cmd.Flags().String("volume-snapshot-class", "", "Kubernetes volume snapshot class")
			cmd.Flags().String("volume-attributes-class", "", "Kubernetes volume attributes class")
			cmd.Flags().String("db-log-level", "", `Database log level ("trace", "debug", "info", "warn", "error", "off")`)
			cmd.Flags().String("api-key-secret", "", "API key Kubernetes secret ('secretName:key')")
			cmd.Flags().String("read-only-api-key-secret", "", "Read-only API key Kubernetes secret ('secretName:key')")
			cmd.Flags().Bool("enable-tls", false, "Enable TLS for the database service")
			cmd.Flags().String("tls-cert-secret", "", "TLS certificate Kubernetes secret ('secretName:key')")
			cmd.Flags().String("tls-key-secret", "", "TLS private key Kubernetes secret ('secretName:key')")

			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
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
						pkg,
							err = clusterutil.ResolvePackageByID(ctx,
							client.Booking(),
							accountID,
							cloudProvider,
							&cloudRegion,
							packageValue,
						)
						if err != nil {
							return nil, err
						}
					}
				} else {
					pkg, err = clusterutil.ResolvePackageByName(
						ctx,
						client.Booking(),
						accountID,
						cloudProvider,
						&cloudRegion,
						packageValue,
					)
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

			if cmd.Flags().Changed("cost-allocation-label") {
				v, _ := cmd.Flags().GetString("cost-allocation-label")
				cluster.CostAllocationLabel = &v
			}

			// Storage configuration
			if util.AnyFlagChanged(cmd, storageConfigFlags) {
				if cluster.Configuration.ClusterStorageConfiguration == nil {
					cluster.Configuration.ClusterStorageConfiguration = &clusterv1.ClusterStorageConfiguration{}
				}
				sc := cluster.Configuration.ClusterStorageConfiguration
				if cmd.Flags().Changed("database-storage-class") {
					v, _ := cmd.Flags().GetString("database-storage-class")
					sc.DatabaseStorageClass = &v
				}
				if cmd.Flags().Changed("snapshot-storage-class") {
					v, _ := cmd.Flags().GetString("snapshot-storage-class")
					sc.SnapshotStorageClass = &v
				}
				if cmd.Flags().Changed("volume-snapshot-class") {
					v, _ := cmd.Flags().GetString("volume-snapshot-class")
					sc.VolumeSnapshotClass = &v
				}
				if cmd.Flags().Changed("volume-attributes-class") {
					v, _ := cmd.Flags().GetString("volume-attributes-class")
					sc.VolumeAttributesClass = &v
				}
			}

			// Database configuration flags
			if util.AnyFlagChanged(cmd, allDBConfigFlags) {
				if cluster.Configuration.DatabaseConfiguration == nil {
					cluster.Configuration.DatabaseConfiguration = &clusterv1.DatabaseConfiguration{}
				}
				dbCfg := cluster.Configuration.DatabaseConfiguration

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

				if util.AnyFlagChanged(cmd, collectionFlags) {
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
					if cmd.Flags().Changed("vectors-on-disk") {
						v, _ := cmd.Flags().GetBool("vectors-on-disk")
						if dbCfg.Collection.Vectors == nil {
							dbCfg.Collection.Vectors = &clusterv1.DatabaseConfigurationCollectionVectors{}
						}
						dbCfg.Collection.Vectors.OnDisk = &v
					}
				}

				if cmd.Flags().Changed("db-log-level") {
					v, _ := cmd.Flags().GetString("db-log-level")
					ll, err := parseDBLogLevel(v)
					if err != nil {
						return nil, err
					}
					dbCfg.LogLevel = ll.Enum()
				}

				if util.AnyFlagChanged(cmd, serviceFlags) {
					if dbCfg.Service == nil {
						dbCfg.Service = &clusterv1.DatabaseConfigurationService{}
					}
					if cmd.Flags().Changed("enable-tls") {
						v, _ := cmd.Flags().GetBool("enable-tls")
						dbCfg.Service.EnableTls = &v
					}
					if cmd.Flags().Changed("api-key-secret") {
						v, _ := cmd.Flags().GetString("api-key-secret")
						ref, err := parseSecretKeyRef(v)
						if err != nil {
							return nil, fmt.Errorf("--api-key-secret: %w", err)
						}
						dbCfg.Service.ApiKey = ref
					}
					if cmd.Flags().Changed("read-only-api-key-secret") {
						v, _ := cmd.Flags().GetString("read-only-api-key-secret")
						ref, err := parseSecretKeyRef(v)
						if err != nil {
							return nil, fmt.Errorf("--read-only-api-key-secret: %w", err)
						}
						dbCfg.Service.ReadOnlyApiKey = ref
					}
				}

				if util.AnyFlagChanged(cmd, tlsFlags) {
					if dbCfg.Tls == nil {
						dbCfg.Tls = &clusterv1.DatabaseConfigurationTls{}
					}
					if cmd.Flags().Changed("tls-cert-secret") {
						v, _ := cmd.Flags().GetString("tls-cert-secret")
						ref, err := parseSecretKeyRef(v)
						if err != nil {
							return nil, fmt.Errorf("--tls-cert-secret: %w", err)
						}
						dbCfg.Tls.Cert = ref
					}
					if cmd.Flags().Changed("tls-key-secret") {
						v, _ := cmd.Flags().GetString("tls-key-secret")
						ref, err := parseSecretKeyRef(v)
						if err != nil {
							return nil, fmt.Errorf("--tls-key-secret: %w", err)
						}
						dbCfg.Tls.Key = ref
					}
				}

				if util.AnyFlagChanged(cmd, auditLoggingFlags) {
					if dbCfg.AuditLogging == nil {
						dbCfg.AuditLogging = &clusterv1.DatabaseConfigurationAuditLogging{}
					}
					if cmd.Flags().Changed("audit-logging") {
						v, _ := cmd.Flags().GetBool("audit-logging")
						dbCfg.AuditLogging.Enabled = v
					}
					if cmd.Flags().Changed("audit-log-rotation") {
						v, _ := cmd.Flags().GetString("audit-log-rotation")
						r, err := parseAuditLogRotation(v)
						if err != nil {
							return nil, err
						}
						dbCfg.AuditLogging.Rotation = r.Enum()
					}
					if cmd.Flags().Changed("audit-log-max-files") {
						v, _ := cmd.Flags().GetUint32("audit-log-max-files")
						dbCfg.AuditLogging.MaxLogFiles = &v
					}
					if cmd.Flags().Changed("audit-log-trust-forwarded-headers") {
						v, _ := cmd.Flags().GetBool("audit-log-trust-forwarded-headers")
						dbCfg.AuditLogging.TrustForwardedHeaders = &v
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

			// Hybrid Cloud flags
			if cmd.Flags().Changed("node-selector") {
				raw, _ := cmd.Flags().GetStringArray("node-selector")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--node-selector: %w", err)
				}
				for k, v := range changes.Set {
					cluster.Configuration.NodeSelector = append(cluster.Configuration.NodeSelector, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("annotation") {
				raw, _ := cmd.Flags().GetStringArray("annotation")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--annotation: %w", err)
				}
				for k, v := range changes.Set {
					cluster.Configuration.Annotations = append(cluster.Configuration.Annotations, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("pod-label") {
				raw, _ := cmd.Flags().GetStringArray("pod-label")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--pod-label: %w", err)
				}
				for k, v := range changes.Set {
					cluster.Configuration.PodLabels = append(cluster.Configuration.PodLabels, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("service-annotation") {
				raw, _ := cmd.Flags().GetStringArray("service-annotation")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--service-annotation: %w", err)
				}
				for k, v := range changes.Set {
					cluster.Configuration.ServiceAnnotations = append(cluster.Configuration.ServiceAnnotations, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("reserved-cpu-percentage") {
				v, _ := cmd.Flags().GetUint32("reserved-cpu-percentage")
				cluster.Configuration.ReservedCpuPercentage = &v
			}

			if cmd.Flags().Changed("reserved-memory-percentage") {
				v, _ := cmd.Flags().GetUint32("reserved-memory-percentage")
				cluster.Configuration.ReservedMemoryPercentage = &v
			}

			if cmd.Flags().Changed("toleration") {
				rawTols, _ := cmd.Flags().GetStringArray("toleration")
				for _, raw := range rawTols {
					tol, err := parseToleration(raw)
					if err != nil {
						return nil, err
					}
					cluster.Configuration.Tolerations = append(cluster.Configuration.Tolerations, tol)
				}
			}

			if cmd.Flags().Changed("topology-spread-constraint") {
				raw, _ := cmd.Flags().GetStringArray("topology-spread-constraint")
				for _, r := range raw {
					tsc, err := parseTopologySpreadConstraint(r)
					if err != nil {
						return nil, err
					}
					cluster.Configuration.TopologySpreadConstraints = append(cluster.Configuration.TopologySpreadConstraints, tsc)
				}
			}

			if cmd.Flags().Changed("service-type") {
				v, _ := cmd.Flags().GetString("service-type")
				st, err := parseServiceType(v)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.ServiceType = st.Enum()
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
	_ = cmd.RegisterFlagCompletionFunc("package", completion.PackageNameCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("version", completion.VersionCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("cpu", completion.CPUCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("ram", completion.RAMCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk", completion.DiskCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("gpu", completion.GPUCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk-performance", diskPerformanceCompletion())
	_ = cmd.RegisterFlagCompletionFunc("restart-mode", restartModeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("rebalance-strategy", rebalanceStrategyCompletion())
	_ = cmd.RegisterFlagCompletionFunc("service-type", serviceTypeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("db-log-level", dbLogLevelCompletion())
	_ = cmd.RegisterFlagCompletionFunc("audit-log-rotation", auditLogRotationCompletion())
	return cmd
}
