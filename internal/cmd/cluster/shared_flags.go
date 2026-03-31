package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/util"
)

// addSharedClusterFlags registers all cluster flags shared between the create and
// update commands. Create-only flags (cloud-provider, cloud-region, package, cpu,
// ram, disk, gpu, multi-az, nodes, name, wait, wait-timeout, wait-poll-interval)
// are NOT included here.
func addSharedClusterFlags(cmd *cobra.Command) {
	cmd.Flags().String("version", "", `Qdrant version (e.g. "v1.17.0" or "latest")`)
	cmd.Flags().StringArray("label", nil, "Label ('key=value') to add/overwrite; append '-' to remove ('key-'), can be specified multiple times")
	cmd.Flags().String("disk-performance", "", `Disk performance tier ("balanced", "cost-optimised", "performance")`)
	cmd.Flags().Bool("async-scorer", false, "Enable async scorer (uses io_uring on Linux)")
	cmd.Flags().Int32("optimizer-cpu-budget", 0, `CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)`)
	cmd.Flags().StringArray("allowed-ip", nil, "Allowed client IP CIDR range (e.g. \"10.0.0.0/8\"); append '-' to remove; max 20")
	cmd.Flags().String("restart-mode", "", `Restart policy ("rolling", "parallel", "automatic")`)
	cmd.Flags().String("rebalance-strategy", "", `Shard rebalance strategy ("by-count", "by-size", "by-count-and-size")`)
	cmd.Flags().String("cost-allocation-label", "", "Label for billing reports")

	// Collection
	cmd.Flags().Uint32("replication-factor", 0, "Default replication factor for new collections")
	cmd.Flags().Int32("write-consistency-factor", 0, "Default write consistency factor for new collections")
	cmd.Flags().Bool("vectors-on-disk", false, "Store vectors in memmap storage for new collections")

	// Audit logging
	cmd.Flags().Bool("audit-logging", false, "Enable audit logging")
	cmd.Flags().String("audit-log-rotation", "", `Audit log rotation ("daily", "hourly")`)
	cmd.Flags().Uint32("audit-log-max-files", 0, "Maximum number of audit log files (1-1000)")
	cmd.Flags().Bool("audit-log-trust-forwarded-headers", false, "Trust forwarded headers in audit logs")

	// Hybrid Cluster flags
	cmd.Flags().String("service-type", "", `(cloud-provider: hybrid) Kubernetes service type ("cluster-ip", "node-port", "load-balancer")`)
	cmd.Flags().StringArray("node-selector", nil, "(cloud-provider: hybrid) Node selector label ('key=value'); append '-' to remove, can be specified multiple times")
	cmd.Flags().StringArray("toleration", nil, "(cloud-provider: hybrid) Toleration ('key=value:Effect' or 'key:Exists:Effect'); use 'key-' to remove by key, can be specified multiple times")
	cmd.Flags().StringArray("topology-spread-constraint", nil, "(cloud-provider: hybrid) Topology spread constraint ('topologyKey[:maxSkew[:whenUnsatisfiable]]'); use 'topologyKey-' to remove, can be specified multiple times")
	cmd.Flags().StringArray("annotation", nil, "(cloud-provider: hybrid) Pod annotation ('key=value'); append '-' to remove, can be specified multiple times")
	cmd.Flags().StringArray("pod-label", nil, "(cloud-provider: hybrid) Pod label ('key=value'); append '-' to remove, can be specified multiple times")
	cmd.Flags().StringArray("service-annotation", nil, "(cloud-provider: hybrid) Service annotation ('key=value'); append '-' to remove, can be specified multiple times")
	cmd.Flags().Uint32("reserved-cpu-percentage", 0, "(cloud-provider: hybrid) Percentage of CPU reserved for system components, 1-80 (default 20)")
	cmd.Flags().Uint32("reserved-memory-percentage", 0, "(cloud-provider: hybrid) Percentage of memory reserved for system components, 1-80 (default 20)")
	cmd.Flags().String("database-storage-class", "", "(cloud-provider: hybrid) Kubernetes storage class for database volumes")
	cmd.Flags().String("snapshot-storage-class", "", "(cloud-provider: hybrid) Kubernetes storage class for snapshot volumes")
	cmd.Flags().String("volume-snapshot-class", "", "(cloud-provider: hybrid) Kubernetes volume snapshot class")
	cmd.Flags().String("volume-attributes-class", "", "(cloud-provider: hybrid) Kubernetes volume attributes class")
	cmd.Flags().String("db-log-level", "", `Database log level ("trace", "debug", "info", "warn", "error", "off")`)
	cmd.Flags().String("api-key-secret", "", "(cloud-provider: hybrid) API key Kubernetes secret ('secretName:key')")
	cmd.Flags().String("read-only-api-key-secret", "", "(cloud-provider: hybrid) Read-only API key Kubernetes secret ('secretName:key')")
	cmd.Flags().Bool("enable-tls", false, "(cloud-provider: hybrid) Enable TLS for the database service")
	cmd.Flags().String("tls-cert-secret", "", "(cloud-provider: hybrid) TLS certificate Kubernetes secret ('secretName:key')")
	cmd.Flags().String("tls-key-secret", "", "(cloud-provider: hybrid) TLS private key Kubernetes secret ('secretName:key')")
}

// applySharedClusterFlags applies all shared flags to cluster. It works for both
// create (empty cluster, nil existing fields) and update (pre-cloned cluster with
// existing values). Key-value fields (labels, node-selector, etc.) are merged via
// util.ApplyLabels so that both 'key=value' adds and 'key-' removes work in both
// commands.
func applySharedClusterFlags(cmd *cobra.Command, cluster *clusterv1.Cluster) error {
	// Labels
	if cmd.Flags().Changed("label") {
		raw, _ := cmd.Flags().GetStringArray("label")
		changes, err := util.ParseLabels(raw)
		if err != nil {
			return err
		}
		cluster.Labels = util.ApplyLabels(cluster.Labels, changes)
	}

	// Ensure configuration exists.
	if cluster.Configuration == nil {
		cluster.Configuration = &clusterv1.ClusterConfiguration{}
	}
	cfg := cluster.Configuration

	// Version
	if cmd.Flags().Changed("version") {
		v, _ := cmd.Flags().GetString("version")
		cfg.Version = &v
	}

	// Disk performance
	if cmd.Flags().Changed("disk-performance") {
		perfStr, _ := cmd.Flags().GetString("disk-performance")
		tierType, err := parseDiskPerformance(perfStr)
		if err != nil {
			return err
		}
		if cfg.ClusterStorageConfiguration == nil {
			cfg.ClusterStorageConfiguration = &clusterv1.ClusterStorageConfiguration{}
		}
		cfg.ClusterStorageConfiguration.StorageTierType = tierType
	}

	// Cost allocation label
	if cmd.Flags().Changed("cost-allocation-label") {
		v, _ := cmd.Flags().GetString("cost-allocation-label")
		cluster.CostAllocationLabel = &v
	}

	// Storage config flags
	if util.AnyFlagChanged(cmd, storageConfigFlags) {
		if cfg.ClusterStorageConfiguration == nil {
			cfg.ClusterStorageConfiguration = &clusterv1.ClusterStorageConfiguration{}
		}
		sc := cfg.ClusterStorageConfiguration
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
		if cfg.DatabaseConfiguration == nil {
			cfg.DatabaseConfiguration = &clusterv1.DatabaseConfiguration{}
		}
		dbCfg := cfg.DatabaseConfiguration

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
				return err
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
					return fmt.Errorf("--api-key-secret: %w", err)
				}
				dbCfg.Service.ApiKey = ref
			}
			if cmd.Flags().Changed("read-only-api-key-secret") {
				v, _ := cmd.Flags().GetString("read-only-api-key-secret")
				ref, err := parseSecretKeyRef(v)
				if err != nil {
					return fmt.Errorf("--read-only-api-key-secret: %w", err)
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
					return fmt.Errorf("--tls-cert-secret: %w", err)
				}
				dbCfg.Tls.Cert = ref
			}
			if cmd.Flags().Changed("tls-key-secret") {
				v, _ := cmd.Flags().GetString("tls-key-secret")
				ref, err := parseSecretKeyRef(v)
				if err != nil {
					return fmt.Errorf("--tls-key-secret: %w", err)
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
					return err
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

	// Allowed IPs (merged: nil existing in create = set; existing values in update = merge)
	if cmd.Flags().Changed("allowed-ip") {
		raw, _ := cmd.Flags().GetStringArray("allowed-ip")
		changes, err := util.ParseIPs(raw)
		if err != nil {
			return err
		}
		cfg.AllowedIpSourceRanges = util.ApplyIPs(cfg.AllowedIpSourceRanges, changes)
	}

	// Restart mode
	if cmd.Flags().Changed("restart-mode") {
		modeStr, _ := cmd.Flags().GetString("restart-mode")
		mode, err := parseRestartMode(modeStr)
		if err != nil {
			return err
		}
		cfg.RestartPolicy = mode.Enum()
	}

	// Rebalance strategy
	if cmd.Flags().Changed("rebalance-strategy") {
		stratStr, _ := cmd.Flags().GetString("rebalance-strategy")
		strat, err := parseRebalanceStrategy(stratStr)
		if err != nil {
			return err
		}
		cfg.RebalanceStrategy = strat.Enum()
	}

	// Hybrid flags
	if cmd.Flags().Changed("node-selector") {
		raw, _ := cmd.Flags().GetStringArray("node-selector")
		changes, err := util.ParseLabels(raw)
		if err != nil {
			return fmt.Errorf("--node-selector: %w", err)
		}
		cfg.NodeSelector = util.ApplyLabels(cfg.NodeSelector, changes)
	}

	if cmd.Flags().Changed("annotation") {
		raw, _ := cmd.Flags().GetStringArray("annotation")
		changes, err := util.ParseLabels(raw)
		if err != nil {
			return fmt.Errorf("--annotation: %w", err)
		}
		cfg.Annotations = util.ApplyLabels(cfg.Annotations, changes)
	}

	if cmd.Flags().Changed("pod-label") {
		raw, _ := cmd.Flags().GetStringArray("pod-label")
		changes, err := util.ParseLabels(raw)
		if err != nil {
			return fmt.Errorf("--pod-label: %w", err)
		}
		cfg.PodLabels = util.ApplyLabels(cfg.PodLabels, changes)
	}

	if cmd.Flags().Changed("service-annotation") {
		raw, _ := cmd.Flags().GetStringArray("service-annotation")
		changes, err := util.ParseLabels(raw)
		if err != nil {
			return fmt.Errorf("--service-annotation: %w", err)
		}
		cfg.ServiceAnnotations = util.ApplyLabels(cfg.ServiceAnnotations, changes)
	}

	if cmd.Flags().Changed("reserved-cpu-percentage") {
		v, _ := cmd.Flags().GetUint32("reserved-cpu-percentage")
		cfg.ReservedCpuPercentage = &v
	}

	if cmd.Flags().Changed("reserved-memory-percentage") {
		v, _ := cmd.Flags().GetUint32("reserved-memory-percentage")
		cfg.ReservedMemoryPercentage = &v
	}

	if cmd.Flags().Changed("toleration") {
		raw, _ := cmd.Flags().GetStringArray("toleration")
		tols, err := applyTolerations(cfg.Tolerations, raw)
		if err != nil {
			return err
		}
		cfg.Tolerations = tols
	}

	if cmd.Flags().Changed("topology-spread-constraint") {
		raw, _ := cmd.Flags().GetStringArray("topology-spread-constraint")
		tscs, err := applyTopologySpreadConstraints(cfg.TopologySpreadConstraints, raw)
		if err != nil {
			return err
		}
		cfg.TopologySpreadConstraints = tscs
	}

	if cmd.Flags().Changed("service-type") {
		v, _ := cmd.Flags().GetString("service-type")
		st, err := parseServiceType(v)
		if err != nil {
			return err
		}
		cfg.ServiceType = st.Enum()
	}

	return nil
}
