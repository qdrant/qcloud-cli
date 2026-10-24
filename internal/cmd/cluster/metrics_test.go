package cluster_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	monitoringv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/monitoring/v1"

	"github.com/qdrant/qcloud-cli/internal/qcloudapi"
	"github.com/qdrant/qcloud-cli/internal/testutil"
)

const (
	metricsClusterID = "7b2ea926-724b-4de2-b73a-8675c42a6ebe"
	gib              = 1 << 30
)

func overview(avg, total float64, unit string) *monitoringv1.ClusterMetricOverview {
	return &monitoringv1.ClusterMetricOverview{
		Avg:   []*monitoringv1.IntervalAverage{{Value: avg}},
		Total: &monitoringv1.ResourceValue{Value: total, Unit: unit},
	}
}

func nodeMetrics(node string, cpu, ramQdrant, ramRss, cache, disk float64) *monitoringv1.ClusterNodeMetrics {
	return &monitoringv1.ClusterNodeMetrics{
		NodeId:       fmt.Sprintf("qdrant-%s-%s", metricsClusterID, node),
		Cpu:          overview(cpu, 2, "vcpu"),
		Ram:          overview(ramRss+cache, 8*gib, "bytes"),
		RamCache:     overview(cache, 8*gib, "bytes"),
		RamRss:       overview(ramRss, 8*gib, "bytes"),
		RamQdrantRss: overview(ramQdrant, 8*gib, "bytes"),
		Disk:         overview(disk, 32*gib, "bytes"),
	}
}

func metricsCluster(cloudProvider string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		Id:              metricsClusterID,
		CloudProviderId: cloudProvider,
		State: &clusterv1.ClusterState{
			Resources: &clusterv1.ClusterNodeResourcesSummary{
				Cpu:  &clusterv1.ClusterNodeResources{Base: 2000, Reserved: 200},
				Ram:  &clusterv1.ClusterNodeResources{Base: 8, Reserved: 0.5},
				Disk: &clusterv1.ClusterNodeResources{Base: 32},
			},
			Nodes: []*clusterv1.ClusterNodeInfo{
				{
					Name:             fmt.Sprintf("qdrant-%s-0", metricsClusterID),
					State:            clusterv1.ClusterNodeState_CLUSTER_NODE_STATE_HEALTHY,
					Version:          "v1.16.3",
					AvailabilityZone: new("us-east-1a"),
				},
				{
					Name:    fmt.Sprintf("qdrant-%s-1", metricsClusterID),
					State:   clusterv1.ClusterNodeState_CLUSTER_NODE_STATE_RECOVERING,
					Version: "v1.16.3",
				},
			},
		},
	}
}

func setupMetrics(env *testutil.TestEnv, cloudProvider string, nodes ...*monitoringv1.ClusterNodeMetrics) {
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: metricsCluster(cloudProvider)}, nil)
	env.MonitoringServer.GetClusterSummaryMetricsCalls.Returns(&monitoringv1.GetClusterSummaryMetricsResponse{Nodes: nodes}, nil)
}

func TestClusterMetrics_TextOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, "aws",
		nodeMetrics("1", 0.3, 3*gib, 4*gib, 1*gib, 4*gib),
		nodeMetrics("0", 0.5, 2*gib, 3*gib, 1.5*gib, 8*gib),
	)

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "Node 0  HEALTHY  v1.16.3  us-east-1a")
	assert.Contains(t, stdout, "Node 1  RECOVERING  v1.16.3")
	assert.Less(t, strings.Index(stdout, "Node 0"), strings.Index(stdout, "Node 1"), "nodes are sorted by number")

	// CPU: 0.2 reserved + 0.5 used of 2 vCPU.
	assert.Contains(t, stdout, " 35.0%  0.7 of 2 vCPU")
	assert.Contains(t, stdout, "\n        system 0.2  usage 0.5  available 1.3\n")
	// Small usage keeps two significant digits instead of rounding to 0.
	assert.Contains(t, stdout, "\n        system 0.2  usage 0.3  available 1.5\n")
	// RAM: 0.5 GB (477 MiB) reserved + 1.5 GiB cache + 2 GiB data of 8 GiB.
	assert.Contains(t, stdout, " 49.6%  4.0 GiB of 8.0 GiB")
	assert.Contains(t, stdout, "system 477 MiB  cache 1.5 GiB  data 2.0 GiB  available 4.0 GiB")
	// Disk: 8 GiB of 32 GiB.
	assert.Contains(t, stdout, " 25.0%  8.0 GiB of 32 GiB")
	assert.Contains(t, stdout, "used 8.0 GiB  available 24 GiB")
	// Bars are bracketed without color.
	assert.Contains(t, stdout, "CPU   [")
	assert.NotContains(t, stdout, "\x1b[")
}

func TestClusterMetrics_SmallCPUUsage(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, "aws", nodeMetrics("0", 0.0044613, 2*gib, 3*gib, 1.5*gib, 8*gib))

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "  usage 0.0045  ")
	assert.Contains(t, stdout, "0.2 of 2 vCPU")
}

func TestClusterMetrics_HybridUsesNodeRSS(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, qcloudapi.HybridCloudProviderID, nodeMetrics("0", 0.5, 2*gib, 3*gib, 1.5*gib, 8*gib))

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "data 3.0 GiB")
	assert.Contains(t, stdout, " 62.1%  5.0 GiB of 8.0 GiB")
}

func TestClusterMetrics_NoData(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, "aws", &monitoringv1.ClusterNodeMetrics{
		NodeId: fmt.Sprintf("qdrant-%s-0", metricsClusterID),
		Cpu:    &monitoringv1.ClusterMetricOverview{Total: &monitoringv1.ResourceValue{Value: 2, Unit: "vcpu"}},
		Disk:   overview(1*gib, 32*gib, "bytes"),
	})

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "CPU   no data")
	assert.Contains(t, stdout, "RAM   no data")
	assert.Contains(t, stdout, "1.0 GiB of 32 GiB")
}

func TestClusterMetrics_NoNodes(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, "aws")

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.NoError(t, err)

	assert.Contains(t, stdout, "No metrics available for this cluster.")
}

func TestClusterMetrics_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, "aws", nodeMetrics("0", 0.5, 2*gib, 3*gib, 1.5*gib, 8*gib))

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID, "--json")
	require.NoError(t, err)

	var result struct {
		ClusterID string `json:"clusterId"`
		Nodes     []struct {
			Node             string `json:"node"`
			NodeID           string `json:"nodeId"`
			State            string `json:"state"`
			AvailabilityZone string `json:"availabilityZone"`
			CPU              struct {
				UsedVCPUs      float64 `json:"usedVcpus"`
				ReservedVCPUs  float64 `json:"reservedVcpus"`
				AvailableVCPUs float64 `json:"availableVcpus"`
				TotalVCPUs     float64 `json:"totalVcpus"`
				UsedPercent    float64 `json:"usedPercent"`
			} `json:"cpu"`
			RAM struct {
				DataBytes      float64 `json:"dataBytes"`
				CacheBytes     float64 `json:"cacheBytes"`
				ReservedBytes  float64 `json:"reservedBytes"`
				AvailableBytes float64 `json:"availableBytes"`
				TotalBytes     float64 `json:"totalBytes"`
				UsedPercent    float64 `json:"usedPercent"`
			} `json:"ram"`
			Disk struct {
				UsedBytes      float64 `json:"usedBytes"`
				AvailableBytes float64 `json:"availableBytes"`
				TotalBytes     float64 `json:"totalBytes"`
				UsedPercent    float64 `json:"usedPercent"`
			} `json:"disk"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))

	assert.Equal(t, metricsClusterID, result.ClusterID)
	require.Len(t, result.Nodes, 1)
	n := result.Nodes[0]
	assert.Equal(t, "0", n.Node)
	assert.Equal(t, fmt.Sprintf("qdrant-%s-0", metricsClusterID), n.NodeID)
	assert.Equal(t, "HEALTHY", n.State)
	assert.Equal(t, "us-east-1a", n.AvailabilityZone)
	assert.InDelta(t, 0.5, n.CPU.UsedVCPUs, 1e-9)
	assert.InDelta(t, 0.2, n.CPU.ReservedVCPUs, 1e-9)
	assert.InDelta(t, 1.3, n.CPU.AvailableVCPUs, 1e-9)
	assert.InDelta(t, 2, n.CPU.TotalVCPUs, 1e-9)
	assert.InDelta(t, 35, n.CPU.UsedPercent, 1e-9)
	assert.InDelta(t, 2*gib, n.RAM.DataBytes, 1)
	assert.InDelta(t, 1.5*gib, n.RAM.CacheBytes, 1)
	assert.InDelta(t, 0.5e9, n.RAM.ReservedBytes, 1)
	assert.InDelta(t, 4.5*gib-0.5e9, n.RAM.AvailableBytes, 1)
	assert.InDelta(t, 8*gib, n.RAM.TotalBytes, 1)
	assert.InDelta(t, 49.6, n.RAM.UsedPercent, 1e-9)
	assert.InDelta(t, 8*gib, n.Disk.UsedBytes, 1)
	assert.InDelta(t, 24*gib, n.Disk.AvailableBytes, 1)
	assert.InDelta(t, 32*gib, n.Disk.TotalBytes, 1)
	assert.InDelta(t, 25, n.Disk.UsedPercent, 1e-9)
}

func TestClusterMetrics_AvailableNeverNegative(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, "aws", nodeMetrics("0", 1.9, 7.8*gib, 7.8*gib, 0, 33*gib))

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID, "--json")
	require.NoError(t, err)

	var result struct {
		Nodes []struct {
			CPU struct {
				AvailableVCPUs float64 `json:"availableVcpus"`
			} `json:"cpu"`
			RAM struct {
				AvailableBytes float64 `json:"availableBytes"`
			} `json:"ram"`
			Disk struct {
				AvailableBytes float64 `json:"availableBytes"`
			} `json:"disk"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	require.Len(t, result.Nodes, 1)
	assert.Zero(t, result.Nodes[0].CPU.AvailableVCPUs)
	assert.Zero(t, result.Nodes[0].RAM.AvailableBytes)
	assert.Zero(t, result.Nodes[0].Disk.AvailableBytes)
}

func TestClusterMetrics_Requests(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupMetrics(env, "aws", nodeMetrics("0", 0.5, 2*gib, 3*gib, 1.5*gib, 8*gib))

	_, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.NoError(t, err)

	clusterReq, ok := env.Server.GetClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", clusterReq.GetAccountId())
	assert.Equal(t, metricsClusterID, clusterReq.GetClusterId())

	metricsReq, ok := env.MonitoringServer.GetClusterSummaryMetricsCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", metricsReq.GetAccountId())
	assert.Equal(t, metricsClusterID, metricsReq.GetClusterId())
}

func TestClusterMetrics_MetricsError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{Cluster: metricsCluster("aws")}, nil)
	env.MonitoringServer.GetClusterSummaryMetricsCalls.Returns(nil, fmt.Errorf("metrics unavailable"))

	_, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cluster metrics")
}

func TestClusterMetrics_ClusterError(t *testing.T) {
	env := testutil.NewTestEnv(t)
	env.Server.GetClusterCalls.Returns(nil, fmt.Errorf("not found"))

	_, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get cluster")
	assert.Equal(t, 0, env.MonitoringServer.GetClusterSummaryMetricsCalls.Count())
}

func TestClusterMetrics_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "metrics")
	require.Error(t, err)
}

func TestClusterMetrics_FitsIn80Columns(t *testing.T) {
	env := testutil.NewTestEnv(t)
	node := nodeMetrics("0", 1.25, 12*gib, 12*gib, 9*gib, 100*gib)
	node.Cpu = overview(1.25, 16, "vcpu")
	node.RamCache = overview(9*gib, 128*gib, "bytes")
	node.RamQdrantRss = overview(12*gib, 128*gib, "bytes")
	node.Disk = overview(100*gib, 1024*gib, "bytes")
	setupMetrics(env, "aws", node)

	stdout, _, err := testutil.Exec(t, env, "cluster", "metrics", metricsClusterID)
	require.NoError(t, err)

	for line := range strings.SplitSeq(stdout, "\n") {
		assert.LessOrEqual(t, utf8.RuneCountInString(line), 80, line)
	}

	assert.Contains(t, stdout, "\n        system 477 MiB  cache 9.0 GiB  data 12 GiB  available 107 GiB\n")
}
