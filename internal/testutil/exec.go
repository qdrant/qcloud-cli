package testutil

import (
	"bytes"
	"testing"

	"github.com/qdrant/qcloud-cli/internal/cli"
)

// Exec builds and executes the root command with the given args, returning
// captured stdout, stderr, and any error from cmd.Execute().
func Exec(t *testing.T, env *TestEnv, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer

	cmd := cli.NewRootCommand(env.State)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}
