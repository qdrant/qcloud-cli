package cmdexec

import "context"

// Runner executes external commands.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (*CmdResult, error)
}

// CmdResult holds the output of an executed command.
type CmdResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

