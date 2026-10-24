package cluster

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

const bytesPerGiB = 1 << 30

type nodeMetricsHistory struct {
	ClusterID string      `json:"clusterId"`
	Node      string      `json:"node"`
	NodeID    string      `json:"nodeId"`
	Since     time.Time   `json:"since"`
	Until     time.Time   `json:"until"`
	CPU       cpuHistory  `json:"cpu"`
	RAM       ramHistory  `json:"ram"`
	Disk      diskHistory `json:"disk"`
}

type cpuHistory struct {
	ReservedVCPUs float64        `json:"reservedVcpus"`
	TotalVCPUs    float64        `json:"totalVcpus"`
	Usage         []historyPoint `json:"usage"`
}

type ramHistory struct {
	ReservedBytes float64        `json:"reservedBytes"`
	TotalBytes    float64        `json:"totalBytes"`
	Cache         []historyPoint `json:"cache"`
	Data          []historyPoint `json:"data"`
}

type diskHistory struct {
	TotalBytes float64        `json:"totalBytes"`
	Used       []historyPoint `json:"used"`
}

func newNodeMetricsHistoryCommand(s *state.State) *cobra.Command {
	cmd := base.DescribeCmd[*nodeMetricsHistory]{
		Use:   "node-metrics-history <cluster-id> <node>",
		Short: "Show the CPU, RAM and disk usage of a cluster node over time",
		Long: `Show the CPU, RAM and disk usage of a cluster node over time.

The node is given by its number, such as 0, as shown by "qcloud cluster metrics",
or by its full node ID. Each resource is drawn as a stacked area chart, with a
red line marking the capacity of the node: CPU shows the resources reserved for
the system with the usage of the database on top, and RAM shows the reserved
memory, then the page cache, then the memory used by the database. Without
color, the areas are drawn as stacked lines instead.

By default, the usage over the last 7 days is shown. --since and --until accept
an RFC3339 timestamp, a YYYY-MM-DD date in UTC, or a duration ago such as 6h or
7d. When only --until is given, the range covers the 7 days before it. Times on
the charts are in UTC.`,
		Example: `# Show the usage of node 0 over the last 7 days
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0

# Show the usage of node 1 over the last 24 hours
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 1 --since 24h

# Show the usage in a specific time range
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0 --since 2026-10-01 --until 2026-10-07

# Output the time series as JSON
qcloud cluster node-metrics-history 7b2ea926-724b-4de2-b73a-8675c42a6ebe 0 --json`,
		Args: util.ExactArgs(2, "a cluster ID and a node"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*nodeMetricsHistory, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, err
			}

			since, until, err := readMetricsHistoryRange(cmd, time.Now().Truncate(time.Second))
			if err != nil {
				return nil, err
			}

			clusterResp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
				AccountId: accountID,
				ClusterId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster: %w", err)
			}

			metricsResp, err := client.Monitoring().GetClusterUsageMetrics(ctx, &monitoringv1.GetClusterUsageMetricsRequest{
				AccountId: accountID,
				ClusterId: args[0],
				Since:     timestamppb.New(since),
				Until:     timestamppb.New(until),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster metrics: %w", err)
			}

			return buildNodeMetricsHistory(clusterResp.GetCluster(), metricsResp.GetNodes(), args[1], since, until)
		},
		PrintText: func(_ *cobra.Command, w io.Writer, h *nodeMetricsHistory) error {
			printNodeMetricsHistory(w, h, metricsChartWidth(w), output.ColorEnabled(w))
			return nil
		},
		ValidArgsFunction: clusterNodeCompletion(s),
	}.CobraCommand(s)

	addMetricsHistoryRangeFlags(cmd)

	return cmd
}

// buildNodeMetricsHistory returns the usage of node, given by number or full
// ID. A node of the cluster without metrics in the time range has empty series.
func buildNodeMetricsHistory(cluster *clusterv1.Cluster, nodes []*monitoringv1.ClusterNodeUsageMetrics, node string, since, until time.Time) (*nodeMetricsHistory, error) {
	var usage *monitoringv1.ClusterNodeUsageMetrics
	var known []string
	matches := func(nodeID string) bool {
		known = append(known, nodeNumber(cluster.GetId(), nodeID))
		return nodeID == node || nodeNumber(cluster.GetId(), nodeID) == node
	}

	for _, n := range nodes {
		if usage == nil && matches(n.GetNodeId()) {
			usage = n
		}
	}

	for _, info := range cluster.GetState().GetNodes() {
		if usage == nil && matches(info.GetName()) {
			usage = &monitoringv1.ClusterNodeUsageMetrics{NodeId: info.GetName()}
		}
	}

	if usage == nil {
		slices.SortFunc(known, compareNodes)
		known = slices.Compact(known)
		if len(known) == 0 {
			return nil, fmt.Errorf("node %q not found: the cluster has no nodes", node)
		}

		return nil, fmt.Errorf("node %q not found, available nodes: %s", node, strings.Join(known, ", "))
	}

	resources := cluster.GetState().GetResources()
	cache := historyPoints(usage.GetRamCache())

	// Hybrid Cloud clusters do not report the RSS of the database process
	// separately, so use the RSS of the whole node there.
	ramMetric := usage.GetRamQdrantRss()
	if cluster.GetCloudProviderId() == qcloudapi.HybridCloudProviderID {
		ramMetric = usage.GetRamRss()
	}

	data := historyPoints(ramMetric)
	if len(data) == 0 {
		// Without an RSS series, derive the data from the RAM including the
		// cache, like the Qdrant Cloud console does.
		cacheAt := pointsByTime(cache)
		for _, p := range historyPoints(usage.GetRam()) {
			data = append(data, historyPoint{Timestamp: p.Timestamp, Value: max(p.Value-cacheAt[p.Timestamp.Unix()], 0)})
		}
	}

	return &nodeMetricsHistory{
		ClusterID: cluster.GetId(),
		Node:      nodeNumber(cluster.GetId(), usage.GetNodeId()),
		NodeID:    usage.GetNodeId(),
		Since:     since.UTC(),
		Until:     until.UTC(),
		CPU: cpuHistory{
			ReservedVCPUs: resources.GetCpu().GetReserved() / 1000,
			TotalVCPUs:    resourceTotal(resources.GetCpu()) / 1000,
			Usage:         historyPoints(usage.GetCpu()),
		},
		RAM: ramHistory{
			ReservedBytes: resources.GetRam().GetReserved() * bytesPerGB,
			TotalBytes:    resourceTotal(resources.GetRam()) * bytesPerGiB,
			Cache:         cache,
			Data:          data,
		},
		Disk: diskHistory{
			TotalBytes: resourceTotal(resources.GetDisk()) * bytesPerGiB,
			Used:       historyPoints(usage.GetDisk()),
		},
	}, nil
}

// clusterNodeCompletion completes a cluster ID, then the numbers of its nodes.
func clusterNodeCompletion(s *state.State) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	clusterIDs := completion.ClusterIDCompletion(s)
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return clusterIDs(cmd, args, toComplete)
		case 1:
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			accountID, err := s.AccountID()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			resp, err := client.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{AccountId: accountID, ClusterId: args[0]})
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			nodes := make([]string, 0, len(resp.GetCluster().GetState().GetNodes()))
			for _, n := range resp.GetCluster().GetState().GetNodes() {
				nodes = append(nodes, nodeNumber(args[0], n.GetName()))
			}

			slices.SortFunc(nodes, compareNodes)

			return nodes, cobra.ShellCompDirectiveNoFileComp
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}

// resourceTotal returns the capacity of a node resource, which is the sum of
// its base, complimentary and additional parts.
func resourceTotal(r *clusterv1.ClusterNodeResources) float64 {
	return r.GetBase() + r.GetComplimentary() + r.GetAdditional()
}

func printNodeMetricsHistory(w io.Writer, h *nodeMetricsHistory, width int, color bool) {
	fmt.Fprintf(w, "Node %s  %s - %s\n", h.Node, output.FullDateTime(h.Since), output.FullDateTime(h.Until))

	for _, c := range nodeMetricsHistoryCharts(*h) {
		printMetricsChart(w, c, h.Since, h.Until, width, color)
	}
}

func nodeMetricsHistoryCharts(n nodeMetricsHistory) []metricsChart {
	cpu, ram, disk := n.CPU, n.RAM, n.Disk
	vcpus := func(v float64) string { return formatVCPUs(v) + " vCPU" }

	cpuSeries := []output.ChartSeries{}
	if cpu.ReservedVCPUs > 0 {
		cpuSeries = append(cpuSeries, output.ChartSeries{
			Name: "system " + vcpus(cpu.ReservedVCPUs), Color: output.BarColorSystem, Points: constant(cpu.Usage, cpu.ReservedVCPUs), Fill: true,
		})
	}

	cpuSeries = append(cpuSeries, output.ChartSeries{
		Name: "+ usage", Color: output.BarColorData, Points: stack(cpu.Usage, cpu.ReservedVCPUs), Fill: true,
	})

	ramSeries := []output.ChartSeries{}
	if ram.ReservedBytes > 0 {
		ramSeries = append(ramSeries, output.ChartSeries{
			Name: "system " + formatBytes(ram.ReservedBytes), Color: output.BarColorSystem, Points: constant(ram.Data, ram.ReservedBytes), Fill: true,
		})
	}

	ramSeries = append(ramSeries,
		output.ChartSeries{Name: "+ cache", Color: output.BarColorCache, Points: stack(ram.Cache, ram.ReservedBytes), Fill: true},
		output.ChartSeries{Name: "+ data", Color: output.BarColorData, Points: stack(ram.Data, ram.ReservedBytes, ram.Cache), Fill: true},
	)

	return []metricsChart{
		{title: "CPU", base: cpu.Usage, total: cpu.TotalVCPUs, series: cpuSeries, yLabel: vcpus},
		{title: "RAM", base: ram.Data, total: ram.TotalBytes, series: ramSeries, yLabel: formatBytes, yUnit: bytesPerGiB},
		{
			title:  "Disk",
			base:   disk.Used,
			total:  disk.TotalBytes,
			series: []output.ChartSeries{{Name: "used", Color: output.BarColorData, Points: stack(disk.Used, 0), Fill: true}},
			yLabel: formatBytes,
			yUnit:  bytesPerGiB,
		},
	}
}
