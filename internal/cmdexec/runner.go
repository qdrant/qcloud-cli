package cmdexec

import "context"

// Runner executes external commands.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (*CmdResult, error)
}
