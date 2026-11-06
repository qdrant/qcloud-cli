package output_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

func TestColorEnabled_NotATerminal(t *testing.T) {
	assert.False(t, output.ColorEnabled(&bytes.Buffer{}))
}

func TestTerminalWidth_NotATerminal(t *testing.T) {
	assert.Equal(t, 100, output.TerminalWidth(&bytes.Buffer{}, 100))
}
