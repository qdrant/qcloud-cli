package cluster_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"
	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestUpdateCluster_SetLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster"},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{
			Cluster: req.GetCluster(),
		}, nil
	})

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--label", "env=prod",
		"--label", "team=platform",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-abc")
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	capturedLabels := make(map[string]string)
	for _, kv := range req.GetCluster().GetLabels() {
		capturedLabels[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, map[string]string{"env": "prod", "team": "platform"}, capturedLabels)
}

func TestUpdateCluster_ClearLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Labels: []*commonv1.KeyValue{
					{Key: "env", Value: "staging"},
					{Key: "team", Value: "infra"},
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{
			Cluster: req.GetCluster(),
		}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--label", "env-",
		"--label", "team-",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Empty(t, req.GetCluster().GetLabels())
}

func TestUpdateCluster_ApplyLabels(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Labels: []*commonv1.KeyValue{
					{Key: "env", Value: "staging"},
					{Key: "team", Value: "infra"},
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--label", "env=prod",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	capturedLabels := make(map[string]string)
	for _, kv := range req.GetCluster().GetLabels() {
		capturedLabels[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, map[string]string{"env": "prod", "team": "infra"}, capturedLabels)
}

func TestUpdateCluster_RemoveLabel(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Labels: []*commonv1.KeyValue{
					{Key: "env", Value: "staging"},
					{Key: "team", Value: "infra"},
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--label", "team-",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	capturedLabels := make(map[string]string)
	for _, kv := range req.GetCluster().GetLabels() {
		capturedLabels[kv.GetKey()] = kv.GetValue()
	}
	assert.Equal(t, map[string]string{"env": "staging"}, capturedLabels)
}

func TestUpdateCluster_InvalidLabelFormat(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--label", "badformat",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --label value")
}

func TestUpdateCluster_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId()},
		}, nil
	})
	env.Server.UpdateClusterCalls.Returns(nil, fmt.Errorf("internal server error"))

	_, _, err := testutil.Exec(t, env, "cluster", "update", "cluster-abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update cluster")
}

func TestUpdateCluster_ReplicationFactor(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--replication-factor", "3",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	rf := req.GetCluster().GetConfiguration().GetDatabaseConfiguration().GetCollection().GetReplicationFactor()
	assert.Equal(t, uint32(3), rf)
}

func TestUpdateCluster_WriteConsistencyFactor(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--write-consistency-factor", "2",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	wcf := req.GetCluster().GetConfiguration().GetDatabaseConfiguration().GetCollection().GetWriteConsistencyFactor()
	assert.Equal(t, int32(2), wcf)
}

func TestUpdateCluster_AsyncScorer(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--async-scorer",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	as := req.GetCluster().GetConfiguration().GetDatabaseConfiguration().GetStorage().GetPerformance().GetAsyncScorer()
	assert.True(t, as)
}

func TestUpdateCluster_OptimizerCPUBudget(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--optimizer-cpu-budget", "4",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	budget := req.GetCluster().GetConfiguration().GetDatabaseConfiguration().GetStorage().GetPerformance().GetOptimizerCpuBudget()
	assert.Equal(t, int32(4), budget)
}

func TestUpdateCluster_AllowedIPs(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--allowed-ips", "10.0.0.0/8,172.16.0.0/12",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	ips := req.GetCluster().GetConfiguration().GetAllowedIpSourceRanges()
	assert.Equal(t, []string{"10.0.0.0/8", "172.16.0.0/12"}, ips)
}

func TestUpdateCluster_RestartMode(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--restart-mode", "parallel",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	rp := req.GetCluster().GetConfiguration().GetRestartPolicy()
	assert.Equal(t, clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_PARALLEL, rp)
}

func TestUpdateCluster_RebalanceStrategy(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--rebalance-strategy", "by-count-and-size",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	rs := req.GetCluster().GetConfiguration().GetRebalanceStrategy()
	assert.Equal(t, clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_COUNT_AND_SIZE, rs)
}

func TestUpdateCluster_InvalidRestartMode(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--restart-mode", "invalid",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid restart mode")
}

func TestUpdateCluster_InvalidRebalanceStrategy(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--rebalance-strategy", "invalid",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rebalance strategy")
}

func TestUpdateCluster_MultipleDBConfigFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--replication-factor", "3",
		"--write-consistency-factor", "2",
		"--async-scorer",
		"--optimizer-cpu-budget", "-1",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	cluster := req.GetCluster()
	dbCfg := cluster.GetConfiguration().GetDatabaseConfiguration()
	assert.Equal(t, uint32(3), dbCfg.GetCollection().GetReplicationFactor())
	assert.Equal(t, int32(2), dbCfg.GetCollection().GetWriteConsistencyFactor())
	assert.True(t, dbCfg.GetStorage().GetPerformance().GetAsyncScorer())
	assert.Equal(t, int32(-1), dbCfg.GetStorage().GetPerformance().GetOptimizerCpuBudget())
}

func TestUpdateCluster_PreservesExistingConfig(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		existing := clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{
					AllowedIpSourceRanges: []string{"10.0.0.0/8"},
					RebalanceStrategy:     &existing,
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	// Only change restart-mode, everything else should be preserved
	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--restart-mode", "rolling",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	cfg := req.GetCluster().GetConfiguration()
	assert.Equal(t, []string{"10.0.0.0/8"}, cfg.GetAllowedIpSourceRanges())
	assert.Equal(t, clusterv1.ClusterConfigurationRebalanceStrategy_CLUSTER_CONFIGURATION_REBALANCE_STRATEGY_BY_SIZE, cfg.GetRebalanceStrategy())
	assert.Equal(t, clusterv1.ClusterConfigurationRestartPolicy_CLUSTER_CONFIGURATION_RESTART_POLICY_ROLLING, cfg.GetRestartPolicy())
}

// setupUpdateHandlers configures the standard Get/Update handlers for update tests.
func setupUpdateHandlers(env *testutil.TestEnv) {
	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{Id: req.GetClusterId(), Name: "my-cluster"},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})
}
