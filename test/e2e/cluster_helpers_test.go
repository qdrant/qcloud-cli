package e2e_test

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/qdrant/qcloud-cli/test/e2e/framework"
)

// e2eClusterPrefix is the name prefix used for every cluster created via
// createCluster. sweepLeakedClusters targets the same prefix.
const e2eClusterPrefix = "e2e-"

// createClusterOpts captures the knobs tests commonly tweak when creating a
// cluster. Fields left at zero values pick sensible defaults.
type createClusterOpts struct {
	// Name is the cluster name. If empty, a random "e2e-<hex>" name is used.
	Name string
	// CloudProvider defaults to "aws".
	CloudProvider string
	// CloudRegion defaults to "eu-central-1".
	CloudRegion string
	// Package defaults to "free2".
	Package string
}

// createCluster creates a cluster and registers a t.Cleanup that force-deletes
// it when the test finishes. The returned ID is the cluster's GUID. The
// cluster is not waited on; call waitCluster if readiness matters to the test.
func createCluster(t *testing.T, env *framework.Env, opts createClusterOpts) string {
	t.Helper()

	if opts.Name == "" {
		opts.Name = e2eClusterPrefix + randomSuffix(t)
	}
	if opts.CloudProvider == "" {
		opts.CloudProvider = "aws"
	}
	if opts.CloudRegion == "" {
		opts.CloudRegion = "eu-central-1"
	}
	if opts.Package == "" {
		opts.Package = "free2"
	}

	var cluster struct {
		ID string `json:"id"`
	}
	env.RunJSON(t, &cluster,
		"cluster", "create",
		"--name", opts.Name,
		"--cloud-provider", opts.CloudProvider,
		"--cloud-region", opts.CloudRegion,
		"--package", opts.Package,
	)
	if cluster.ID == "" {
		t.Fatalf("cluster create returned empty id")
	}

	t.Logf("created cluster %s (%s)", cluster.ID, opts.Name)
	t.Cleanup(func() { deleteCluster(t, env, cluster.ID) })
	return cluster.ID
}

// waitCluster blocks until the cluster is healthy or the timeout elapses.
func waitCluster(t *testing.T, env *framework.Env, id string, timeout time.Duration) {
	t.Helper()
	env.RunWithTimeout(t, timeout+30*time.Second,
		"cluster", "wait", id, "--timeout", timeout.String(),
	)
}

// deleteCluster force-deletes a cluster, logging but not failing on error so
// it is safe to call from t.Cleanup.
func deleteCluster(t *testing.T, env *framework.Env, id string) {
	t.Helper()
	if id == "" {
		return
	}
	res := env.RunAllowFail(t, "cluster", "delete", id, "--force")
	if res.Err != nil {
		t.Logf("cleanup: cluster delete %s failed: %v", id, res.Err)
	}
}

// sweepLeakedClusters lists clusters in the current account and force-deletes
// any whose name starts with e2eClusterPrefix and whose CreatedAt is older
// than maxAge. It never fails the test — leaks are a best-effort safety net.
func sweepLeakedClusters(t *testing.T, env *framework.Env, maxAge time.Duration) {
	t.Helper()

	var resp struct {
		Items []struct {
			ID        string    `json:"id"`
			Name      string    `json:"name"`
			CreatedAt time.Time `json:"createdAt"`
		} `json:"items"`
	}

	res := env.RunAllowFail(t, "cluster", "list", "--json")
	if res.Err != nil {
		t.Logf("leak sweep: cluster list failed: %v", res.Err)
		return
	}
	if err := json.Unmarshal([]byte(res.Stdout), &resp); err != nil {
		t.Logf("leak sweep: decoding cluster list: %v", err)
		return
	}

	cutoff := time.Now().Add(-maxAge)
	for _, c := range resp.Items {
		if !strings.HasPrefix(c.Name, e2eClusterPrefix) {
			continue
		}
		if !c.CreatedAt.IsZero() && c.CreatedAt.After(cutoff) {
			continue
		}
		t.Logf("leak sweep: deleting stale cluster %s (%s, created %s)",
			c.ID, c.Name, c.CreatedAt.Format(time.RFC3339))
		deleteCluster(t, env, c.ID)
	}
}

// randomSuffix returns 8 hex chars for disambiguating parallel test runs.
func randomSuffix(t *testing.T) string {
	t.Helper()
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return hex.EncodeToString(b[:])
}
