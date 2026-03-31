package cluster

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newUpdateCommand(s *state.State) *cobra.Command {
	cmd := base.UpdateCmd[*clusterv1.Cluster]{
		Example: `# Add a label to a cluster
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --label env=staging

# Remove a label
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --label env-

# Restrict access to specific IPs
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --allowed-ip 10.0.0.0/8

# Upgrade the Qdrant version
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --version v1.17.0

# Change replication factor (triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --replication-factor 3 --force

# Set service type to load balancer (hybrid only, triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --service-type load-balancer

# Add a node selector and toleration (hybrid only, triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe \
  --node-selector disktype=ssd --toleration "dedicated=qdrant:NoSchedule"

# Remove a node selector (hybrid only)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --node-selector disktype-

# Change database storage class (hybrid only, triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --database-storage-class fast-ssd`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <cluster-id>",
				Short: "Update an existing cluster",
				Long: `Updates the configuration of a cluster.

Use this command to modify cluster settings such as the Qdrant version, labels,
database defaults, IP restrictions, restart mode, rebalance strategy, and hybrid
cluster configuration.

Version upgrades (--version) will trigger a rolling restart of the cluster.

Database configuration changes (--replication-factor, --write-consistency-factor,
--async-scorer, --optimizer-cpu-budget, --vectors-on-disk, --db-log-level,
--audit-logging and related flags, --enable-tls, --api-key-secret,
--read-only-api-key-secret, --tls-cert-secret, --tls-key-secret) will trigger a
rolling restart of the cluster. The cluster remains available during the restart,
but individual nodes will be briefly unavailable as they cycle.

Hybrid cluster configuration changes (--service-type, --node-selector,
--toleration, --topology-spread-constraint, --annotation, --pod-label,
--service-annotation, --reserved-cpu-percentage, --reserved-memory-percentage,
and storage class flags) will also trigger a rolling restart.

Cluster configuration changes (--allowed-ip, --restart-mode, --rebalance-strategy,
--disk-performance, --cost-allocation-label) and label changes take effect without
a restart.

Labels are merged with existing labels by default. Use 'key=value' to add or
overwrite a label, and 'key-' (with a trailing dash) to remove one.

Allowed IPs are merged with existing IPs by default. Specify an IP CIDR to add
it, or append '-' (e.g. '10.0.0.0/8-') to remove one.

Node selectors, annotations, pod labels, and service annotations support the same
'key=value' / 'key-' merge syntax as labels.

Tolerations are merged with existing tolerations. Use 'key-' to remove all
tolerations matching that key.

Topology spread constraints are merged by topologyKey. Use 'topologyKey-' to
remove a constraint.`,
				Args: util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
			addSharedClusterFlags(cmd)
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
				ClusterId: args[0],
				AccountId: accountID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster: %w", err)
			}

			return resp.GetCluster(), nil
		},
		Update: func(s *state.State, cmd *cobra.Command, cluster *clusterv1.Cluster) (*clusterv1.Cluster, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			updated := proto.CloneOf(cluster)

			if err := applySharedClusterFlags(cmd, updated); err != nil {
				return nil, err
			}

			// Determine which changes trigger a rolling restart.
			versionChanged := cmd.Flags().Changed("version")
			dbChanged := util.AnyFlagChanged(cmd, allDBConfigFlags)
			hybridChanged := util.AnyFlagChanged(cmd, hybridConfigFlags)

			if versionChanged || dbChanged || hybridChanged {
				force, _ := cmd.Flags().GetBool("force")
				prompt := updateRestartPrompt(cluster, updated, cmd, versionChanged, dbChanged, hybridChanged)
				if !util.ConfirmAction(force, cmd.ErrOrStderr(), prompt) {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil, nil
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
			fmt.Fprintf(out, "Cluster %s (%s) updated successfully.\n", updated.GetId(), updated.GetName())
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)

	_ = cmd.RegisterFlagCompletionFunc("version", completion.VersionCompletion(s))
	_ = cmd.RegisterFlagCompletionFunc("disk-performance", diskPerformanceCompletion())
	_ = cmd.RegisterFlagCompletionFunc("restart-mode", restartModeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("rebalance-strategy", rebalanceStrategyCompletion())
	_ = cmd.RegisterFlagCompletionFunc("service-type", serviceTypeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("db-log-level", dbLogLevelCompletion())
	_ = cmd.RegisterFlagCompletionFunc("audit-log-rotation", auditLogRotationCompletion())
	return cmd
}

// updateRestartPrompt builds a single confirmation message for all changes that
// trigger a rolling restart (version upgrade, database configuration, and hybrid
// cluster configuration).
func updateRestartPrompt(old, updated *clusterv1.Cluster, cmd *cobra.Command, versionChanged, dbChanged, hybridChanged bool) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Updating cluster %s (%s) will change:", old.GetId(), old.GetName()))

	notSet := "(not set)"

	if versionChanged {
		oldVersion := old.GetState().GetVersion()
		if oldVersion == "" {
			oldVersion = old.GetConfiguration().GetVersion()
		}
		lines = append(lines, fmt.Sprintf("  Version:                         %s", output.DiffValue(oldVersion, updated.GetConfiguration().GetVersion())))
	}

	if dbChanged {
		oldCol := old.GetConfiguration().GetDatabaseConfiguration().GetCollection()
		newCol := updated.GetConfiguration().GetDatabaseConfiguration().GetCollection()
		oldPerf := old.GetConfiguration().GetDatabaseConfiguration().GetStorage().GetPerformance()
		newPerf := updated.GetConfiguration().GetDatabaseConfiguration().GetStorage().GetPerformance()
		oldAudit := old.GetConfiguration().GetDatabaseConfiguration().GetAuditLogging()
		newAudit := updated.GetConfiguration().GetDatabaseConfiguration().GetAuditLogging()
		oldSvc := old.GetConfiguration().GetDatabaseConfiguration().GetService()
		newSvc := updated.GetConfiguration().GetDatabaseConfiguration().GetService()

		if cmd.Flags().Changed("replication-factor") {
			var oldRF *uint32
			if oldCol != nil {
				oldRF = oldCol.ReplicationFactor
			}
			lines = append(lines, fmt.Sprintf("  Replication factor:              %s", output.DiffValue(output.OptionalValue(oldRF, notSet), fmt.Sprintf("%d", newCol.GetReplicationFactor()))))
		}
		if cmd.Flags().Changed("write-consistency-factor") {
			var oldWCF *int32
			if oldCol != nil {
				oldWCF = oldCol.WriteConsistencyFactor
			}
			lines = append(lines, fmt.Sprintf("  Write consistency factor:        %s", output.DiffValue(output.OptionalValue(oldWCF, notSet), fmt.Sprintf("%d", newCol.GetWriteConsistencyFactor()))))
		}
		if cmd.Flags().Changed("vectors-on-disk") {
			var oldVOD *bool
			if oldCol != nil && oldCol.Vectors != nil {
				oldVOD = oldCol.Vectors.OnDisk
			}
			var newVOD *bool
			if newCol != nil && newCol.Vectors != nil {
				newVOD = newCol.Vectors.OnDisk
			}
			lines = append(lines, fmt.Sprintf("  Vectors on disk:                 %s", output.DiffValue(output.OptionalValue(oldVOD, notSet), output.OptionalValue(newVOD, notSet))))
		}
		if cmd.Flags().Changed("async-scorer") {
			var oldAS *bool
			if oldPerf != nil {
				oldAS = oldPerf.AsyncScorer
			}
			lines = append(lines, fmt.Sprintf("  Async scorer:                    %s", output.DiffValue(output.OptionalValue(oldAS, notSet), output.BoolYesNo(newPerf.GetAsyncScorer()))))
		}
		if cmd.Flags().Changed("optimizer-cpu-budget") {
			var oldBudget *int32
			if oldPerf != nil {
				oldBudget = oldPerf.OptimizerCpuBudget
			}
			lines = append(lines, fmt.Sprintf("  Optimizer CPU budget:            %s", output.DiffValue(output.OptionalValue(oldBudget, notSet), fmt.Sprintf("%d", newPerf.GetOptimizerCpuBudget()))))
		}
		if cmd.Flags().Changed("db-log-level") {
			oldLL := old.GetConfiguration().GetDatabaseConfiguration().GetLogLevel()
			newLL := updated.GetConfiguration().GetDatabaseConfiguration().GetLogLevel()
			lines = append(lines, fmt.Sprintf("  DB log level:                    %s", output.DiffValue(dbLogLevelString(oldLL), dbLogLevelString(newLL))))
		}
		if cmd.Flags().Changed("audit-logging") {
			var oldEnabled string
			if oldAudit != nil {
				oldEnabled = output.BoolYesNo(oldAudit.Enabled)
			} else {
				oldEnabled = notSet
			}
			lines = append(lines, fmt.Sprintf("  Audit logging:                   %s", output.DiffValue(oldEnabled, output.BoolYesNo(newAudit.GetEnabled()))))
		}
		if cmd.Flags().Changed("audit-log-rotation") {
			var oldRot string
			if oldAudit != nil {
				oldRot = auditLogRotationString(oldAudit.GetRotation())
			}
			if oldRot == "" {
				oldRot = notSet
			}
			lines = append(lines, fmt.Sprintf("  Audit log rotation:              %s", output.DiffValue(oldRot, auditLogRotationString(newAudit.GetRotation()))))
		}
		if cmd.Flags().Changed("audit-log-max-files") {
			var oldMax *uint32
			if oldAudit != nil {
				oldMax = oldAudit.MaxLogFiles
			}
			lines = append(lines, fmt.Sprintf("  Audit log max files:             %s", output.DiffValue(output.OptionalValue(oldMax, notSet), fmt.Sprintf("%d", newAudit.GetMaxLogFiles()))))
		}
		if cmd.Flags().Changed("audit-log-trust-forwarded-headers") {
			var oldTFH *bool
			if oldAudit != nil {
				oldTFH = oldAudit.TrustForwardedHeaders
			}
			lines = append(lines, fmt.Sprintf("  Audit log trust fwd headers:     %s", output.DiffValue(output.OptionalValue(oldTFH, notSet), output.BoolYesNo(newAudit.GetTrustForwardedHeaders()))))
		}
		if cmd.Flags().Changed("enable-tls") {
			var oldTLS *bool
			if oldSvc != nil {
				oldTLS = oldSvc.EnableTls
			}
			lines = append(lines, fmt.Sprintf("  Enable TLS:                      %s", output.DiffValue(output.OptionalValue(oldTLS, notSet), output.BoolYesNo(newSvc.GetEnableTls()))))
		}
		if cmd.Flags().Changed("api-key-secret") {
			lines = append(lines, "  API key secret:                  (changed)")
		}
		if cmd.Flags().Changed("read-only-api-key-secret") {
			lines = append(lines, "  Read-only API key secret:        (changed)")
		}
		if cmd.Flags().Changed("tls-cert-secret") {
			lines = append(lines, "  TLS cert secret:                 (changed)")
		}
		if cmd.Flags().Changed("tls-key-secret") {
			lines = append(lines, "  TLS key secret:                  (changed)")
		}
	}

	if hybridChanged {
		if cmd.Flags().Changed("service-type") {
			oldST := old.GetConfiguration().GetServiceType()
			newST := updated.GetConfiguration().GetServiceType()
			lines = append(lines, fmt.Sprintf("  Service type:                    %s", output.DiffValue(serviceTypeString(oldST), serviceTypeString(newST))))
		}
		if cmd.Flags().Changed("reserved-cpu-percentage") {
			var oldPct *uint32
			if cfg := old.GetConfiguration(); cfg != nil {
				oldPct = cfg.ReservedCpuPercentage
			}
			lines = append(lines, fmt.Sprintf("  Reserved CPU %%:                  %s", output.DiffValue(output.OptionalValue(oldPct, notSet), fmt.Sprintf("%d", updated.GetConfiguration().GetReservedCpuPercentage()))))
		}
		if cmd.Flags().Changed("reserved-memory-percentage") {
			var oldPct *uint32
			if cfg := old.GetConfiguration(); cfg != nil {
				oldPct = cfg.ReservedMemoryPercentage
			}
			lines = append(lines, fmt.Sprintf("  Reserved memory %%:               %s", output.DiffValue(output.OptionalValue(oldPct, notSet), fmt.Sprintf("%d", updated.GetConfiguration().GetReservedMemoryPercentage()))))
		}
		for _, flag := range undiffableFlags {
			if cmd.Flags().Changed(flag) {
				lines = append(lines, fmt.Sprintf("  %-32s (changed)", flag+":"))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, "WARNING: These changes will result in a rolling restart of your cluster.")
	lines = append(lines, "Proceed?")
	return strings.Join(lines, "\n")
}
