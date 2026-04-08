package base_test

import (
	"bytes"
	"fmt"
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
		OutputTable: func(_ *cobra.Command, out io.Writer, resp string) output.TableRenderer {
			return stringTableRenderer(out, resp)
		},
	}

	stdout, err := execListCmd(t, lc)
	require.NoError(t, err)
	assert.Contains(t, stdout, "VALUE")
	assert.Contains(t, stdout, "hello")
}

func TestListCmd_OutputTable_NoHeaders(t *testing.T) {
	lc := base.ListCmd[string]{
		Use:   "test",
		Fetch: fetchHello,
		OutputTable: func(_ *cobra.Command, out io.Writer, resp string) output.TableRenderer {
			return stringTableRenderer(out, resp)
		},
	}

	stdout, err := execListCmd(t, lc, "--no-headers")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "VALUE")
	assert.Contains(t, stdout, "hello")
}

func TestListCmd_PrintText(t *testing.T) {
	lc := base.ListCmd[string]{
		Use:  "test",
		Fetch: fetchHello,
		PrintText: func(_ *cobra.Command, out io.Writer, resp string) error {
			_, err := fmt.Fprintln(out, resp)
			return err
		},
	}

	stdout, err := execListCmd(t, lc)
	require.NoError(t, err)
	assert.Contains(t, stdout, "hello")
}

func TestListCmd_NeitherOutputTableNorPrintText_Panics(t *testing.T) {
	lc := base.ListCmd[string]{
		Use:   "test",
		Fetch: fetchHello,
	}

	assert.Panics(t, func() {
		_, _ = execListCmd(t, lc)
	})
}
