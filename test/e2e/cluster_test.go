package e2e_test

import (
	"testing"
	"time"

	"github.com/qdrant/qcloud-cli/test/e2e/framework"
)

// TestClusterLifecycle exercises the happy path of cluster create → wait →
// key create → delete against a real backend.
func TestClusterLifecycle(t *testing.T) {
	env := framework.NewEnv(t)

	// Best-effort cleanup of clusters orphaned by killed previous runs.
	sweepLeakedClusters(t, env, time.Hour)

	id := createCluster(t, env, createClusterOpts{
		CloudProvider: "aws",
		CloudRegion:   "eu-central-1",
		Package:       "free2",
	})

	waitCluster(t, env, id, 15*time.Minute)

	env.Run(t,
		"cluster", "key", "create", id,
		"--name", "e2e-test-key",
		"--wait",
		"--wait-timeout", "5m",
	)
}
