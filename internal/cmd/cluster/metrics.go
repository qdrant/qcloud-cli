package cluster

import (
	"cmp"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/completion"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/cmd/util"
	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/state"
)

const metricsBarWidth = 30

// The reserved RAM of a cluster is given in decimal gigabytes, unlike the
// configured RAM, which is in GiB. This matches the Qdrant Cloud console.
const bytesPerGB = 1000 * 1000 * 1000

type clusterMetrics struct {
	ClusterID string        `json:"clusterId"`
	Nodes     []nodeMetrics `json:"nodes"`
}

type nodeMetrics struct {
	Node             string      `json:"node"`
	NodeID           string      `json:"nodeId"`
	State            string      `json:"state,omitempty"`
	Version          string      `json:"version,omitempty"`
	AvailabilityZone string      `json:"availabilityZone,omitempty"`
	CPU              cpuUsage    `json:"cpu"`
	RAM              memoryUsage `json:"ram"`
	Disk             diskUsage   `json:"disk"`
}

type cpuUsage struct {
	UsedVCPUs      float64 `json:"usedVcpus"`
	ReservedVCPUs  float64 `json:"reservedVcpus"`
	AvailableVCPUs float64 `json:"availableVcpus"`
	TotalVCPUs     float64 `json:"totalVcpus"`
	UsedPercent    float64 `json:"usedPercent"`
}

type memoryUsage struct {
	DataBytes      float64 `json:"dataBytes"`
	CacheBytes     float64 `json:"cacheBytes"`
	ReservedBytes  float64 `json:"reservedBytes"`
	AvailableBytes float64 `json:"availableBytes"`
	TotalBytes     float64 `json:"totalBytes"`
	UsedPercent    float64 `json:"usedPercent"`
}

type diskUsage struct {
	UsedBytes      float64 `json:"usedBytes"`
	AvailableBytes float64 `json:"availableBytes"`
	TotalBytes     float64 `json:"totalBytes"`
	UsedPercent    float64 `json:"usedPercent"`
}

func newMetricsCommand(s *state.State) *cobra.Command {
	return base.DescribeCmd[*clusterMetrics]{
		Use:   "metrics <cluster-id>",
		Short: "Show the current CPU, RAM and disk usage of each cluster node",
		Long: `Show the current CPU, RAM and disk usage of each node in a cluster.

Each resource is drawn as a bar relative to the capacity of the node, followed
by the usage in percent and in absolute values. CPU and RAM include the
resources reserved for the system, which the database cannot use. RAM is further
split into the page cache and the memory used by the database itself.

The values are short-term averages reported by the cluster's monitoring, so a
node that has just started may show no data yet.`,
		Example: `# Show resource usage for every node of a cluster
qcloud cluster metrics 7b2ea926-724b-4de2-b73a-8675c42a6ebe

# Output the usage as JSON
qcloud cluster metrics 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json

# List the RAM usage in percent per node
qcloud cluster metrics 7b2ea926-724b-4de2-b73a-8675c42a6ebe --json | jq -r '.nodes[] | "\(.node) \(.ram.usedPercent)"'`,
		Args: util.ExactArgs(1, "a cluster ID"),
		Fetch: func(s *state.State, cmd *cobra.Command, args []string) (*clusterMetrics, error) {
			ctx := cmd.Context()
			client, err := s.Client(ctx)
			if err != nil {
				return nil, err
			}

			accountID, err := s.AccountID()
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

			metricsResp, err := client.Monitoring().GetClusterSummaryMetrics(ctx, &monitoringv1.GetClusterSummaryMetricsRequest{
				AccountId: accountID,
				ClusterId: args[0],
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get cluster metrics: %w", err)
			}

			return buildClusterMetrics(clusterResp.GetCluster(), metricsResp.GetNodes()), nil
		},
		PrintText: func(_ *cobra.Command, w io.Writer, m *clusterMetrics) error {
			printClusterMetrics(w, m, output.ColorEnabled(w))
			return nil
		},
		ValidArgsFunction: completion.ClusterIDCompletion(s),
	}.CobraCommand(s)
}

func buildClusterMetrics(cluster *clusterv1.Cluster, nodes []*monitoringv1.ClusterNodeMetrics) *clusterMetrics {
	resources := cluster.GetState().GetResources()
	reservedVCPUs := resources.GetCpu().GetReserved() / 1000
	reservedRAM := resources.GetRam().GetReserved() * bytesPerGB
	isHybrid := cluster.GetCloudProviderId() == qcloudapi.HybridCloudProviderID

	result := &clusterMetrics{ClusterID: cluster.GetId(), Nodes: make([]nodeMetrics, 0, len(nodes))}
	for _, n := range nodes {
		nm := nodeMetrics{
			Node:   nodeNumber(cluster.GetId(), n.GetNodeId()),
			NodeID: n.GetNodeId(),
		}

		for _, info := range cluster.GetState().GetNodes() {
			if info.GetName() == n.GetNodeId() {
				nm.State = output.ClusterNodeState(info.GetState())
				nm.Version = info.GetVersion()
				nm.AvailabilityZone = info.GetAvailabilityZone()
			}
		}

		cpuUsed, cpuTotal := metricOverview(n.GetCpu())
		nm.CPU = cpuUsage{UsedVCPUs: cpuUsed, TotalVCPUs: cpuTotal}
		if cpuTotal > 0 {
			nm.CPU.ReservedVCPUs = reservedVCPUs
			nm.CPU.AvailableVCPUs = available(cpuUsed+reservedVCPUs, cpuTotal)
			nm.CPU.UsedPercent = percent(cpuUsed+reservedVCPUs, cpuTotal)
		}

		// Like the Qdrant Cloud console, use the RSS of the database process,
		// or of the whole node on Hybrid Cloud.
		ramMetric := n.GetRamQdrantRss()
		if isHybrid {
			ramMetric = n.GetRamRss()
		}

		ramData, ramTotal := metricOverview(ramMetric)
		ramCache, _ := metricOverview(n.GetRamCache())
		nm.RAM = memoryUsage{DataBytes: ramData, CacheBytes: ramCache, TotalBytes: ramTotal}
		if ramTotal > 0 {
			nm.RAM.ReservedBytes = reservedRAM
			nm.RAM.AvailableBytes = available(ramData+ramCache+reservedRAM, ramTotal)
			nm.RAM.UsedPercent = percent(ramData+ramCache+reservedRAM, ramTotal)
		}

		diskUsed, diskTotal := metricOverview(n.GetDisk())
		nm.Disk = diskUsage{
			UsedBytes:      diskUsed,
			AvailableBytes: available(diskUsed, diskTotal),
			TotalBytes:     diskTotal,
			UsedPercent:    percent(diskUsed, diskTotal),
		}

		result.Nodes = append(result.Nodes, nm)
	}

	slices.SortStableFunc(result.Nodes, func(a, b nodeMetrics) int { return compareNodes(a.Node, b.Node) })

	return result
}

// nodeNumber returns the short name of a node, such as "0", from its full ID.
func nodeNumber(clusterID, nodeID string) string {
	return strings.TrimPrefix(nodeID, "qdrant-"+clusterID+"-")
}

// compareNodes orders node numbers numerically, falling back to text.
func compareNodes(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	if aErr == nil && bErr == nil {
		return cmp.Compare(an, bn)
	}

	return cmp.Compare(a, b)
}

// metricOverview returns the first interval average and the total capacity of
// a metric, like the Qdrant Cloud console. Both are 0 when the node has not
// reported the metric yet.
func metricOverview(m *monitoringv1.ClusterMetricOverview) (float64, float64) {
	if len(m.GetAvg()) == 0 || m.GetTotal().GetValue() == 0 {
		return 0, 0
	}

	return m.GetAvg()[0].GetValue(), m.GetTotal().GetValue()
}

// available returns the capacity left on top of used, which is never negative.
func available(used, total float64) float64 {
	return max(total-used, 0)
}

func percent(v, total float64) float64 {
	if total == 0 {
		return 0
	}

	return math.Round(v/total*1000) / 10
}

type metricsRow struct {
	label    string
	segments []output.BarSegment
	total    float64
	percent  float64
	amount   string
	details  []string
}

func printClusterMetrics(w io.Writer, m *clusterMetrics, color bool) {
	if len(m.Nodes) == 0 {
		fmt.Fprintln(w, "No metrics available for this cluster.")
		return
	}

	for i, n := range m.Nodes {
		if i > 0 {
			fmt.Fprintln(w)
		}

		header := []string{"Node " + n.Node}
		for _, v := range []string{n.State, n.Version, n.AvailabilityZone} {
			if v != "" {
				header = append(header, v)
			}
		}

		fmt.Fprintln(w, strings.Join(header, "  "))

		for _, r := range metricsRows(n, color) {
			if r.total == 0 {
				fmt.Fprintf(w, "  %-4s  no data\n", r.label)
				continue
			}

			// The breakdown goes on its own line, indented to the bar, so the
			// rows fit into a terminal of 80 columns.
			fmt.Fprintf(w, "  %-4s  %s  %5.1f%%  %s\n", r.label, output.UsageBar(r.segments, r.total, metricsBarWidth, color), r.percent, r.amount)
			fmt.Fprintf(w, "        %s\n", strings.Join(r.details, "  "))
		}
	}
}

func metricsRows(n nodeMetrics, color bool) []metricsRow {
	cpu, ram, disk := n.CPU, n.RAM, n.Disk
	return []metricsRow{
		{
			label: "CPU",
			segments: []output.BarSegment{
				{Value: cpu.ReservedVCPUs, Color: output.BarColorSystem},
				{Value: cpu.UsedVCPUs, Color: output.BarColorData},
			},
			total:   cpu.TotalVCPUs,
			percent: cpu.UsedPercent,
			amount:  fmt.Sprintf("%s of %s vCPU", formatVCPUs(cpu.ReservedVCPUs+cpu.UsedVCPUs), formatVCPUs(cpu.TotalVCPUs)),
			details: []string{
				output.BarLegend("system "+formatVCPUs(cpu.ReservedVCPUs), output.BarColorSystem, color),
				output.BarLegend("usage "+formatVCPUs(cpu.UsedVCPUs), output.BarColorData, color),
				output.BarLegend("available "+formatVCPUs(cpu.AvailableVCPUs), output.BarColorFree, color),
			},
		},
		{
			label: "RAM",
			segments: []output.BarSegment{
				{Value: ram.ReservedBytes, Color: output.BarColorSystem},
				{Value: ram.CacheBytes, Color: output.BarColorCache},
				{Value: ram.DataBytes, Color: output.BarColorData},
			},
			total:   ram.TotalBytes,
			percent: ram.UsedPercent,
			amount:  fmt.Sprintf("%s of %s", formatBytes(ram.ReservedBytes+ram.CacheBytes+ram.DataBytes), formatBytes(ram.TotalBytes)),
			details: []string{
				output.BarLegend("system "+formatBytes(ram.ReservedBytes), output.BarColorSystem, color),
				output.BarLegend("cache "+formatBytes(ram.CacheBytes), output.BarColorCache, color),
				output.BarLegend("data "+formatBytes(ram.DataBytes), output.BarColorData, color),
				output.BarLegend("available "+formatBytes(ram.AvailableBytes), output.BarColorFree, color),
			},
		},
		{
			label: "Disk",
			segments: []output.BarSegment{
				{Value: disk.UsedBytes, Color: output.BarColorData},
			},
			total:   disk.TotalBytes,
			percent: disk.UsedPercent,
			amount:  fmt.Sprintf("%s of %s", formatBytes(disk.UsedBytes), formatBytes(disk.TotalBytes)),
			details: []string{
				output.BarLegend("used "+formatBytes(disk.UsedBytes), output.BarColorData, color),
				output.BarLegend("available "+formatBytes(disk.AvailableBytes), output.BarColorFree, color),
			},
		},
	}
}

// formatVCPUs formats a number of vCPUs with two decimals, or two significant
// digits below 0.1, so the small usage of an idle node does not show as 0.
func formatVCPUs(v float64) string {
	if math.Abs(v) < 0.1 {
		return output.Significant(v, 2)
	}

	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

func formatBytes(v float64) string {
	return output.Bytes(uint64(max(v, 0)))
}
