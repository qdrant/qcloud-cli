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

func TestUpdateCluster_SetAllowedIPs(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--allowed-ip", "10.0.0.0/8",
		"--allowed-ip", "172.16.0.0/12",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	ips := req.GetCluster().GetConfiguration().GetAllowedIpSourceRanges()
	assert.Equal(t, []string{"10.0.0.0/8", "172.16.0.0/12"}, ips)
}

func TestUpdateCluster_AddAllowedIP(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{
					AllowedIpSourceRanges: []string{"10.0.0.0/8"},
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--allowed-ip", "172.16.0.0/12",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	ips := req.GetCluster().GetConfiguration().GetAllowedIpSourceRanges()
	assert.Equal(t, []string{"10.0.0.0/8", "172.16.0.0/12"}, ips)
}

func TestUpdateCluster_RemoveAllowedIP(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{
					AllowedIpSourceRanges: []string{"10.0.0.0/8", "172.16.0.0/12"},
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--allowed-ip", "172.16.0.0/12-",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	ips := req.GetCluster().GetConfiguration().GetAllowedIpSourceRanges()
	assert.Equal(t, []string{"10.0.0.0/8"}, ips)
}

func TestUpdateCluster_ClearAllowedIPs(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{
					AllowedIpSourceRanges: []string{"10.0.0.0/8", "172.16.0.0/12"},
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--allowed-ip", "10.0.0.0/8-",
		"--allowed-ip", "172.16.0.0/12-",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	ips := req.GetCluster().GetConfiguration().GetAllowedIpSourceRanges()
	assert.Empty(t, ips)
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

func TestUpdateCluster_VersionUpgrade(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--version", "v1.17.0",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "v1.17.0", req.GetCluster().GetConfiguration().GetVersion())
}

func TestUpdateCluster_VersionPromptShowsDiff(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:            req.GetClusterId(),
				Name:          "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{},
				State: &clusterv1.ClusterState{
					Version: "v1.16.2",
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--version", "v1.17.0",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Contains(t, stderr, "v1.16.2 => v1.17.0")
	assert.Contains(t, stderr, "rolling restart")
	assert.Equal(t, 0, env.Server.UpdateClusterCalls.Count())
}

func TestUpdateCluster_DBConfigPromptShowsDiff(t *testing.T) {
	env := testutil.NewTestEnv(t)

	rf := uint32(1)
	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{
					DatabaseConfiguration: &clusterv1.DatabaseConfiguration{
						Collection: &clusterv1.DatabaseConfigurationCollection{
							ReplicationFactor: &rf,
						},
					},
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--replication-factor", "3",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Contains(t, stderr, "1 => 3")
	assert.Contains(t, stderr, "rolling restart")
	assert.NotContains(t, stderr, "Version:")
	assert.Equal(t, 0, env.Server.UpdateClusterCalls.Count())
}

func TestUpdateCluster_VersionAndDBConfigShowSinglePrompt(t *testing.T) {
	env := testutil.NewTestEnv(t)

	rf := uint32(1)
	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{
					DatabaseConfiguration: &clusterv1.DatabaseConfiguration{
						Collection: &clusterv1.DatabaseConfigurationCollection{
							ReplicationFactor: &rf,
						},
					},
				},
				State: &clusterv1.ClusterState{
					Version: "v1.16.2",
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--version", "v1.17.0",
		"--replication-factor", "3",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Contains(t, stderr, "v1.16.2 => v1.17.0")
	assert.Contains(t, stderr, "1 => 3")
	assert.Contains(t, stderr, "rolling restart")
	assert.Equal(t, 0, env.Server.UpdateClusterCalls.Count())
}

func TestUpdateCluster_VersionAndDBConfigForceAppliesBoth(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:            req.GetClusterId(),
				Name:          "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{},
				State: &clusterv1.ClusterState{
					Version: "v1.16.2",
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--version", "v1.17.0",
		"--replication-factor", "3",
		"--force",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "updated successfully")

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "v1.17.0", req.GetCluster().GetConfiguration().GetVersion())
	assert.Equal(t, uint32(3), req.GetCluster().GetConfiguration().GetDatabaseConfiguration().GetCollection().GetReplicationFactor())
}

func TestUpdateCluster_VersionFallsBackToConfigVersion(t *testing.T) {
	env := testutil.NewTestEnv(t)

	v := "v1.15.0"
	env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
		return &clusterv1.GetClusterResponse{
			Cluster: &clusterv1.Cluster{
				Id:   req.GetClusterId(),
				Name: "my-cluster",
				Configuration: &clusterv1.ClusterConfiguration{
					Version: &v,
				},
			},
		}, nil
	})
	env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
		return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
	})

	_, stderr, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--version", "v1.17.0",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "v1.15.0 => v1.17.0")
}

func TestUpdateCluster_DBConfigExtended(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--vectors-on-disk",
		"--db-log-level", "debug",
		"--force",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	dbCfg := req.GetCluster().GetConfiguration().GetDatabaseConfiguration()
	assert.True(t, dbCfg.GetCollection().GetVectors().GetOnDisk())
	assert.Equal(t, clusterv1.DatabaseConfigurationLogLevel_DATABASE_CONFIGURATION_LOG_LEVEL_DEBUG, dbCfg.GetLogLevel())
}

func TestUpdateCluster_AuditLogging(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--audit-logging",
		"--audit-log-rotation", "daily",
		"--audit-log-max-files", "10",
		"--audit-log-trust-forwarded-headers",
		"--force",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	audit := req.GetCluster().GetConfiguration().GetDatabaseConfiguration().GetAuditLogging()
	assert.True(t, audit.GetEnabled())
	assert.Equal(t, clusterv1.AuditLogRotation_AUDIT_LOG_ROTATION_DAILY, audit.GetRotation())
	assert.Equal(t, uint32(10), audit.GetMaxLogFiles())
	assert.True(t, audit.GetTrustForwardedHeaders())
}

func TestUpdateCluster_TLSAndSecrets(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--enable-tls",
		"--api-key-secret", "my-secret:api-key",
		"--read-only-api-key-secret", "ro-secret:rokey",
		"--tls-cert-secret", "cert-secret:tls.crt",
		"--tls-key-secret", "key-secret:tls.key",
		"--force",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	dbCfg := req.GetCluster().GetConfiguration().GetDatabaseConfiguration()
	assert.True(t, dbCfg.GetService().GetEnableTls())
	assert.Equal(t, "my-secret", dbCfg.GetService().GetApiKey().GetName())
	assert.Equal(t, "api-key", dbCfg.GetService().GetApiKey().GetKey())
	assert.Equal(t, "ro-secret", dbCfg.GetService().GetReadOnlyApiKey().GetName())
	assert.Equal(t, "rokey", dbCfg.GetService().GetReadOnlyApiKey().GetKey())
	assert.Equal(t, "cert-secret", dbCfg.GetTls().GetCert().GetName())
	assert.Equal(t, "tls.crt", dbCfg.GetTls().GetCert().GetKey())
	assert.Equal(t, "key-secret", dbCfg.GetTls().GetKey().GetName())
	assert.Equal(t, "tls.key", dbCfg.GetTls().GetKey().GetKey())
}

func TestUpdateCluster_DiskPerformanceAndCostLabel(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--disk-performance", "performance",
		"--cost-allocation-label", "billing-team",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, commonv1.StorageTierType_STORAGE_TIER_TYPE_PERFORMANCE, req.GetCluster().GetConfiguration().GetClusterStorageConfiguration().GetStorageTierType())
	assert.Equal(t, "billing-team", req.GetCluster().GetCostAllocationLabel())
}

func TestUpdateCluster_InvalidDiskPerformance(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--disk-performance", "invalid",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid disk performance")
}

func TestUpdateCluster_HybridServiceType(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--service-type", "load-balancer",
		"--force",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	assert.Equal(t, clusterv1.ClusterServiceType_CLUSTER_SERVICE_TYPE_LOAD_BALANCER, req.GetCluster().GetConfiguration().GetServiceType())
}

func TestUpdateCluster_HybridReservedResources(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--reserved-cpu-percentage", "25",
		"--reserved-memory-percentage", "30",
		"--force",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	cfg := req.GetCluster().GetConfiguration()
	assert.Equal(t, uint32(25), cfg.GetReservedCpuPercentage())
	assert.Equal(t, uint32(30), cfg.GetReservedMemoryPercentage())
}

func TestUpdateCluster_HybridKeyValueFlags(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--node-selector", "zone=us-east",
		"--annotation", "app=qdrant",
		"--pod-label", "tier=db",
		"--service-annotation", "lb=internal",
		"--force",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	cfg := req.GetCluster().GetConfiguration()

	toMap := func(kvs []*commonv1.KeyValue) map[string]string {
		m := make(map[string]string)
		for _, kv := range kvs {
			m[kv.GetKey()] = kv.GetValue()
		}
		return m
	}

	assert.Equal(t, map[string]string{"zone": "us-east"}, toMap(cfg.GetNodeSelector()))
	assert.Equal(t, map[string]string{"app": "qdrant"}, toMap(cfg.GetAnnotations()))
	assert.Equal(t, map[string]string{"tier": "db"}, toMap(cfg.GetPodLabels()))
	assert.Equal(t, map[string]string{"lb": "internal"}, toMap(cfg.GetServiceAnnotations()))
}

func TestUpdateCluster_HybridTolerations(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		setupUpdateHandlers(env)

		_, _, err := testutil.Exec(t, env,
			"cluster", "update", "cluster-abc",
			"--toleration", "env=prod:NoSchedule",
			"--force",
		)
		require.NoError(t, err)

		req, ok := env.Server.UpdateClusterCalls.Last()
		require.True(t, ok)
		tols := req.GetCluster().GetConfiguration().GetTolerations()
		require.Len(t, tols, 1)
		assert.Equal(t, "env", tols[0].GetKey())
		assert.Equal(t, "prod", tols[0].GetValue())
	})

	t.Run("remove", func(t *testing.T) {
		env := testutil.NewTestEnv(t)

		env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			tolKey := "env"
			tolValue := "prod"
			opEqual := clusterv1.TolerationOperator_TOLERATION_OPERATOR_EQUAL
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:   req.GetClusterId(),
					Name: "my-cluster",
					Configuration: &clusterv1.ClusterConfiguration{
						Tolerations: []*clusterv1.Toleration{
							{Key: &tolKey, Value: &tolValue, Operator: &opEqual},
						},
					},
				},
			}, nil
		})
		env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
			return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
		})

		_, _, err := testutil.Exec(t, env,
			"cluster", "update", "cluster-abc",
			"--toleration", "env-",
			"--force",
		)
		require.NoError(t, err)

		req, ok := env.Server.UpdateClusterCalls.Last()
		require.True(t, ok)
		assert.Empty(t, req.GetCluster().GetConfiguration().GetTolerations())
	})
}

func TestUpdateCluster_HybridTopologySpreadConstraints(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		env := testutil.NewTestEnv(t)
		setupUpdateHandlers(env)

		_, _, err := testutil.Exec(t, env,
			"cluster", "update", "cluster-abc",
			"--topology-spread-constraint", "hostname:2:do-not-schedule",
			"--force",
		)
		require.NoError(t, err)

		req, ok := env.Server.UpdateClusterCalls.Last()
		require.True(t, ok)
		tscs := req.GetCluster().GetConfiguration().GetTopologySpreadConstraints()
		require.Len(t, tscs, 1)
		assert.Equal(t, "hostname", tscs[0].GetTopologyKey())
		assert.Equal(t, int32(2), tscs[0].GetMaxSkew())
	})

	t.Run("remove", func(t *testing.T) {
		env := testutil.NewTestEnv(t)

		maxSkew := int32(1)
		env.Server.GetClusterCalls.Always(func(_ context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:   req.GetClusterId(),
					Name: "my-cluster",
					Configuration: &clusterv1.ClusterConfiguration{
						TopologySpreadConstraints: []*commonv1.TopologySpreadConstraint{
							{TopologyKey: "hostname", MaxSkew: &maxSkew},
						},
					},
				},
			}, nil
		})
		env.Server.UpdateClusterCalls.Always(func(_ context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
			return &clusterv1.UpdateClusterResponse{Cluster: req.GetCluster()}, nil
		})

		_, _, err := testutil.Exec(t, env,
			"cluster", "update", "cluster-abc",
			"--topology-spread-constraint", "hostname-",
			"--force",
		)
		require.NoError(t, err)

		req, ok := env.Server.UpdateClusterCalls.Last()
		require.True(t, ok)
		assert.Empty(t, req.GetCluster().GetConfiguration().GetTopologySpreadConstraints())
	})
}

func TestUpdateCluster_HybridStorageClasses(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	_, _, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--database-storage-class", "fast-ssd",
		"--snapshot-storage-class", "standard",
		"--volume-snapshot-class", "default",
		"--volume-attributes-class", "perf",
		"--force",
	)
	require.NoError(t, err)

	req, ok := env.Server.UpdateClusterCalls.Last()
	require.True(t, ok)
	sc := req.GetCluster().GetConfiguration().GetClusterStorageConfiguration()
	assert.Equal(t, "fast-ssd", sc.GetDatabaseStorageClass())
	assert.Equal(t, "standard", sc.GetSnapshotStorageClass())
	assert.Equal(t, "default", sc.GetVolumeSnapshotClass())
	assert.Equal(t, "perf", sc.GetVolumeAttributesClass())
}

func TestUpdateCluster_HybridPromptShowsDiff(t *testing.T) {
	env := testutil.NewTestEnv(t)
	setupUpdateHandlers(env)

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "update", "cluster-abc",
		"--service-type", "node-port",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "Aborted.")
	assert.Contains(t, stderr, "rolling restart")
	assert.Equal(t, 0, env.Server.UpdateClusterCalls.Count())
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
