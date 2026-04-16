package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// runner executes the qcloud binary with the appropriate environment.
type runner struct {
	binaryPath string
	apiKey     string
	accountID  string
	endpoint   string
	homeDir    string
}

// run executes the qcloud binary with the given arguments. It fails the test
// on non-zero exit and returns stdout.
func (r *runner) run(t *testing.T, args ...string) string {
	t.Helper()
	return r.runWithTimeout(t, 2*time.Minute, args...)
}

// runWithTimeout executes the qcloud binary with the given arguments and
// timeout. It fails the test on non-zero exit and returns stdout.
func (r *runner) runWithTimeout(t *testing.T, timeout time.Duration, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.binaryPath, args...)
	cmd.Env = r.env()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	t.Logf("exec: qcloud %s", strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		t.Logf("stdout:\n%s", r.redact(stdout.String()))
		t.Logf("stderr:\n%s", r.redact(stderr.String()))
		require.NoError(t, err, "command failed: %s %s", r.binaryPath, strings.Join(args, " "))
	}

	return stdout.String()
}

// redact replaces the API key value in s to prevent leaking secrets in CI logs.
func (r *runner) redact(s string) string {
	if r.apiKey != "" {
		s = strings.ReplaceAll(s, r.apiKey, "[REDACTED]")
	}
	return s
}

func (r *runner) env() []string {
	env := []string{
		fmt.Sprintf("QDRANT_CLOUD_API_KEY=%s", r.apiKey),
		fmt.Sprintf("QDRANT_CLOUD_ACCOUNT_ID=%s", r.accountID),
		"HOME=" + r.homeDir,
	}
	if r.endpoint != "" {
		env = append(env, fmt.Sprintf("QDRANT_CLOUD_ENDPOINT=%s", r.endpoint))
	}
	return env
}
