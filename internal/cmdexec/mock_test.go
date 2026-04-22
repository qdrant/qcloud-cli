package cmdexec

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockRunner_RecordsCalls(t *testing.T) {
	runner := NewMockRunner().
		Respond("git", []string{"status"}, &CmdResult{Stdout: []byte("ok")}, nil)

	result, err := runner.Run(context.Background(), "git", "status")

	require.NoError(t, err)
	assert.Equal(t, []byte("ok"), result.Stdout)
	require.Equal(t, 1, runner.CallCount())
	assert.Equal(t, Invocation{Name: "git", Args: []string{"status"}}, runner.Call(0))
}

func TestMockRunner_UnconfiguredCommandReturnsError(t *testing.T) {
	runner := NewMockRunner()

	result, err := runner.Run(context.Background(), "unknown")

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

func TestMockRunner_RespondWithError(t *testing.T) {
	runner := NewMockRunner().
		Respond("brew", []string{"--prefix"}, nil, fmt.Errorf("not found"))

	result, err := runner.Run(context.Background(), "brew", "--prefix")

	assert.Nil(t, result)
	require.EqualError(t, err, "not found")
	require.Equal(t, 1, runner.CallCount())
	assert.Equal(t, Invocation{Name: "brew", Args: []string{"--prefix"}}, runner.Call(0))
}

func TestMockRunner_MultipleCalls(t *testing.T) {
	runner := NewMockRunner().
		Respond("ls", []string{"-la"}, &CmdResult{}, nil).
		Respond("ls", []string{"-R"}, &CmdResult{}, nil)

	_, _ = runner.Run(context.Background(), "ls", "-la")
	_, _ = runner.Run(context.Background(), "ls", "-R")

	require.Equal(t, 2, runner.CallCount())
	assert.Equal(t, Invocation{Name: "ls", Args: []string{"-la"}}, runner.Call(0))
	assert.Equal(t, Invocation{Name: "ls", Args: []string{"-R"}}, runner.Call(1))
}

func TestMockRunner_NoCalls(t *testing.T) {
	runner := NewMockRunner()

	assert.Equal(t, 0, runner.CallCount())
}
