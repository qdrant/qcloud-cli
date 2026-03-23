package cluster_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestDescribeCluster_BasicInfo(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:                    "cluster-abc",
			Name:                  "my-cluster",
			CloudProviderId:       "aws",
			CloudProviderRegionId: "us-east-1",
			State: &clusterv1.ClusterState{
				Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
			},
			Configuration: &clusterv1.ClusterConfiguration{
				Version:       new("1.13.0"),
				NumberOfNodes: 3,
				PackageId:     "pkg-001",
			},
			CreatedAt: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "cluster-abc")
	require.NoError(t, err)

	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "my-cluster")
	assert.Contains(t, stdout, "HEALTHY")
	assert.Contains(t, stdout, "1.13.0")
	assert.Contains(t, stdout, "3")
	assert.Contains(t, stdout, "pkg-001")
	assert.Contains(t, stdout, "aws")
	assert.Contains(t, stdout, "us-east-1")
	assert.Contains(t, stdout, "ago")
}

func TestDescribeCluster_MinimalCluster(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "minimal-id",
			Name: "minimal-name",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "minimal-id")
	require.NoError(t, err)

	assert.Contains(t, stdout, "minimal-id")
	assert.Contains(t, stdout, "minimal-name")
	assert.NotContains(t, stdout, "Status:")
	assert.NotContains(t, stdout, "Version:")
}

func TestDescribeCluster_Labels(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "labeled-cluster",
			Name: "labeled",
			Labels: []*commonv1.KeyValue{
				{Key: "env", Value: "production"},
				{Key: "team", Value: "platform"},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "labeled-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "env=production")
	assert.Contains(t, stdout, "team=platform")
}

func TestDescribeCluster_DatabaseConfiguration(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "dbcfg-cluster",
			Name: "dbcfg",
			Configuration: &clusterv1.ClusterConfiguration{
				DatabaseConfiguration: &clusterv1.DatabaseConfiguration{
					Collection: &clusterv1.DatabaseConfigurationCollection{
						ReplicationFactor:      new(uint32(3)),
						WriteConsistencyFactor: new(int32(2)),
						Vectors: &clusterv1.DatabaseConfigurationCollectionVectors{
							OnDisk: new(true),
						},
					},
					Storage: &clusterv1.DatabaseConfigurationStorage{
						Performance: &clusterv1.DatabaseConfigurationStoragePerformance{
							OptimizerCpuBudget: new(int32(4)),
							AsyncScorer:        new(true),
						},
					},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "dbcfg-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Database Configuration:")
	assert.Contains(t, stdout, "Collection Defaults:")
	assert.Contains(t, stdout, "Replication Factor:")
	assert.Contains(t, stdout, "3")
	assert.Contains(t, stdout, "Write Consistency Factor:")
	assert.Contains(t, stdout, "2")
	assert.Contains(t, stdout, "On Disk:")
	assert.Contains(t, stdout, "yes")
	assert.Contains(t, stdout, "Advanced Optimizations:")
	assert.Contains(t, stdout, "Optimizer CPU Budget:")
	assert.Contains(t, stdout, "4")
	assert.Contains(t, stdout, "Async Scorer:")
}

func TestDescribeCluster_DatabaseConfiguration_NotSet(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "dbcfg-notset",
			Name: "dbcfg-notset",
			Configuration: &clusterv1.ClusterConfiguration{
				DatabaseConfiguration: &clusterv1.DatabaseConfiguration{},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "dbcfg-notset")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Database Configuration:")
	assert.Contains(t, stdout, "(not set)")
}

func TestDescribeCluster_ClusterConfiguration(t *testing.T) {
	env := testutil.NewTestEnv(t)

	rollingPolicy := clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_ROLLING
	byCountStrategy := clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "clustercfg-id",
			Name: "clustercfg",
			Configuration: &clusterv1.ClusterConfiguration{
				AllowedIpSourceRanges: []string{"10.0.0.0/8", "192.168.1.0/24"},
				RestartPolicy:         &rollingPolicy,
				RebalanceStrategy:     &byCountStrategy,
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "clustercfg-id")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Cluster Configuration:")
	assert.Contains(t, stdout, "10.0.0.0/8, 192.168.1.0/24")
	assert.Contains(t, stdout, "rolling")
	assert.Contains(t, stdout, "by-count")
}

func TestDescribeCluster_Endpoint(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "endpoint-cluster",
			Name: "endpoint",
			State: &clusterv1.ClusterState{
				Endpoint: &clusterv1.ClusterEndpoint{
					Url:      "https://my-cluster.cloud.qdrant.io",
					RestPort: 6333,
					GrpcPort: 6334,
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "endpoint-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "https://my-cluster.cloud.qdrant.io")
	assert.Contains(t, stdout, "6333")
	assert.Contains(t, stdout, "6334")
}

func TestDescribeCluster_Resources(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "resources-cluster",
			Name: "resources",
			Configuration: &clusterv1.ClusterConfiguration{
				ClusterStorageConfiguration: &clusterv1.ClusterStorageConfiguration{
					StorageTierType: commonv1.StorageTierType_STORAGE_TIER_TYPE_PERFORMANCE,
				},
			},
			State: &clusterv1.ClusterState{
				Resources: &clusterv1.ClusterNodeResourcesSummary{
					Disk: &clusterv1.ClusterNodeResources{
						Base:      20.0,
						Available: 18.5,
					},
					Ram: &clusterv1.ClusterNodeResources{
						Base:      8.0,
						Reserved:  1.6,
						Available: 6.4,
					},
					Cpu: &clusterv1.ClusterNodeResources{
						Base:      4000.0,
						Reserved:  800.0,
						Available: 3200.0,
					},
					Gpu: &clusterv1.ClusterNodeResources{
						Base:      1000.0,
						Reserved:  0.0,
						Available: 1000.0,
					},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "resources-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Resources (per node):")
	assert.Contains(t, stdout, "20.00 GiB")
	assert.Contains(t, stdout, "18.50 GiB")
	assert.Contains(t, stdout, "performance")
	assert.Contains(t, stdout, "8.00 GiB")
	assert.Contains(t, stdout, "6.40 GiB")
	assert.Contains(t, stdout, "4000m")
	assert.Contains(t, stdout, "3200m")
	assert.Contains(t, stdout, "1000m")
}

func TestDescribeCluster_Nodes(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "nodes-cluster",
			Name: "nodes",
			State: &clusterv1.ClusterState{
				Nodes: []*clusterv1.ClusterNodeInfo{
					{
						Name:      "node-0",
						State:     clusterv1.ClusterNodeState_CLUSTER_NODE_STATE_HEALTHY,
						Version:   "1.13.0",
						StartedAt: timestamppb.New(time.Now().Add(-2 * time.Hour)),
					},
					{
						Name:    "node-1",
						State:   clusterv1.ClusterNodeState_CLUSTER_NODE_STATE_HEALTHY,
						Version: "1.13.0",
					},
				},
			},
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "nodes-cluster")
	require.NoError(t, err)

	assert.Contains(t, stdout, "Nodes:")
	assert.Contains(t, stdout, "node-0")
	assert.Contains(t, stdout, "node-1")
	assert.Contains(t, stdout, "HEALTHY")
	assert.Contains(t, stdout, "1.13.0")
	assert.Contains(t, stdout, "ago")
}

func TestDescribeCluster_JSONOutput(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{
			Id:   "json-cluster",
			Name: "json-name",
		},
	}, nil)

	stdout, _, err := testutil.Exec(t, env, "cluster", "describe", "json-cluster", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"id"`)
	assert.Contains(t, stdout, `"json-cluster"`)
	assert.Contains(t, stdout, `"json-name"`)
}

func TestDescribeCluster_AccountIDPassedToServer(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "any"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "describe", "any")
	require.NoError(t, err)

	req, ok := env.Server.GetClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
}

func TestDescribeCluster_ClusterIDPassedToServer(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Returns(&clusterv1.GetClusterResponse{
		Cluster: &clusterv1.Cluster{Id: "target-id"},
	}, nil)

	_, _, err := testutil.Exec(t, env, "cluster", "describe", "target-id")
	require.NoError(t, err)

	req, ok := env.Server.GetClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "target-id", req.GetClusterId())
}

func TestDescribeCluster_MissingArg(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env, "cluster", "describe")
	require.Error(t, err)
}
