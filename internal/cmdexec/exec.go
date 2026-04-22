package cmdexec

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

// ExecRunner is the default Runner that delegates to os/exec.
type ExecRunner struct{}

// Run executes a command and returns its stdout, stderr, and exit code.
// The returned error is non-nil only when the command cannot be started.
func (ExecRunner) Run(ctx context.Context, name string, args ...string) (*CmdResult, error) {
	c := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	return &CmdResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: exitCode,
	}, nil
}
