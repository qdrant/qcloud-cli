package cluster

import (
	"fmt"
	"io"
	"slices"
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

// dbConfigFlags lists flags that trigger a rolling restart.
var dbConfigFlags = []string{
	"replication-factor",
	"write-consistency-factor",
	"async-scorer",
	"optimizer-cpu-budget",
}

func newUpdateCommand(s *state.State) *cobra.Command {
	cmd := base.UpdateCmd[*clusterv1.Cluster]{
		Example: `# Add a label to a cluster
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --label env=staging

# Remove a label
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --label env-

# Restrict access to specific IPs
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --allowed-ip 10.0.0.0/8

# Change replication factor (triggers rolling restart)
qcloud cluster update 7b2ea926-724b-4de2-b73a-8675c42a6ebe --replication-factor 3 --force`,
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "update <cluster-id>",
				Short: "Update an existing cluster",
				Long: `Updates the configuration of a cluster.

Use this command to modify cluster settings such as labels, database defaults,
IP restrictions, restart mode, and rebalance strategy.

Database configuration changes (--replication-factor, --write-consistency-factor,
--async-scorer, --optimizer-cpu-budget) will trigger a rolling restart of the
cluster. The cluster remains available during the restart, but individual nodes
will be briefly unavailable as they cycle.

Cluster configuration changes (--allowed-ip, --restart-mode, --rebalance-strategy)
and label changes take effect without a restart.

Labels are merged with existing labels by default. Use 'key=value' to add or
overwrite a label, and 'key-' (with a trailing dash) to remove one.

Allowed IPs are merged with existing IPs by default. Specify an IP CIDR to add
it, or append '-' (e.g. '10.0.0.0/8-') to remove one.`,
				Args: util.ExactArgs(1, "a cluster ID"),
			}
			cmd.Flags().StringArray("label", nil, "Label to set ('key=value') or remove ('key-'); merges with existing labels")
			cmd.Flags().Uint32("replication-factor", 0, "Default replication factor for new collections")
			cmd.Flags().Int32("write-consistency-factor", 0, "Default write consistency factor for new collections")
			cmd.Flags().Bool("async-scorer", false, "Enable async scorer (uses io_uring on Linux)")
			cmd.Flags().Int32("optimizer-cpu-budget", 0, `CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)`)
			cmd.Flags().StringArray("allowed-ip", nil, "Allowed IP CIDR to add or remove (suffix with '-'); merges with existing IPs")
			cmd.Flags().String("restart-mode", "", `Restart policy ("rolling", "parallel", "automatic")`)
			cmd.Flags().String("rebalance-strategy", "", `Shard rebalance strategy ("by-count", "by-size", "by-count-and-size")`)
			cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
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

			cluster := resp.GetCluster()
			if cluster.GetCloudProviderId() == hybridCloudProviderID {
				return nil, fmt.Errorf("cluster %s is a hybrid cloud cluster; use \"qcloud hybrid cluster update\" instead", args[0])
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

			// Labels
			if cmd.Flags().Changed("label") {
				raw, _ := cmd.Flags().GetStringArray("label")
				changes, err := util.ParseLabels(raw)
				if err != nil {
					return nil, err
				}
				updated.Labels = util.ApplyLabels(updated.Labels, changes)
			}

			// Ensure configuration exists
			if updated.Configuration == nil {
				updated.Configuration = &clusterv1.ClusterConfiguration{}
			}
			cfg := updated.Configuration

			// --- Database configuration flags (trigger rolling restart) ---

			dbChanged := slices.ContainsFunc(dbConfigFlags, func(f string) bool {
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

				// Confirmation prompt for rolling restart
				force, _ := cmd.Flags().GetBool("force")
				prompt := updateDBConfigPrompt(cluster, updated, cmd)
				if !util.ConfirmAction(force, cmd.ErrOrStderr(), prompt) {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil, nil
				}
			}

			// --- Cluster configuration flags (no restart) ---

			if cmd.Flags().Changed("allowed-ip") {
				raw, _ := cmd.Flags().GetStringArray("allowed-ip")
				changes, err := util.ParseIPs(raw)
				if err != nil {
					return nil, err
				}
				cfg.AllowedIpSourceRanges = util.ApplyIPs(cfg.AllowedIpSourceRanges, changes)
			}

			if cmd.Flags().Changed("restart-mode") {
				modeStr, _ := cmd.Flags().GetString("restart-mode")
				mode, err := parseRestartMode(modeStr)
				if err != nil {
					return nil, err
				}
				cfg.RestartPolicy = mode.Enum()
			}

			if cmd.Flags().Changed("rebalance-strategy") {
				stratStr, _ := cmd.Flags().GetString("rebalance-strategy")
				strat, err := parseRebalanceStrategy(stratStr)
				if err != nil {
					return nil, err
				}
				cfg.RebalanceStrategy = strat.Enum()
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

	_ = cmd.RegisterFlagCompletionFunc("restart-mode", restartModeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("rebalance-strategy", rebalanceStrategyCompletion())
	return cmd
}

// updateDBConfigPrompt builds the confirmation message shown when database
// configuration flags are changed, warning about the rolling restart.
// It compares old (before mutation) and updated (after mutation) cluster objects
// to display a diff of each changed field.
func updateDBConfigPrompt(old, updated *clusterv1.Cluster, cmd *cobra.Command) string {
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
		lines = append(lines, fmt.Sprintf("  Replication factor:        %s", output.DiffValue(output.OptionalValue(oldRF, notSet), fmt.Sprintf("%d", newCol.GetReplicationFactor()))))
	}
	if cmd.Flags().Changed("write-consistency-factor") {
		var oldWCF *int32
		if oldCol != nil {
			oldWCF = oldCol.WriteConsistencyFactor
		}
		lines = append(lines, fmt.Sprintf("  Write consistency factor:  %s", output.DiffValue(output.OptionalValue(oldWCF, notSet), fmt.Sprintf("%d", newCol.GetWriteConsistencyFactor()))))
	}
	if cmd.Flags().Changed("async-scorer") {
		var oldAS *bool
		if oldPerf != nil {
			oldAS = oldPerf.AsyncScorer
		}
		lines = append(lines, fmt.Sprintf("  Async scorer:              %s", output.DiffValue(output.OptionalValue(oldAS, notSet), boolToYesNo(newPerf.GetAsyncScorer()))))
	}
	if cmd.Flags().Changed("optimizer-cpu-budget") {
		var oldBudget *int32
		if oldPerf != nil {
			oldBudget = oldPerf.OptimizerCpuBudget
		}
		lines = append(lines, fmt.Sprintf("  Optimizer CPU budget:      %s", output.DiffValue(output.OptionalValue(oldBudget, notSet), fmt.Sprintf("%d", newPerf.GetOptimizerCpuBudget()))))
	}

	lines = append(lines, "")
	lines = append(lines, "WARNING: Database configuration changes will result in a rolling restart of your cluster.")
	lines = append(lines, "Proceed?")
	return strings.Join(lines, "\n")
}
