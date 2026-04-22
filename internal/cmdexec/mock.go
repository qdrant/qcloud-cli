package cmdexec

import (
	"context"
	"fmt"
	"strings"
)

// Call records the name and arguments of a single command execution.
type Call struct {
	Name string
	Args []string
}

type mockResponse struct {
	result *CmdResult
	err    error
}

// MockRunner is a test double for Runner that records invocations
// and returns pre-configured responses keyed by command.
type MockRunner struct {
	responses map[string]mockResponse
	calls     []Call
}

// NewMockRunner creates a MockRunner with no configured responses.
func NewMockRunner() *MockRunner {
	return &MockRunner{responses: make(map[string]mockResponse)}
}

// Respond configures the response returned when a command with the given
// name and arguments is executed. Returns the receiver for chaining.
func (m *MockRunner) Respond(name string, args []string, result *CmdResult, err error) *MockRunner {
	m.responses[m.cmdKey(name, args)] = mockResponse{result: result, err: err}
	return m
}

// Run records the call and returns the configured response for the command.
// Returns an error if no response is configured.
func (m *MockRunner) Run(_ context.Context, name string, args ...string) (*CmdResult, error) {
	if args == nil {
		args = []string{}
	}
	m.calls = append(m.calls, Call{Name: name, Args: args})
	resp, ok := m.responses[m.cmdKey(name, args)]
	if !ok {
		return nil, fmt.Errorf("MockRunner: no response configured for command %q", name+" "+strings.Join(args, " "))
	}

	return resp.result, resp.err
}

// CallCount returns the total number of recorded invocations.
func (m *MockRunner) CallCount() int {
	return len(m.calls)
}

// Call returns the recorded invocation at index n.
// Panics if n is out of range.
func (m *MockRunner) Call(n int) Call {
	return m.calls[n]
}

func (m *MockRunner) cmdKey(name string, args []string) string {
	return strings.Join(append([]string{name}, args...), "###")
}
