package cmdexec

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockRunner_RecordsCalls(t *testing.T) {
	runner := NewMockRunner().
		Respond([]string{"git", "status"}, &CmdResult{Stdout: []byte("ok")}, nil)

	result, err := runner.Run("git", "status")

	require.NoError(t, err)
	assert.Equal(t, []byte("ok"), result.Stdout)
	require.Equal(t, 1, runner.CallCount())
	assert.Equal(t, []string{"git", "status"}, runner.Call(0))
}

func TestMockRunner_UnconfiguredCommandReturnsError(t *testing.T) {
	runner := NewMockRunner()

	result, err := runner.Run("unknown")

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

func TestMockRunner_RespondWithError(t *testing.T) {
	runner := NewMockRunner().
		Respond([]string{"brew", "--prefix"}, nil, fmt.Errorf("not found"))

	result, err := runner.Run("brew", "--prefix")

	assert.Nil(t, result)
	require.EqualError(t, err, "not found")
	require.Equal(t, 1, runner.CallCount())
	assert.Equal(t, []string{"brew", "--prefix"}, runner.Call(0))
}

func TestMockRunner_MultipleCalls(t *testing.T) {
	runner := NewMockRunner().
		Respond([]string{"ls", "-la"}, &CmdResult{}, nil).
		Respond([]string{"ls", "-R"}, &CmdResult{}, nil)

	_, _ = runner.Run("ls", "-la")
	_, _ = runner.Run("ls", "-R")

	require.Equal(t, 2, runner.CallCount())
	assert.Equal(t, []string{"ls", "-la"}, runner.Call(0))
	assert.Equal(t, []string{"ls", "-R"}, runner.Call(1))
}

func TestMockRunner_NoCalls(t *testing.T) {
	runner := NewMockRunner()

	assert.Equal(t, 0, runner.CallCount())
}
