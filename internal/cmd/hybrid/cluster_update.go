package hybrid

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/state"
)

var hybridClusterDBConfigFlags = []string{
	"replication-factor",
	"write-consistency-factor",
	"async-scorer",
	"optimizer-cpu-budget",
}

func newClusterUpdateCommand(s *state.State) *cobra.Command {
	return base.UpdateCmd[*clusterv1.Cluster]{
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
			cmd.Flags().Int32("optimizer-cpu-budget", 0, "CPU threads for optimization (0=auto, negative=subtract from available, positive=exact)")
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

			return resp.GetCluster(), nil
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

			// Database configuration (triggers rolling restart)
			dbChanged := slices.ContainsFunc(hybridClusterDBConfigFlags, func(f string) bool {
				return cmd.Flags().Changed(f)
			})

			if dbChanged {
				if cfg.DatabaseConfiguration == nil {
					cfg.DatabaseConfiguration = &clusterv1.DatabaseConfiguration{}
				}
				dbCfg := cfg.DatabaseConfiguration

				if cmd.Flags().Changed("replication-factor") || cmd.Flags().Changed("write-consistency-factor") {
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

				if cmd.Flags().Changed("async-scorer") || cmd.Flags().Changed("optimizer-cpu-budget") {
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
}

func hybridClusterUpdateDBPrompt(old, updated *clusterv1.Cluster, cmd *cobra.Command) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Updating cluster %s (%s) will change:", old.GetId(), old.GetName()))

	oldCol := old.GetConfiguration().GetDatabaseConfiguration().GetCollection()
	newCol := updated.GetConfiguration().GetDatabaseConfiguration().GetCollection()
	oldPerf := old.GetConfiguration().GetDatabaseConfiguration().GetStorage().GetPerformance()
	newPerf := updated.GetConfiguration().GetDatabaseConfiguration().GetStorage().GetPerformance()

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
	if cmd.Flags().Changed("async-scorer") {
		var oldAS *bool
		if oldPerf != nil {
			oldAS = oldPerf.AsyncScorer
		}
		lines = append(lines, fmt.Sprintf("  Async scorer:             %s", output.DiffValue(output.OptionalValue(oldAS, notSet), boolToYesNo(newPerf.GetAsyncScorer()))))
	}
	if cmd.Flags().Changed("optimizer-cpu-budget") {
		var oldBudget *int32
		if oldPerf != nil {
			oldBudget = oldPerf.OptimizerCpuBudget
		}
		lines = append(lines, fmt.Sprintf("  Optimizer CPU budget:     %s", output.DiffValue(output.OptionalValue(oldBudget, notSet), fmt.Sprintf("%d", newPerf.GetOptimizerCpuBudget()))))
	}

	lines = append(lines, "")
	lines = append(lines, "WARNING: Database configuration changes will result in a rolling restart of your cluster.")
	lines = append(lines, "Proceed?")
	return strings.Join(lines, "\n")
}
