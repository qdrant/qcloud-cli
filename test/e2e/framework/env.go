package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	envEnable    = "QCLOUD_E2E"
	envAPIKey    = "QDRANT_CLOUD_API_KEY"
	envAccountID = "QDRANT_CLOUD_ACCOUNT_ID"
	envEndpoint  = "QDRANT_CLOUD_ENDPOINT"
	envConfig    = "QDRANT_CLOUD_CONFIG"

	defaultRunTimeout = 2 * time.Minute
)

// Env bundles everything an e2e test needs to shell out to qcloud against a
// real Qdrant Cloud backend.
type Env struct {
	BinaryPath string
	APIKey     string
	AccountID  string
	Endpoint   string

	configPath string
}

// NewEnv returns an Env wired up with credentials from the environment and an
// isolated empty config file. It skips the test if QCLOUD_E2E is not set.
//
// The binary is the one returned by Binary; it is resolved once per process.
func NewEnv(t *testing.T) *Env {
	t.Helper()

	if os.Getenv(envEnable) == "" {
		t.Skipf("set %s=1 to run e2e tests", envEnable)
	}

	apiKey := mustEnv(t, envAPIKey)
	accountID := mustEnv(t, envAccountID)

	bin, err := Binary()
	if err != nil {
		t.Fatalf("resolving qcloud binary: %v", err)
	}

	configPath := writeEmptyConfig(t)

	return &Env{
		BinaryPath: bin,
		APIKey:     apiKey,
		AccountID:  accountID,
		Endpoint:   os.Getenv(envEndpoint),
		configPath: configPath,
	}
}

func mustEnv(t *testing.T, name string) string {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Fatalf("%s must be set", name)
	}
	return v
}

// writeEmptyConfig creates an empty config.yaml in t.TempDir() so the CLI
// doesn't try to read ~/.config/qcloud/config.yaml.
func writeEmptyConfig(t *testing.T) string {
	t.Helper()
	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
		t.Fatalf("writing empty config: %v", err)
	}
	return path
}

// RunResult holds the outcome of a single binary invocation.
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Run invokes the binary with the given args and the default timeout, failing
// the test if the command exits non-zero. Stdout and stderr are streamed to
// t.Log.
func (e *Env) Run(t *testing.T, args ...string) RunResult {
	t.Helper()
	return e.RunWithTimeout(t, defaultRunTimeout, args...)
}

// RunWithTimeout is Run with a caller-supplied timeout.
func (e *Env) RunWithTimeout(t *testing.T, timeout time.Duration, args ...string) RunResult {
	t.Helper()
	res := e.runRaw(t, timeout, args)
	if res.Err != nil {
		t.Fatalf("qcloud %s failed: %v", strings.Join(args, " "), res.Err)
	}
	return res
}

// RunAllowFail is like Run but returns the result without failing the test.
// Useful for commands where a non-zero exit is part of the assertion.
func (e *Env) RunAllowFail(t *testing.T, args ...string) RunResult {
	t.Helper()
	return e.runRaw(t, defaultRunTimeout, args)
}

// RunJSON appends --json to args, runs the binary, and decodes stdout into v.
func (e *Env) RunJSON(t *testing.T, v any, args ...string) {
	t.Helper()
	res := e.Run(t, append(args, "--json")...)
	if err := json.Unmarshal([]byte(res.Stdout), v); err != nil {
		t.Fatalf("decoding JSON output of %q: %v\nstdout: %s",
			strings.Join(args, " "), err, res.Stdout)
	}
}

func (e *Env) runRaw(t *testing.T, timeout time.Duration, args []string) RunResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.BinaryPath, args...)
	cmd.Env = e.commandEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, testLogWriter{t: t, prefix: "stdout: ", redact: e.APIKey})
	cmd.Stderr = io.MultiWriter(&stderr, testLogWriter{t: t, prefix: "stderr: ", redact: e.APIKey})

	t.Logf("exec: qcloud %s", strings.Join(args, " "))

	err := cmd.Run()
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	return RunResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Err:      err,
	}
}

func (e *Env) commandEnv() []string {
	env := []string{
		fmt.Sprintf("%s=%s", envAPIKey, e.APIKey),
		fmt.Sprintf("%s=%s", envAccountID, e.AccountID),
		fmt.Sprintf("%s=%s", envConfig, e.configPath),
		"HOME=" + e.homeDir(),
		"PATH=" + os.Getenv("PATH"),
	}
	if e.Endpoint != "" {
		env = append(env, fmt.Sprintf("%s=%s", envEndpoint, e.Endpoint))
	}
	return env
}

// homeDir points at the directory containing the empty config file so any
// code that falls through to $HOME still lands inside the test sandbox.
func (e *Env) homeDir() string {
	if i := strings.LastIndex(e.configPath, "/"); i > 0 {
		return e.configPath[:i]
	}
	return os.TempDir()
}

// Redact masks the API key and JWT-like tokens in s. Useful when asserting
// against stdout/stderr.
func (e *Env) Redact(s string) string {
	if e.APIKey != "" {
		s = strings.ReplaceAll(s, e.APIKey, "[REDACTED]")
	}
	return jwtLikeRe.ReplaceAllString(s, "[REDACTED]")
}

// jwtLikeRe matches strings that look like JWTs: three base64url segments
// separated by dots, where the first two segments decode to JSON objects.
// We anchor on word boundaries so we don't clip larger tokens.
var jwtLikeRe = regexp.MustCompile(`\b[A-Za-z0-9_-]{4,}\.` +
	`[A-Za-z0-9_-]{4,}\.` +
	`[A-Za-z0-9_-]{4,}\b`)

// testLogWriter forwards writes to t.Log, splitting on newlines and redacting
// secrets so they never end up in CI output.
type testLogWriter struct {
	t      *testing.T
	prefix string
	redact string
}

func (w testLogWriter) Write(p []byte) (int, error) {
	s := string(p)
	if w.redact != "" {
		s = strings.ReplaceAll(s, w.redact, "[REDACTED]")
	}
	s = jwtLikeRe.ReplaceAllString(s, "[REDACTED]")
	for line := range strings.SplitSeq(strings.TrimRight(s, "\n"), "\n") {
		w.t.Log(w.prefix + line)
	}
	return len(p), nil
}
