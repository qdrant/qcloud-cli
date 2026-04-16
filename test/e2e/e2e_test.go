package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var clusterIDRe = regexp.MustCompile(`Cluster\s+([0-9a-f-]{36})`)

func TestE2EClusterCreationFlow(t *testing.T) {
	if os.Getenv("QCLOUD_E2E") == "" {
		t.Skip("set QCLOUD_E2E=1 to run e2e tests")
	}

	apiKey := os.Getenv("QDRANT_CLOUD_API_KEY")
	accountID := os.Getenv("QDRANT_CLOUD_ACCOUNT_ID")
	require.NotEmpty(t, apiKey, "QDRANT_CLOUD_API_KEY must be set")
	require.NotEmpty(t, accountID, "QDRANT_CLOUD_ACCOUNT_ID must be set")

	binaryPath := setupBinary(t)

	r := &runner{
		binaryPath: binaryPath,
		apiKey:     apiKey,
		accountID:  accountID,
		endpoint:   os.Getenv("QDRANT_CLOUD_ENDPOINT"),
		homeDir:    t.TempDir(),
	}

	var clusterID string

	t.Cleanup(func() {
		if clusterID == "" {
			return
		}

		// Best-effort cleanup with a fresh context.
		t.Logf("cleanup: deleting cluster %s", clusterID)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, r.binaryPath, "cluster", "delete", clusterID, "--force")
		cmd.Env = r.env()
		_ = cmd.Run()
	})

	t.Run("cluster-create", func(t *testing.T) {
		out := r.run(t,
			"cluster", "create",
			"--cloud-provider", "aws",
			"--cloud-region", "eu-central-1",
			"--package", "free2",
		)

		matches := clusterIDRe.FindStringSubmatch(out)
		require.NotEmpty(t, matches, "could not extract cluster ID from output: %s", out)
		clusterID = matches[1]
		t.Logf("created cluster %s", clusterID)
	})
	if clusterID == "" {
		t.Fatal("cluster create failed, cannot continue")
	}

	t.Run("cluster-wait", func(t *testing.T) {
		r.runWithTimeout(t, 15*time.Minute,
			"cluster", "wait", clusterID, "--timeout", "15m",
		)
	})

	t.Run("cluster-key-create", func(t *testing.T) {
		r.run(t,
			"cluster", "key", "create", clusterID,
			"--name", "e2e-test-key",
			"--wait",
		)
	})

	t.Run("cluster-delete", func(t *testing.T) {
		r.run(t,
			"cluster", "delete", clusterID, "--force",
		)
		clusterID = "" // Prevent double-delete in t.Cleanup.
	})
}
