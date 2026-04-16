package e2e_test

import (
	"context"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var clusterIDRe = regexp.MustCompile(`Cluster\s+([0-9a-f-]{36})`)

func TestE2EClusterCreationFlow(t *testing.T) {
	env := NewE2EEnv(t)

	var clusterID string

	t.Cleanup(func() {
		if clusterID == "" {
			return
		}

		// Best-effort cleanup with a fresh context.
		t.Logf("cleanup: deleting cluster %s", clusterID)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, env.BinaryPath, "cluster", "delete", clusterID, "--force")
		cmd.Env = env.Env()
		_ = cmd.Run()
	})

	t.Run("cluster-create", func(t *testing.T) {
		out := env.Run(t,
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
		env.RunWithTimeout(t, 15*time.Minute,
			"cluster", "wait", clusterID, "--timeout", "15m",
		)
	})

	t.Run("cluster-key-create", func(t *testing.T) {
		env.Run(t,
			"cluster", "key", "create", clusterID,
			"--name", "e2e-test-key",
			"--wait",
		)
	})

	t.Run("cluster-delete", func(t *testing.T) {
		env.Run(t,
			"cluster", "delete", clusterID, "--force",
		)
		clusterID = "" // Prevent double-delete in t.Cleanup.
	})
}
