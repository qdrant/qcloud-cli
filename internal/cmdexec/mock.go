package cmdexec

import (
	"fmt"
	"strings"
)

type mockResponse struct {
	result *CmdResult
	err    error
}

// MockRunner is a test double for CommandRunner that records invocations
// and returns pre-configured responses keyed by command name.
type MockRunner struct {
	responses map[string]mockResponse
	calls     [][]string
}

// NewMockRunner creates a MockRunner with no configured responses.
func NewMockRunner() *MockRunner {
	return &MockRunner{responses: make(map[string]mockResponse)}
}

// Respond configures the response returned when a command with the given name
// is executed. Returns the receiver for chaining.
func (m *MockRunner) Respond(cmd []string, result *CmdResult, err error) *MockRunner {
	m.responses[m.cmdKey(cmd)] = mockResponse{result: result, err: err}
	return m
}

// Run records the call and returns the configured response for the command name.
// Returns an error if no response is configured.
func (m *MockRunner) Run(cmd ...string) (*CmdResult, error) {
	m.calls = append(m.calls, cmd)
	resp, ok := m.responses[m.cmdKey(cmd)]
	if !ok {
		return nil, fmt.Errorf("MockRunner: no response configured for command %q", strings.Join(cmd, " "))
	}

	return resp.result, resp.err
}

// CallCount returns the total number of recorded invocations.
func (m *MockRunner) CallCount() int {
	return len(m.calls)
}

// Call returns the recorded invocation at index n.
// Panics if n is out of range.
func (m *MockRunner) Call(n int) []string {
	return m.calls[n]
}

func (m *MockRunner) cmdKey(cmd []string) string {
	return strings.Join(cmd, "###")
}

