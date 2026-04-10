package base_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qdrant/qcloud-cli/internal/cmd/base"
	"github.com/qdrant/qcloud-cli/internal/cmd/output"
	"github.com/qdrant/qcloud-cli/internal/state"
)

var fetchHello = func(_ *state.State, _ *cobra.Command) (string, error) { return "hello", nil }

func execListCmd(t *testing.T, lc base.ListCmd[string], args ...string) (string, error) {
	t.Helper()
	cmd := lc.CobraCommand(state.New(""))
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func stringTableRenderer(out io.Writer, val string) output.TableRenderer {
	tbl := output.NewTable[string](out)
	tbl.AddField("VALUE", func(v string) string { return v })
	tbl.SetItems([]string{val})
	return tbl
}

func TestListCmd_OutputTable(t *testing.T) {
	lc := base.ListCmd[string]{
		Use:   "test",
		Fetch: fetchHello,
		OutputTable: func(_ *cobra.Command, out io.Writer, resp string) (output.TableRenderer, error) {
			return stringTableRenderer(out, resp), nil
		},
	}

	stdout, err := execListCmd(t, lc)
	require.NoError(t, err)
	assert.Contains(t, stdout, "VALUE")
	assert.Contains(t, stdout, "hello")
}

func TestListCmd_WithArgs(t *testing.T) {
	lc := base.ListCmd[string]{
		Use:  "test <value>",
		Args: cobra.ExactArgs(1),
		Fetch: func(_ *state.State, cmd *cobra.Command) (string, error) {
			return cmd.Flags().Arg(0), nil
		},
		OutputTable: func(_ *cobra.Command, out io.Writer, resp string) (output.TableRenderer, error) {
			return stringTableRenderer(out, resp), nil
		},
	}

	stdout, err := execListCmd(t, lc, "world")
	require.NoError(t, err)
	assert.Contains(t, stdout, "world")
}

func TestListCmd_OutputTable_NoHeaders(t *testing.T) {
	lc := base.ListCmd[string]{
		Use:   "test",
		Fetch: fetchHello,
		OutputTable: func(_ *cobra.Command, out io.Writer, resp string) (output.TableRenderer, error) {
			return stringTableRenderer(out, resp), nil
		},
	}

	stdout, err := execListCmd(t, lc, "--no-headers")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "VALUE")
	assert.Contains(t, stdout, "hello")
}
