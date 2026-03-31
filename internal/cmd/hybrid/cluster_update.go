package hybrid

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newClusterUpdateCommand(s *state.State) *cobra.Command {
	cmd := base.UpdateCmd[*clusterv1.Cluster]{
		ValidArgsFunction: hybridClusterIDCompletion(s),
		Example: `# Add a label to a cluster
qcloud hybrid cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --label env=staging

# Change the service type
qcloud hybrid cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --service-type load-balancer

# Update database configuration (triggers rolling restart)
qcloud hybrid cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --optimizer-cpu-budget 4`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <cluster-id>",
				Short: "Update a cluster in a hybrid cloud environment",
				Args:  util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().StringArray("label", nil, "Label to set ('key=value') or remove ('key-'); merges with existing labels")
			cmd.Flags().StringArray("allowed-ip", nil, "Allowed IP CIDR to add or remove (suffix with '-'); merges with existing IPs")
			cmd.Flags().String("service-type", "", `Kubernetes service type ("cluster-ip", "node-port", "load-balancer")`)
			cmd.Flags().StringArray("node-selector", nil, "Node selector label ('key=value'), replaces existing node selectors")
			cmd.Flags().StringArray("toleration", nil, "Toleration ('key=value:Effect' or 'key:Exists:Effect'), replaces existing tolerations")
			cmd.Flags().StringArray("annotation", nil, "Annotation ('key=value'), replaces existing annotations")
			cmd.Flags().StringArray("pod-label", nil, "Pod label ('key=value'), replaces existing pod labels")
			cmd.Flags().StringArray("service-annotation", nil, "Service annotation ('key=value'), replaces existing service annotations")
			cmd.Flags().Uint32("replication-factor", 0, "Default replication factor for new collections")
			cmd.Flags().Int32("write-consistency-factor", 0, "Default write consistency factor for new collections")
			cmd.Flags().Bool("async-scorer", false, "Enable async scorer (uses io_uring on Linux)")
			cmd.Flags().Int32("optimizer-cpu-budget", 0, "CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)")
			cmd.Flags().String("restart-policy", "", `Restart policy ("rolling", "parallel", "automatic")`)
			cmd.Flags().String("rebalance-strategy", "", `Rebalance strategy ("by-count", "by-size", "by-count-and-size")`)
			cmd.Flags().StringArray("topology-spread-constraint", nil, "Topology spread constraint ('topologyKey[:maxSkew[:whenUnsatisfiable]]'), replaces existing constraints")
			cmd.Flags().String("database-storage-class", "", "Kubernetes storage class for database volumes")
			cmd.Flags().String("snapshot-storage-class", "", "Kubernetes storage class for snapshot volumes")
			cmd.Flags().String("volume-snapshot-class", "", "Kubernetes volume snapshot class")
			cmd.Flags().String("volume-attributes-class", "", "Kubernetes volume attributes class")
			cmd.Flags().Uint32("additional-disk", 0, "Additional disk in GiB")
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
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt for database configuration changes")
			return cmd
		},
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

			cluster := resp.GetCluster()
			if cluster.GetCloudProviderId() != qcloudapi.HybridCloudProviderID {
				return nil, fmt.Errorf("cluster %s is not a hybrid cloud cluster; use \"qcloud cluster update\" instead", args[0])
			}

			return cluster, nil
		},
		Update: func(s *state.State, cmd *cobra.Command, cluster *clusterv1.Cluster) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			updated := proto.CloneOf(cluster)

			if updated.Configuration == nil {
				updated.Configuration = &clusterv1.ClusterConfiguration{}
			}
			cfg := updated.Configuration

			// Labels
			if cmd.Flags().Changed("label") {
				raw, _ := cmd.Flags().GetStringArray("label")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, err
				}
				updated.Labels = util.ApplyLabels(updated.Labels, changes)
			}

			// Allowed IPs
			if cmd.Flags().Changed("allowed-ip") {
				raw, _ := cmd.Flags().GetStringArray("allowed-ip")
				changes, err := util.ParseIPs(raw)
				if err != nil {
					return nil, err
				}
				cfg.AllowedIpSourceRanges = util.ApplyIPs(cfg.AllowedIpSourceRanges, changes)
			}

			// Hybrid configuration
			if cmd.Flags().Changed("service-type") {
				stStr, _ := cmd.Flags().GetString("service-type")
				st, err := parseServiceType(stStr)
				if err != nil {
					return nil, err
				}
				cfg.ServiceType = st.Enum()
			}

			if cmd.Flags().Changed("node-selector") {
				raw, _ := cmd.Flags().GetStringArray("node-selector")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--node-selector: %w", err)
				}
				cfg.NodeSelector = nil
				for k, v := range changes.Set {
					cfg.NodeSelector = append(cfg.NodeSelector, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("toleration") {
				rawTols, _ := cmd.Flags().GetStringArray("toleration")
				cfg.Tolerations = nil
				for _, raw := range rawTols {
					tol, err := parseToleration(raw)
					if err != nil {
						return nil, err
					}
					cfg.Tolerations = append(cfg.Tolerations, tol)
				}
			}

			if cmd.Flags().Changed("annotation") {
				raw, _ := cmd.Flags().GetStringArray("annotation")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--annotation: %w", err)
				}
				cfg.Annotations = nil
				for k, v := range changes.Set {
					cfg.Annotations = append(cfg.Annotations, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("pod-label") {
				raw, _ := cmd.Flags().GetStringArray("pod-label")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--pod-label: %w", err)
				}
				cfg.PodLabels = nil
				for k, v := range changes.Set {
					cfg.PodLabels = append(cfg.PodLabels, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("service-annotation") {
				raw, _ := cmd.Flags().GetStringArray("service-annotation")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, fmt.Errorf("--service-annotation: %w", err)
				}
				cfg.ServiceAnnotations = nil
				for k, v := range changes.Set {
					cfg.ServiceAnnotations = append(cfg.ServiceAnnotations, &commonv1.KeyValue{Key: k, Value: v})
				}
			}

			if cmd.Flags().Changed("restart-policy") {
				v, _ := cmd.Flags().GetString("restart-policy")
				rp, err := parseRestartPolicy(v)
				if err != nil {
					return nil, err
				}
				cfg.RestartPolicy = rp.Enum()
			}

			if cmd.Flags().Changed("rebalance-strategy") {
				v, _ := cmd.Flags().GetString("rebalance-strategy")
				rs, err := parseRebalanceStrategy(v)
				if err != nil {
					return nil, err
				}
				cfg.RebalanceStrategy = rs.Enum()
			}

			if cmd.Flags().Changed("topology-spread-constraint") {
				raw, _ := cmd.Flags().GetStringArray("topology-spread-constraint")
				cfg.TopologySpreadConstraints = nil
				for _, r := range raw {
					tsc, err := parseTopologySpreadConstraint(r)
					if err != nil {
						return nil, err
					}
					cfg.TopologySpreadConstraints = append(cfg.TopologySpreadConstraints, tsc)
				}
			}

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

			if cmd.Flags().Changed("additional-disk") {
				v, _ := cmd.Flags().GetUint32("additional-disk")
				cfg.AdditionalResources = &clusterv1.AdditionalResources{Disk: v}
			}

			if cmd.Flags().Changed("cost-allocation-label") {
				v, _ := cmd.Flags().GetString("cost-allocation-label")
				updated.CostAllocationLabel = &v
			}

			// Database configuration (triggers rolling restart)
			dbChanged := util.AnyFlagChanged(cmd, hybridClusterDBConfigFlags)

			if dbChanged {
				if cfg.DatabaseConfiguration == nil {
					cfg.DatabaseConfiguration = &clusterv1.DatabaseConfiguration{}
				}
				dbCfg := cfg.DatabaseConfiguration

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

				force, _ := cmd.Flags().GetBool("force")
				prompt := hybridClusterUpdateDBPrompt(cluster, updated, cmd)
				if !util.ConfirmAction(force, cmd.ErrOrStderr(), prompt) {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil, nil //nolint:nilnil
				}
			}

			resp, err := client.Cluster().UpdateCluster(ctx, &clusterv1.UpdateClusterRequest{
				Cluster: updated,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update cluster: %w", err)
			}

			return resp.GetCluster(), nil
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, updated *clusterv1.Cluster) {
			if updated == nil {
				return
			}
			fmt.Fprintf(out, "Cluster %s (%s) updated.\n", updated.GetId(), updated.GetName())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("service-type", serviceTypeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("restart-policy", restartPolicyCompletion())
	_ = cmd.RegisterFlagCompletionFunc("rebalance-strategy", rebalanceStrategyCompletion())
	_ = cmd.RegisterFlagCompletionFunc("gpu-type", gpuTypeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("db-log-level", dbLogLevelCompletion())
	_ = cmd.RegisterFlagCompletionFunc("audit-log-rotation", auditLogRotationCompletion())
	return cmd
}

func hybridClusterUpdateDBPrompt(old, updated *clusterv1.Cluster, cmd *cobra.Command) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Updating cluster %s (%s) will change:", old.GetId(), old.GetName()))

	oldDB := old.GetConfiguration().GetDatabaseConfiguration()
	newDB := updated.GetConfiguration().GetDatabaseConfiguration()
	oldCol := oldDB.GetCollection()
	newCol := newDB.GetCollection()
	oldPerf := oldDB.GetStorage().GetPerformance()
	newPerf := newDB.GetStorage().GetPerformance()

	notSet := "(not set)"

	if cmd.Flags().Changed("replication-factor") {
		var oldRF *uint32
		if oldCol != nil {
			oldRF = oldCol.ReplicationFactor
		}
		lines = append(lines, fmt.Sprintf("  Replication factor:       %s", output.DiffValue(output.OptionalValue(oldRF, notSet), fmt.Sprintf("%d", newCol.GetReplicationFactor()))))
	}
	if cmd.Flags().Changed("write-consistency-factor") {
		var oldWCF *int32
		if oldCol != nil {
			oldWCF = oldCol.WriteConsistencyFactor
		}
		lines = append(lines, fmt.Sprintf("  Write consistency factor: %s", output.DiffValue(output.OptionalValue(oldWCF, notSet), fmt.Sprintf("%d", newCol.GetWriteConsistencyFactor()))))
	}
	if cmd.Flags().Changed("vectors-on-disk") {
		var oldV *bool
		if oldCol.GetVectors() != nil {
			oldV = oldCol.GetVectors().OnDisk
		}
		lines = append(lines, fmt.Sprintf("  Vectors on disk:          %s", output.DiffValue(output.OptionalValue(oldV, notSet), output.BoolYesNo(newCol.GetVectors().GetOnDisk()))))
	}
	if cmd.Flags().Changed("async-scorer") {
		var oldAS *bool
		if oldPerf != nil {
			oldAS = oldPerf.AsyncScorer
		}
		lines = append(lines, fmt.Sprintf("  Async scorer:             %s", output.DiffValue(output.OptionalValue(oldAS, notSet), output.BoolYesNo(newPerf.GetAsyncScorer()))))
	}
	if cmd.Flags().Changed("optimizer-cpu-budget") {
		var oldBudget *int32
		if oldPerf != nil {
			oldBudget = oldPerf.OptimizerCpuBudget
		}
		lines = append(lines, fmt.Sprintf("  Optimizer CPU budget:     %s", output.DiffValue(output.OptionalValue(oldBudget, notSet), fmt.Sprintf("%d", newPerf.GetOptimizerCpuBudget()))))
	}
	if cmd.Flags().Changed("db-log-level") {
		oldLL := dbLogLevelString(oldDB.GetLogLevel())
		if oldLL == "" {
			oldLL = notSet
		}
		lines = append(lines, fmt.Sprintf("  DB log level:             %s", output.DiffValue(oldLL, dbLogLevelString(newDB.GetLogLevel()))))
	}
	if cmd.Flags().Changed("enable-tls") {
		var oldV *bool
		if oldDB.GetService() != nil {
			oldV = oldDB.GetService().EnableTls
		}
		lines = append(lines, fmt.Sprintf("  Enable TLS:               %s", output.DiffValue(output.OptionalValue(oldV, notSet), output.BoolYesNo(newDB.GetService().GetEnableTls()))))
	}
	if cmd.Flags().Changed("api-key-secret") {
		oldV := secretKeyRefString(oldDB.GetService().GetApiKey())
		if oldV == "" {
			oldV = notSet
		}
		lines = append(lines, fmt.Sprintf("  API key secret:           %s", output.DiffValue(oldV, secretKeyRefString(newDB.GetService().GetApiKey()))))
	}
	if cmd.Flags().Changed("read-only-api-key-secret") {
		oldV := secretKeyRefString(oldDB.GetService().GetReadOnlyApiKey())
		if oldV == "" {
			oldV = notSet
		}
		lines = append(lines, fmt.Sprintf("  Read-only API key secret: %s", output.DiffValue(oldV, secretKeyRefString(newDB.GetService().GetReadOnlyApiKey()))))
	}
	if cmd.Flags().Changed("tls-cert-secret") {
		oldV := secretKeyRefString(oldDB.GetTls().GetCert())
		if oldV == "" {
			oldV = notSet
		}
		lines = append(lines, fmt.Sprintf("  TLS cert secret:          %s", output.DiffValue(oldV, secretKeyRefString(newDB.GetTls().GetCert()))))
	}
	if cmd.Flags().Changed("tls-key-secret") {
		oldV := secretKeyRefString(oldDB.GetTls().GetKey())
		if oldV == "" {
			oldV = notSet
		}
		lines = append(lines, fmt.Sprintf("  TLS key secret:           %s", output.DiffValue(oldV, secretKeyRefString(newDB.GetTls().GetKey()))))
	}
	if cmd.Flags().Changed("audit-logging") {
		oldV := oldDB.GetAuditLogging().GetEnabled()
		lines = append(lines, fmt.Sprintf("  Audit logging:            %s", output.DiffValue(output.BoolYesNo(oldV), output.BoolYesNo(newDB.GetAuditLogging().GetEnabled()))))
	}
	if cmd.Flags().Changed("audit-log-rotation") {
		oldV := auditLogRotationString(oldDB.GetAuditLogging().GetRotation())
		if oldV == "" {
			oldV = notSet
		}
		lines = append(lines, fmt.Sprintf("  Audit log rotation:       %s", output.DiffValue(oldV, auditLogRotationString(newDB.GetAuditLogging().GetRotation()))))
	}
	if cmd.Flags().Changed("audit-log-max-files") {
		var oldV *uint32
		if oldDB.GetAuditLogging() != nil {
			oldV = oldDB.GetAuditLogging().MaxLogFiles
		}
		lines = append(lines, fmt.Sprintf("  Audit log max files:      %s", output.DiffValue(output.OptionalValue(oldV, notSet), fmt.Sprintf("%d", newDB.GetAuditLogging().GetMaxLogFiles()))))
	}
	if cmd.Flags().Changed("audit-log-trust-forwarded-headers") {
		var oldV *bool
		if oldDB.GetAuditLogging() != nil {
			oldV = oldDB.GetAuditLogging().TrustForwardedHeaders
		}
		lines = append(lines, fmt.Sprintf("  Trust forwarded headers:  %s", output.DiffValue(output.OptionalValue(oldV, notSet), output.BoolYesNo(newDB.GetAuditLogging().GetTrustForwardedHeaders()))))
	}

	lines = append(lines, "")
	lines = append(lines, "WARNING: Database configuration changes will result in a rolling restart of your cluster.")
	lines = append(lines, "Proceed?")
	return strings.Join(lines, "\n")
}
