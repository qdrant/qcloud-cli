package hybrid

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
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/resource"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newClusterCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*clusterv1.Cluster]{
		ValidArgsFunction: envIDCompletion(s),
		Example: `# Create a cluster with defaults
qcloud hybrid cluster create 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Create a cluster with a specific package
qcloud hybrid cluster create 7b2ea926-724b-4de2-b73a-8675c42a6ebe --package my-package

# Create a cluster by CPU and RAM
qcloud hybrid cluster create 7b2ea926-724b-4de2-b73a-8675c42a6ebe --cpu 2 --ram 8Gi

# Create with extra disk beyond the package default
qcloud hybrid cluster create 7b2ea926-724b-4de2-b73a-8675c42a6ebe --cpu 4 --ram 32Gi --disk 200Gi

# Create a named 3-node cluster and wait for it to become healthy
qcloud hybrid cluster create 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --name my-cluster --nodes 3 --wait

# Create with node selectors and tolerations
qcloud hybrid cluster create 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --node-selector dedicated=qdrant --toleration dedicated=qdrant:NoSchedule`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create <env-id>",
				Short: "Create a cluster in a hybrid cloud environment",
				Long:  "Create a new Qdrant cluster inside a hybrid cloud environment.",
				Args:  util.ExactArgs(1, "a hybrid cloud environment ID"),
			}
			cmd.Flags().String("name", "", "Cluster name (auto-generated if not provided)")
			cmd.Flags().Uint32("nodes", 1, "Number of nodes")
			cmd.Flags().String("version", "", "Qdrant version (default latest)")
			cmd.Flags().String("service-type", "", `Kubernetes service type ("cluster-ip", "node-port", "load-balancer")`)
			cmd.Flags().StringArray("node-selector", nil, "Node selector label ('key=value'), can be specified multiple times")
			cmd.Flags().StringArray("annotation", nil, "Annotation ('key=value'), can be specified multiple times")
			cmd.Flags().StringArray("pod-label", nil, "Pod label ('key=value'), can be specified multiple times")
			cmd.Flags().StringArray("service-annotation", nil, "Service annotation ('key=value'), can be specified multiple times")
			cmd.Flags().Uint32("reserved-cpu-percentage", 0, "Percentage of CPU reserved for system components, 1-80 (default 20)")
			cmd.Flags().Uint32("reserved-memory-percentage", 0, "Percentage of memory reserved for system components, 1-80 (default 20)")
			cmd.Flags().StringArray("toleration", nil, "Toleration ('key=value:Effect' or 'key:Exists:Effect'), can be specified multiple times")
			cmd.Flags().StringArray("label", nil, "Cluster label ('key=value'), can be specified multiple times")
			cmd.Flags().Uint32("replication-factor", 0, "Default replication factor for new collections")
			cmd.Flags().Int32("write-consistency-factor", 0, "Default write consistency factor for new collections")
			cmd.Flags().Bool("async-scorer", false, "Enable async scorer (uses io_uring on Linux)")
			cmd.Flags().Int32("optimizer-cpu-budget", 0, "CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)")
			cmd.Flags().StringArray("allowed-ip", nil, "Allowed client IP CIDR ranges; max 20")
			cmd.Flags().String("restart-policy", "", `Restart policy ("rolling", "parallel", "automatic")`)
			cmd.Flags().String("rebalance-strategy", "", `Rebalance strategy ("by-count", "by-size", "by-count-and-size")`)
			cmd.Flags().StringArray("topology-spread-constraint", nil, "Topology spread constraint ('topologyKey[:maxSkew[:whenUnsatisfiable]]'), can be specified multiple times")
			cmd.Flags().String("database-storage-class", "", "Kubernetes storage class for database volumes")
			cmd.Flags().String("snapshot-storage-class", "", "Kubernetes storage class for snapshot volumes")
			cmd.Flags().String("volume-snapshot-class", "", "Kubernetes volume snapshot class")
			cmd.Flags().String("volume-attributes-class", "", "Kubernetes volume attributes class")
			cmd.Flags().String("package", "", "Booking package name or ID (see 'cluster package list --cloud-provider hybrid')")
			cmd.Flags().Var(new(resource.Millicores), "cpu", `CPU to select a package (e.g. "1", "0.5", or "1000m")`)
			cmd.Flags().Var(new(resource.ByteQuantity), "ram", `RAM to select a package (e.g. "8", "8G", "8Gi", or "8GiB")`)
			cmd.Flags().Var(new(resource.ByteQuantity), "disk", `Total disk size (e.g. "200GiB"); if larger than the package's included disk, the difference is provisioned as additional storage`)
			cmd.Flags().String("db-log-level", "", `Database log level ("trace", "debug", "info", "warn", "error", "off")`)
			cmd.Flags().Bool("vectors-on-disk", false, "Store vectors in memmap storage")
			cmd.Flags().Bool("enable-tls", false, "Enable TLS for the database service")
			cmd.Flags().String("api-key-secret", "", "API key Kubernetes secret ('secretName:key')")
			cmd.Flags().String("read-only-api-key-secret", "", "Read-only API key Kubernetes secret ('secretName:key')")
			cmd.Flags().String("tls-cert-secret", "", "TLS certificate Kubernetes secret ('secretName:key')")
			cmd.Flags().String("tls-key-secret", "", "TLS private key Kubernetes secret ('secretName:key')")
			cmd.Flags().Bool("audit-logging", false, "Enable audit logging")
			cmd.Flags().String("audit-log-rotation", "", `Audit log rotation ("daily", "hourly")`)
			cmd.Flags().Uint32("audit-log-max-files", 0, "Maximum number of audit log files (1-1000)")
			cmd.Flags().Bool("audit-log-trust-forwarded-headers", false, "Trust forwarded headers in audit logs")
			cmd.Flags().String("cost-allocation-label", "", "Label for billing reports")
			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
			_ = cmd.Flags().MarkHidden("wait-poll-interval")
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

			envID := args[0]

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

			nodes, _ := cmd.Flags().GetUint32("nodes")
			version, _ := cmd.Flags().GetString("version")

			cluster := &clusterv1.Cluster{
				AccountId:             accountID,
				Name:                  name,
				CloudProviderId:       qcloudapi.HybridCloudProviderID,
				CloudProviderRegionId: envID,
				Configuration: &clusterv1.ClusterConfiguration{
					NumberOfNodes: nodes,
				},
			}

			if version != "" {
				cluster.Configuration.Version = &version
			}

			// Package resolution
			packageValue, _ := cmd.Flags().GetString("package")
			cpuChanged := cmd.Flags().Changed("cpu")
			ramChanged := cmd.Flags().Changed("ram")

			if packageValue != "" && (cpuChanged || ramChanged) {
				return nil, fmt.Errorf("--package and --cpu/--ram are mutually exclusive")
			}

			var pkg *bookingv1.Package

			if packageValue != "" {
				if util.IsUUID(packageValue) {
					cluster.Configuration.PackageId = packageValue
					if cmd.Flags().Changed("disk") {
						pkg, err = clusterutil.ResolvePackageByID(ctx, client.Booking(), accountID, qcloudapi.HybridCloudProviderID, nil, packageValue)
						if err != nil {
							return nil, err
						}
					}
				} else {
					pkg, err = clusterutil.ResolvePackageByName(ctx, client.Booking(), accountID, qcloudapi.HybridCloudProviderID, nil, packageValue)
					if err != nil {
						return nil, err
					}
					cluster.Configuration.PackageId = pkg.GetId()
				}
			} else {
				var cpu resource.Millicores
				var ram resource.ByteQuantity
				if cpuChanged {
					cpu = *cmd.Flags().Lookup("cpu").Value.(*resource.Millicores)
				}
				if ramChanged {
					ram = *cmd.Flags().Lookup("ram").Value.(*resource.ByteQuantity)
				}
				pkg, err = clusterutil.ResolvePackageByResources(ctx, client.Booking(), clusterutil.PackageResourceQuery{
					AccountID:     accountID,
					CloudProvider: qcloudapi.HybridCloudProviderID,
					CPU:           cpu,
					RAM:           ram,
				})
				if err != nil {
					return nil, err
				}
				cluster.Configuration.PackageId = pkg.GetId()
			}

			// Disk calculation
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

			// Labels
			if cmd.Flags().Changed("label") {
				rawLabels, _ := cmd.Flags().GetStringArray("label")
				changes, err := util.ParseLabels(rawLabels)
				if err != nil {
					return nil, err
				}
				for k, v := range changes.Set {
					cluster.Labels = append(cluster.Labels, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			// Service type
			if cmd.Flags().Changed("service-type") {
				stStr, _ := cmd.Flags().GetString("service-type")
				st, err := parseServiceType(stStr)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.ServiceType = st.Enum()
			}

			// Hybrid-specific cluster configuration fields (set directly on ClusterConfiguration)
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

			if cmd.Flags().Changed("restart-policy") {
				v, _ := cmd.Flags().GetString("restart-policy")
				rp, err := parseRestartPolicy(v)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.RestartPolicy = rp.Enum()
			}

			if cmd.Flags().Changed("rebalance-strategy") {
				v, _ := cmd.Flags().GetString("rebalance-strategy")
				rs, err := parseRebalanceStrategy(v)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.RebalanceStrategy = rs.Enum()
			}

			if cmd.Flags().Changed("gpu-type") {
				v, _ := cmd.Flags().GetString("gpu-type")
				gt, err := parseGpuType(v)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.GpuType = gt.Enum()
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

			if cmd.Flags().Changed("cost-allocation-label") {
				v, _ := cmd.Flags().GetString("cost-allocation-label")
				cluster.CostAllocationLabel = &v
			}

			// Database configuration
			if util.AnyFlagChanged(cmd, hybridClusterDBConfigFlags) {
				if cluster.Configuration.DatabaseConfiguration == nil {
					cluster.Configuration.DatabaseConfiguration = &clusterv1.DatabaseConfiguration{}
				}
				dbCfg := cluster.Configuration.DatabaseConfiguration

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

				if util.AnyFlagChanged(cmd, performanceFlags) {
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

			// Allowed IPs
			if cmd.Flags().Changed("allowed-ip") {
				raw, _ := cmd.Flags().GetStringArray("allowed-ip")
				changes, err := util.ParseIPs(raw)
				if err != nil {
					return nil, err
				}
				cluster.Configuration.AllowedIpSourceRanges = changes.Add
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
			fmt.Fprintf(out, "Cluster %s (%s) created.\n", created.GetId(), created.GetName())
		},
	}.CobraCommand(s)
	hybridProviderFn := func(_ *cobra.Command) (string, *string) { return qcloudapi.HybridCloudProviderID, nil }
	_ = cmd.RegisterFlagCompletionFunc("package", completion.PackageNameCompletion(s, hybridProviderFn))
	_ = cmd.RegisterFlagCompletionFunc("cpu", completion.CPUCompletion(s, hybridProviderFn))
	_ = cmd.RegisterFlagCompletionFunc("ram", completion.RAMCompletion(s, hybridProviderFn))
	_ = cmd.RegisterFlagCompletionFunc("disk", completion.DiskCompletion(s, hybridProviderFn))
	_ = cmd.RegisterFlagCompletionFunc("gpu", completion.GPUCompletion(s, hybridProviderFn))
	_ = cmd.RegisterFlagCompletionFunc("service-type", serviceTypeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("version", completion.VersionCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("restart-policy", restartPolicyCompletion())
	_ = cmd.RegisterFlagCompletionFunc("rebalance-strategy", rebalanceStrategyCompletion())
	_ = cmd.RegisterFlagCompletionFunc("gpu-type", gpuTypeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("db-log-level", dbLogLevelCompletion())
	_ = cmd.RegisterFlagCompletionFunc("audit-log-rotation", auditLogRotationCompletion())
	return cmd
}
