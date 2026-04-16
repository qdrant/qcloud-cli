package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// E2EEnv bundles everything an e2e test needs: the binary path, credentials,
// endpoint, and an isolated home directory. It is the e2e counterpart of
// testutil.TestEnv.
type E2EEnv struct {
	BinaryPath string
	APIKey     string
	AccountID  string
	Endpoint   string
	HomeDir    string
}

// NewE2EEnv creates an E2EEnv by reading required environment variables,
// setting up the binary, and creating an isolated home directory. It skips the
// test if QCLOUD_E2E is not set and fails if required credentials are missing.
func NewE2EEnv(t *testing.T) *E2EEnv {
	t.Helper()

	if os.Getenv("QCLOUD_E2E") == "" {
		t.Skip("set QCLOUD_E2E=1 to run e2e tests")
	}

	apiKey := os.Getenv("QDRANT_CLOUD_API_KEY")
	accountID := os.Getenv("QDRANT_CLOUD_ACCOUNT_ID")
	require.NotEmpty(t, apiKey, "QDRANT_CLOUD_API_KEY must be set")
	require.NotEmpty(t, accountID, "QDRANT_CLOUD_ACCOUNT_ID must be set")

	return &E2EEnv{
		BinaryPath: setupBinary(t),
		APIKey:     apiKey,
		AccountID:  accountID,
		Endpoint:   os.Getenv("QDRANT_CLOUD_ENDPOINT"),
		HomeDir:    t.TempDir(),
	}
}

// Run executes the qcloud binary with the given arguments. It fails the test
// on non-zero exit and returns stdout.
func (e *E2EEnv) Run(t *testing.T, args ...string) string {
	t.Helper()
	return e.RunWithTimeout(t, 2*time.Minute, args...)
}

// RunWithTimeout executes the qcloud binary with the given arguments and
// timeout. It fails the test on non-zero exit and returns stdout.
func (e *E2EEnv) RunWithTimeout(t *testing.T, timeout time.Duration, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.BinaryPath, args...)
	cmd.Env = e.Env()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	t.Logf("exec: qcloud %s", strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		t.Logf("stdout:\n%s", e.Redact(stdout.String()))
		t.Logf("stderr:\n%s", e.Redact(stderr.String()))
		require.NoError(t, err, "command failed: %s %s", e.BinaryPath, strings.Join(args, " "))
	}

	return stdout.String()
}

// Redact replaces the API key value in s to prevent leaking secrets in CI logs.
func (e *E2EEnv) Redact(s string) string {
	if e.APIKey != "" {
		s = strings.ReplaceAll(s, e.APIKey, "[REDACTED]")
	}
	return s
}

// Env returns the environment variables for the qcloud binary.
func (e *E2EEnv) Env() []string {
	env := []string{
		fmt.Sprintf("QDRANT_CLOUD_API_KEY=%s", e.APIKey),
		fmt.Sprintf("QDRANT_CLOUD_ACCOUNT_ID=%s", e.AccountID),
		"HOME=" + e.HomeDir,
	}
	if e.Endpoint != "" {
		env = append(env, fmt.Sprintf("QDRANT_CLOUD_ENDPOINT=%s", e.Endpoint))
	}
	return env
}
