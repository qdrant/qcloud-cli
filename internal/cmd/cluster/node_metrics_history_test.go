package cluster_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

var historyStart = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

func series(values ...float64) []*monitoringv1.Metric {
	metrics := make([]*monitoringv1.Metric, 0, len(values))
	for i, v := range values {
		metrics = append(metrics, &monitoringv1.Metric{Timestamp: timestamppb.New(historyStart.Add(time.Duration(i) * time.Hour)), Value: v})
	}

	return metrics
}

func usageNode(node string) *monitoringv1.ClusterNodeUsageMetrics {
	return &monitoringv1.ClusterNodeUsageMetrics{
		NodeId:       fmt.Sprintf("qdrant-%s-%s", metricsClusterID, node),
		Cpu:          series(0.4, 0.6, 0.5),
		Ram:          series(4*gib, 4.5*gib, 5*gib),
		RamCache:     series(1*gib, 1.5*gib, 2*gib),
		RamRss:       series(3*gib, 3*gib, 3*gib),
		RamQdrantRss: series(2*gib, 2.5*gib, 2.5*gib),
		Disk:         series(8*gib, 9*gib, 10*gib),
	}
}

func historyCluster(cloudProvider string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		Id:              metricsClusterID,
		CloudProviderId: cloudProvider,
		State: &clusterv1.ClusterState{
			Resources: &clusterv1.ClusterNodeResourcesSummary{
				Cpu:  &clusterv1.ClusterNodeResources{Base: 1500, Additional: 500, Reserved: 200},
				Ram:  &clusterv1.ClusterNodeResources{Base: 6, Complimentary: 2, Reserved: 0.5},
				Disk: &clusterv1.ClusterNodeResources{Base: 32},
			},
			Nodes: []*clusterv1.ClusterNodeInfo{
				{Name: fmt.Sprintf("qdrant-%s-0", metricsClusterID)},
				{Name: fmt.Sprintf("qdrant-%s-1", metricsClusterID)},
				{Name: fmt.Sprintf("qdrant-%s-2", metricsClusterID)},
			},
		},
	}
}

func setupHistory(env *testutil.TestEnv, cloudProvider string, nodes ...*monitoringv1.ClusterNodeUsageMetrics) {
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: historyCluster(cloudProvider)}, nil)
	env.MonitoringServer.GetClusterUsageMetricsCalls.Returns(&monitoringv1.GetClusterUsageMetricsResponse{Nodes: nodes}, nil)
}

type historyJSON struct {
	ClusterID string    `json:"clusterId"`
	Node      string    `json:"node"`
	NodeID    string    `json:"nodeId"`
	Since     time.Time `json:"since"`
	Until     time.Time `json:"until"`
	CPU       struct {
		ReservedVCPUs float64 `json:"reservedVcpus"`
		TotalVCPUs    float64 `json:"totalVcpus"`
		Usage         []struct {
			Timestamp time.Time `json:"timestamp"`
			Value     float64   `json:"value"`
		} `json:"usage"`
	} `json:"cpu"`
	RAM struct {
		ReservedBytes float64 `json:"reservedBytes"`
		TotalBytes    float64 `json:"totalBytes"`
		Cache         []struct {
			Value float64 `json:"value"`
		} `json:"cache"`
		Data []struct {
			Value float64 `json:"value"`
		} `json:"data"`
	} `json:"ram"`
	Disk struct {
		TotalBytes float64 `json:"totalBytes"`
		Used       []struct {
			Value float64 `json:"value"`
		} `json:"used"`
	} `json:"disk"`
}

func execHistoryJSON(t *testing.T, env *testutil.TestEnv, node string, args ...string) historyJSON {
	t.Helper()
	stdout, _, err := testutil.Exec(t, env, append([]string{"cluster", "node-metrics-history", metricsClusterID, node, "--json"}, args...)...)
	require.NoError(t, err)

	var result historyJSON
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))

	return result
}

func TestNodeMetricsHistory_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("1"), usageNode("0"))

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "1",
		"--since", "2026-10-20T00:00:00Z", "--until", "2026-10-20T02:00:00Z")
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(stdout, "Node 1  2026-10-20 00:00:00 UTC - 2026-10-20 02:00:00 UTC\n"), stdout)
	assert.NotContains(t, stdout, "Node 0")
	assert.Contains(t, stdout, "capacity 2 vCPU  system 0.2 vCPU  + usage")
	assert.Contains(t, stdout, "capacity 8.0 GiB  system 477 MiB  + cache  + data")
	assert.Contains(t, stdout, "capacity 32 GiB  used")
	// The Y axes are rounded to even steps.
	assert.Contains(t, stdout, "2.4 vCPU│")
	assert.Contains(t, stdout, "10 GiB│")
	assert.Contains(t, stdout, "40 GiB│")
	// A range of up to a day shows times only.
	assert.Contains(t, stdout, "00:00")
	assert.NotContains(t, stdout, "10/20")
	assert.NotContains(t, stdout, "\x1b[")
}

func TestNodeMetricsHistory_DefaultRange(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	before := time.Now().Truncate(time.Second)
	result := execHistoryJSON(t, env, "0")
	after := time.Now()

	req, ok := env.MonitoringServer.GetClusterUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, metricsClusterID, req.GetClusterId())
	until := req.GetUntil().AsTime()
	assert.False(t, until.Before(before) || until.After(after), "until defaults to now")
	assert.Equal(t, 7*24*time.Hour, until.Sub(req.GetSince().AsTime()))
	assert.Nil(t, req.Aggregator)

	assert.True(t, result.Until.Equal(until))
	assert.True(t, result.Since.Equal(req.GetSince().AsTime()))
}

func TestNodeMetricsHistory_UntilOnly(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	execHistoryJSON(t, env, "0", "--until", "2026-10-10T00:00:00Z")

	req, ok := env.MonitoringServer.GetClusterUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
	assert.Equal(t, time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC), req.GetUntil().AsTime())
}

func TestNodeMetricsHistory_SinceAndUntil(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	execHistoryJSON(t, env, "0", "--since", "2026-10-01", "--until", "2026-10-02")

	req, ok := env.MonitoringServer.GetClusterUsageMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), req.GetSince().AsTime())
	assert.Equal(t, time.Date(2026, 10, 2, 23, 59, 59, 0, time.UTC), req.GetUntil().AsTime())
}

func TestNodeMetricsHistory_InvalidRange(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	_, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0", "--since", "1h", "--until", "2h")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--since must be before --until")

	_, _, err = testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0", "--since", "2026-10-20", "--until", "2026-10-01")
	require.Error(t, err)

	_, _, err = testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0", "--since", "yesterday")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --since")
	assert.Equal(t, 0, env.MonitoringServer.GetClusterUsageMetricsCalls.Count())
}

func TestNodeMetricsHistory_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	result := execHistoryJSON(t, env, "0")

	assert.Equal(t, metricsClusterID, result.ClusterID)
	n := result
	assert.Equal(t, "0", n.Node)
	assert.Equal(t, fmt.Sprintf("qdrant-%s-0", metricsClusterID), n.NodeID)

	assert.InDelta(t, 0.2, n.CPU.ReservedVCPUs, 1e-9)
	assert.InDelta(t, 2, n.CPU.TotalVCPUs, 1e-9)
	require.Len(t, n.CPU.Usage, 3)
	assert.True(t, n.CPU.Usage[1].Timestamp.Equal(historyStart.Add(time.Hour)))
	assert.InDelta(t, 0.6, n.CPU.Usage[1].Value, 1e-9)

	assert.InDelta(t, 0.5e9, n.RAM.ReservedBytes, 1)
	assert.InDelta(t, 8*gib, n.RAM.TotalBytes, 1)
	require.Len(t, n.RAM.Cache, 3)
	assert.InDelta(t, 1.5*gib, n.RAM.Cache[1].Value, 1)
	require.Len(t, n.RAM.Data, 3)
	assert.InDelta(t, 2.5*gib, n.RAM.Data[1].Value, 1, "data is the RSS of the database")

	assert.InDelta(t, 32*gib, n.Disk.TotalBytes, 1)
	require.Len(t, n.Disk.Used, 3)
	assert.InDelta(t, 10*gib, n.Disk.Used[2].Value, 1)
}

func TestNodeMetricsHistory_HybridUsesNodeRSS(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, qcloudapi.HybridCloudProviderID, usageNode("0"))

	result := execHistoryJSON(t, env, "0")

	require.Len(t, result.RAM.Data, 3)
	assert.InDelta(t, 3*gib, result.RAM.Data[1].Value, 1)
}

func TestNodeMetricsHistory_DataFallsBackToRAMWithoutCache(t *testing.T) {
	env := testutil.NewTestEnv(t)
	node := usageNode("0")
	node.RamQdrantRss = nil
	setupHistory(env, "aws", node)

	result := execHistoryJSON(t, env, "0")

	require.Len(t, result.RAM.Data, 3)
	assert.InDelta(t, 3*gib, result.RAM.Data[1].Value, 1, "RAM 4.5 GiB less cache 1.5 GiB")
}

func TestNodeMetricsHistory_FallbackMatchesNearestCachePoint(t *testing.T) {
	env := testutil.NewTestEnv(t)
	node := usageNode("0")
	node.RamQdrantRss = nil
	// The cache is sampled 30 seconds after the RAM.
	for _, m := range node.RamCache {
		m.Timestamp = timestamppb.New(m.GetTimestamp().AsTime().Add(30 * time.Second))
	}

	setupHistory(env, "aws", node)

	result := execHistoryJSON(t, env, "0")

	require.Len(t, result.RAM.Data, 3)
	assert.InDelta(t, 3*gib, result.RAM.Data[0].Value, 1, "RAM 4 GiB less cache 1 GiB")
	assert.InDelta(t, 3*gib, result.RAM.Data[1].Value, 1, "RAM 4.5 GiB less cache 1.5 GiB")
	assert.InDelta(t, 3*gib, result.RAM.Data[2].Value, 1, "RAM 5 GiB less cache 2 GiB")
}

func TestNodeMetricsHistory_PointsAreSorted(t *testing.T) {
	env := testutil.NewTestEnv(t)
	node := usageNode("0")
	slices.Reverse(node.Cpu)
	setupHistory(env, "aws", node)

	result := execHistoryJSON(t, env, "0")

	require.Len(t, result.CPU.Usage, 3)
	assert.True(t, result.CPU.Usage[0].Timestamp.Equal(historyStart))
	assert.InDelta(t, 0.4, result.CPU.Usage[0].Value, 1e-9)
}

func TestNodeMetricsHistory_NodeByNumberOrID(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"), usageNode("1"), usageNode("2"))

	result := execHistoryJSON(t, env, "1")
	assert.Equal(t, "1", result.Node)
	assert.Equal(t, fmt.Sprintf("qdrant-%s-1", metricsClusterID), result.NodeID)

	result = execHistoryJSON(t, env, fmt.Sprintf("qdrant-%s-2", metricsClusterID))
	assert.Equal(t, "2", result.Node)
}

func TestNodeMetricsHistory_UnknownNode(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("1"), usageNode("0"), usageNode("3"))

	_, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "5")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `node "5" not found, available nodes: 0, 1, 2, 3`)
}

func TestNodeMetricsHistory_NoNodes(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: &clusterv1.Cluster{Id: metricsClusterID}}, nil)
	env.MonitoringServer.GetClusterUsageMetricsCalls.Returns(&monitoringv1.GetClusterUsageMetricsResponse{}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `node "0" not found: the cluster has no nodes`)
}

func TestNodeMetricsHistory_NodeWithoutMetrics(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "2")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Node 2  ")
	assert.Contains(t, stdout, "CPU\nno data")
	assert.Contains(t, stdout, "RAM\nno data")
	assert.Contains(t, stdout, "Disk\nno data")
}

func TestNodeMetricsHistory_PartialData(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", &monitoringv1.ClusterNodeUsageMetrics{
		NodeId: fmt.Sprintf("qdrant-%s-0", metricsClusterID),
		Disk:   series(8*gib, 9*gib),
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.NoError(t, err)

	assert.Contains(t, stdout, "CPU\nno data")
	assert.Contains(t, stdout, "RAM\nno data")
	assert.Contains(t, stdout, "capacity 32 GiB  used")
}

func TestNodeMetricsHistory_NodeCompletion(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws")

	stdout, _, err := testutil.Exec(t, env, "__complete", "cluster", "node-metrics-history", metricsClusterID, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0", "1", "2", ":4"}, strings.Fields(strings.Split(stdout, "Completion ended")[0]))

	req, ok := env.Server.GetClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, metricsClusterID, req.GetClusterId())
}

func TestNodeMetricsHistory_MetricsError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: historyCluster("aws")}, nil)
	env.MonitoringServer.GetClusterUsageMetricsCalls.Returns(nil, fmt.Errorf("metrics unavailable"))

	_, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cluster metrics")
}

func TestNodeMetricsHistory_ClusterError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cluster")
	assert.Equal(t, 0, env.MonitoringServer.GetClusterUsageMetricsCalls.Count())
}

func TestNodeMetricsHistory_MissingArgs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history")
	require.Error(t, err)

	_, _, err = testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID)
	require.Error(t, err)
	assert.Equal(t, 0, env.MonitoringServer.GetClusterUsageMetricsCalls.Count())
}

func gpuCluster() *clusterv1.Cluster {
	cluster := historyCluster("aws")
	cluster.State.Resources.Gpu = &clusterv1.ClusterNodeResources{Base: 1000}
	return cluster
}

func TestNodeMetricsHistory_GPU(t *testing.T) {
	env := testutil.NewTestEnv(t)
	node := usageNode("0")
	node.Gpu = series(40, 80, 60)
	node.GpuRam = series(10*gib, 12*gib, 11*gib)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: gpuCluster()}, nil)
	env.MonitoringServer.GetClusterUsageMetricsCalls.Returns(&monitoringv1.GetClusterUsageMetricsResponse{
		Nodes: []*monitoringv1.ClusterNodeUsageMetrics{node},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.NoError(t, err)

	assert.Regexp(t, "\nGPU\nusage\n *100%│", stdout, "the GPU axis ends at 100% without a capacity line")
	assert.Regexp(t, "\nGPU RAM\nused\n *16 GiB│", stdout)
	assert.Less(t, strings.Index(stdout, "\nDisk\n"), strings.Index(stdout, "\nGPU\n"))

	stdout, _, err = testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0", "--json")
	require.NoError(t, err)

	var result struct {
		GPU struct {
			UsagePercent []struct {
				Value float64 `json:"value"`
			} `json:"usagePercent"`
		} `json:"gpu"`
		GPURAM struct {
			UsedBytes []struct {
				Value float64 `json:"value"`
			} `json:"usedBytes"`
		} `json:"gpuRam"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.GPU.UsagePercent, 3)
	assert.InDelta(t, 80, result.GPU.UsagePercent[1].Value, 1e-9)
	require.Len(t, result.GPURAM.UsedBytes, 3)
	assert.InDelta(t, 12*gib, result.GPURAM.UsedBytes[1].Value, 1)
}

func TestNodeMetricsHistory_GPUClusterWithoutData(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: gpuCluster()}, nil)
	env.MonitoringServer.GetClusterUsageMetricsCalls.Returns(&monitoringv1.GetClusterUsageMetricsResponse{
		Nodes: []*monitoringv1.ClusterNodeUsageMetrics{usageNode("0")},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.NoError(t, err)

	assert.Contains(t, stdout, "\nGPU\nno data\n")
	assert.Contains(t, stdout, "\nGPU RAM\nno data\n")

	stdout, _, err = testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"usagePercent": []`)
	assert.Contains(t, stdout, `"usedBytes": []`)
}

func TestNodeMetricsHistory_GPUDataWithoutGPUResources(t *testing.T) {
	env := testutil.NewTestEnv(t)
	node := usageNode("0")
	node.Gpu = series(40, 80, 60)
	setupHistory(env, qcloudapi.HybridCloudProviderID, node)

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.NoError(t, err)

	assert.Contains(t, stdout, "\nGPU\nusage\n")
	assert.Contains(t, stdout, "\nGPU RAM\nno data\n")
}

func TestNodeMetricsHistory_NoGPU(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "GPU")

	stdout, _, err = testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0", "--json")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "gpu")
}

func TestNodeMetricsHistory_CapacitySpansRange(t *testing.T) {
	env := testutil.NewTestEnv(t)
	// The node only has data in the last hour of the default 7 days.
	recent := func(v float64) []*monitoringv1.Metric {
		return []*monitoringv1.Metric{{Timestamp: timestamppb.New(time.Now().Add(-30 * time.Minute)), Value: v}}
	}
	setupHistory(env, "aws", &monitoringv1.ClusterNodeUsageMetrics{
		NodeId: fmt.Sprintf("qdrant-%s-0", metricsClusterID),
		Cpu:    recent(0.5),
		Disk:   recent(8 * gib),
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0")
	require.NoError(t, err)

	notBraille := regexp.MustCompile("[^\u2801-\u28ff]")
	disk := stdout[strings.Index(stdout, "\nDisk\n"):]
	widest := 0
	for line := range strings.SplitSeq(disk, "\n") {
		widest = max(widest, utf8.RuneCountInString(notBraille.ReplaceAllString(line, "")))
	}

	assert.Greater(t, widest, 80, "the capacity line runs across the chart")
}

func TestNodeMetricsHistory_ChartsLineUp(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupHistory(env, "aws", usageNode("0"))

	stdout, _, err := testutil.Exec(t, env, "cluster", "node-metrics-history", metricsClusterID, "0", "--since", "2026-10-20")
	require.NoError(t, err)

	columns := map[int]bool{}
	for line := range strings.SplitSeq(stdout, "\n") {
		if i := strings.IndexAny(line, "│└"); i >= 0 {
			columns[utf8.RuneCountInString(line[:i])] = true
		}
	}

	assert.Len(t, columns, 1, "the Y axes of all charts are in the same column")
}
