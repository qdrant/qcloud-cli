package cmdexec

import (
	"bytes"
	"errors"
	"os/exec"
)

// CmdResult holds the output of an executed command.
type CmdResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// ExecRunner is the default CmdRunner that delegates to os/exec.
type ExecRunner struct{}

// Run executes a command and returns its stdout, stderr, and exit code.
// The returned error is non-nil only when the command cannot be started.
// Panics if no arguments are provided.
func (ExecRunner) Run(cmd ...string) (*CmdResult, error) {
	if len(cmd) == 0 {
		panic("cmdexec: Run called with no arguments")
	}
	c := exec.Command(cmd[0], cmd[1:]...)
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
