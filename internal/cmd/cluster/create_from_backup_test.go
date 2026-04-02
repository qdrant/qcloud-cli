package cluster_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/v1"

	"github.com/qdrant/qcloud-cli/internal/testutil"
)

func TestCreateFromBackup_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterFromBackupCalls.Returns(&clusterv1.CreateClusterFromBackupResponse{
		Cluster: &clusterv1.Cluster{Id: "cluster-restored", Name: "my-restored-cluster"},
	}, nil)

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "create-from-backup",
		"--backup-id", "backup-abc",
		"--name", "my-restored-cluster",
	)
	require.NoError(t, err)

	req, ok := env.Server.CreateClusterFromBackupCalls.Last()
	require.True(t, ok)
	assert.Equal(t, "test-account-id", req.GetAccountId())
	assert.Equal(t, "backup-abc", req.GetBackupId())
	assert.Equal(t, "my-restored-cluster", req.GetClusterName())
	assert.Contains(t, stdout, "cluster-restored")
	assert.Contains(t, stdout, "my-restored-cluster")
}

func TestCreateFromBackup_MissingBackupID(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create-from-backup",
		"--name", "my-restored-cluster",
	)
	require.Error(t, err)
	assert.Equal(t, 0, env.Server.CreateClusterFromBackupCalls.Count())
}

func TestCreateFromBackup_MissingName(t *testing.T) {
	env := testutil.NewTestEnv(t)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create-from-backup",
		"--backup-id", "backup-abc",
	)
	require.Error(t, err)
	assert.Equal(t, 0, env.Server.CreateClusterFromBackupCalls.Count())
}

func TestCreateFromBackup_APIError(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterFromBackupCalls.Returns(nil, assert.AnError)

	_, _, err := testutil.Exec(t, env,
		"cluster", "create-from-backup",
		"--backup-id", "backup-abc",
		"--name", "my-restored-cluster",
	)
	require.Error(t, err)
}

func TestCreateFromBackup_NoWait(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterFromBackupCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterFromBackupRequest) (*clusterv1.CreateClusterFromBackupResponse, error) {
		return &clusterv1.CreateClusterFromBackupResponse{
			Cluster: &clusterv1.Cluster{
				Id:   "cluster-restored",
				Name: req.GetClusterName(),
			},
		}, nil
	})

	stdout, _, err := testutil.Exec(t, env,
		"cluster", "create-from-backup",
		"--backup-id", "backup-abc",
		"--name", "my-restored-cluster",
	)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cluster-restored")
	assert.Equal(t, 0, env.Server.GetClusterCalls.Count(), "GetCluster should not be called without --wait")
}

func TestCreateFromBackup_WaitSuccess(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.Server.CreateClusterFromBackupCalls.Always(func(_ context.Context, req *clusterv1.CreateClusterFromBackupRequest) (*clusterv1.CreateClusterFromBackupResponse, error) {
		return &clusterv1.CreateClusterFromBackupResponse{
			Cluster: &clusterv1.Cluster{
				Id:   "cluster-restored",
				Name: req.GetClusterName(),
			},
		}, nil
	})
	env.Server.GetClusterCalls.
		OnCall(0, func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id:    "cluster-restored",
					State: &clusterv1.ClusterState{Phase: clusterv1.ClusterPhase_CLUSTER_PHASE_CREATING},
				},
			}, nil
		}).
		Always(func(_ context.Context, _ *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
			return &clusterv1.GetClusterResponse{
				Cluster: &clusterv1.Cluster{
					Id: "cluster-restored",
					State: &clusterv1.ClusterState{
						Phase:    clusterv1.ClusterPhase_CLUSTER_PHASE_HEALTHY,
						Endpoint: &clusterv1.ClusterEndpoint{Url: "https://restored.aws.cloud.qdrant.io"},
					},
				},
			}, nil
		})

	stdout, stderr, err := testutil.Exec(t, env,
		"cluster", "create-from-backup",
		"--backup-id", "backup-abc",
		"--name", "my-restored-cluster",
		"--wait",
		"--wait-timeout", "30s",
		"--wait-poll-interval", "10ms",
	)
	require.NoError(t, err)
	assert.Contains(t, stderr, "cluster-restored")
	assert.Contains(t, stdout, "cluster-restored")
	assert.Contains(t, stdout, "https://restored.aws.cloud.qdrant.io")
	assert.Positive(t, env.Server.GetClusterCalls.Count())
}
