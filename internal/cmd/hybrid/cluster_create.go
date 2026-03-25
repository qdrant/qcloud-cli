package hybrid

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

func newClusterCreateCommand(s *state.State) *cobra.Command {
	cmd := base.CreateCmd[*clusterv1.Cluster]{
		ValidArgsFunction: envIDCompletion(s),
		BaseCobraCommand: func() *cobra.Command {
			cmd := &cobra.Command{
				Use:   "create <env-id>",
				Short: "Create a cluster in a hybrid cloud environment",
				Long: `Create a new Qdrant cluster inside a hybrid cloud environment.

The environment ID is the target hybrid cloud environment where the cluster will run.
The cloud provider and region are fixed to "hybrid" and the environment ID respectively.`,
				Args: util.ExactArgs(1, "a hybrid cloud environment ID"),
			}
			cmd.Flags().String("name", "", "Cluster name (auto-generated if not provided)")
			cmd.Flags().Uint32("nodes", 1, "Number of nodes")
			cmd.Flags().String("version", "", "Qdrant version (default latest)")
			cmd.Flags().String("service-type", "", `Kubernetes service type ("cluster-ip", "node-port", "load-balancer")`)
			cmd.Flags().StringArray("node-selector", nil, "Node selector label ('key=value'), can be specified multiple times (max 10)")
			cmd.Flags().StringArray("annotation", nil, "Annotation ('key=value'), can be specified multiple times (max 10)")
			cmd.Flags().StringArray("pod-label", nil, "Pod label ('key=value'), can be specified multiple times (max 10)")
			cmd.Flags().StringArray("service-annotation", nil, "Service annotation ('key=value'), can be specified multiple times (max 64)")
			cmd.Flags().Uint32("reserved-cpu-percentage", 0, "Reserved CPU percentage (1-80)")
			cmd.Flags().Uint32("reserved-memory-percentage", 0, "Reserved memory percentage (1-80)")
			cmd.Flags().StringArray("toleration", nil, "Toleration ('key=value:Effect' or 'key:Exists:Effect'), can be specified multiple times")
			cmd.Flags().StringArray("label", nil, "Cluster label ('key=value'), can be specified multiple times")
			cmd.Flags().Bool("wait", false, "Wait for the cluster to become healthy")
			cmd.Flags().Duration("wait-timeout", 10*time.Minute, "Maximum time to wait for cluster health")
			cmd.Flags().Duration("wait-poll-interval", 5*time.Second, "How often to poll for cluster health")
			cmd.Flags().Uint32("replication-factor", 0, "Default replication factor for new collections")
			cmd.Flags().Int32("write-consistency-factor", 0, "Default write consistency factor for new collections")
			cmd.Flags().Bool("async-scorer", false, "Enable async scorer (uses io_uring on Linux)")
			cmd.Flags().Int32("optimizer-cpu-budget", 0, "CPU threads for optimization (0=auto, negative=subtract from available CPUs, positive=exact count)")
			cmd.Flags().StringArray("allowed-ip", nil, "Allowed client IP CIDR ranges; max 20")
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
				CloudProviderId:       "hybrid",
				CloudProviderRegionId: envID,
				Configuration: &clusterv1.ClusterConfiguration{
					NumberOfNodes: nodes,
				},
			}

			if version != "" {
				cluster.Configuration.Version = &version
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
				if len(raw) > 10 {
					return nil, fmt.Errorf("--node-selector: maximum 10 entries allowed, got %d", len(raw))
				}
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
				if len(raw) > 10 {
					return nil, fmt.Errorf("--annotation: maximum 10 entries allowed, got %d", len(raw))
				}
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
				if len(raw) > 10 {
					return nil, fmt.Errorf("--pod-label: maximum 10 entries allowed, got %d", len(raw))
				}
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
				if len(raw) > 64 {
					return nil, fmt.Errorf("--service-annotation: maximum 64 entries allowed, got %d", len(raw))
				}
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

			// Database configuration
			if cmd.Flags().Changed("replication-factor") || cmd.Flags().Changed("write-consistency-factor") ||
				cmd.Flags().Changed("async-scorer") || cmd.Flags().Changed("optimizer-cpu-budget") {
				if cluster.Configuration.DatabaseConfiguration == nil {
					cluster.Configuration.DatabaseConfiguration = &clusterv1.DatabaseConfiguration{}
				}
				dbCfg := cluster.Configuration.DatabaseConfiguration

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
			return qcloudapi.WaitForClusterHealthy(ctx, client.Cluster(), cmd.ErrOrStderr(), accountID, created.GetId(), waitTimeout, pollInterval)
		},
		PrintResource: func(_ *cobra.Command, out io.Writer, created *clusterv1.Cluster) {
			fmt.Fprintf(out, "Cluster %s (%s) created.\n", created.GetId(), created.GetName())
		},
	}.CobraCommand(s)
	_ = cmd.RegisterFlagCompletionFunc("service-type", serviceTypeCompletion())
	_ = cmd.RegisterFlagCompletionFunc("version", completion.VersionCompletion(s))
	return cmd
}
