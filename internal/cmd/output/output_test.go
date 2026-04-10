package output_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qdrant/qcloud-cli/internal/cmd/output"
)

type testItem struct {
	ID   string
	Name string
}

func TestTable_Render_WithHeaders(t *testing.T) {
	var buf bytes.Buffer
	tbl := output.NewTable[testItem](&buf)
	tbl.AddField("ID", func(v testItem) string { return v.ID })
	tbl.AddField("NAME", func(v testItem) string { return v.Name })
	tbl.SetItems([]testItem{
		{ID: "1", Name: "alpha"},
	})
	tbl.Render()

	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "1")
	assert.Contains(t, out, "alpha")
}

func TestTable_Render_NoHeaders(t *testing.T) {
	var buf bytes.Buffer
	tbl := output.NewTable[testItem](&buf)
	tbl.AddField("ID", func(v testItem) string { return v.ID })
	tbl.AddField("NAME", func(v testItem) string { return v.Name })
	tbl.SetItems([]testItem{
		{ID: "1", Name: "alpha"},
	})
	tbl.SetNoHeaders(true)
	tbl.Render()

	out := buf.String()
	assert.NotContains(t, out, "ID")
	assert.NotContains(t, out, "NAME")
	assert.Contains(t, out, "1")
	assert.Contains(t, out, "alpha")
}

func TestTable_Write_BackwardCompat(t *testing.T) {
	var buf bytes.Buffer
	tbl := output.NewTable[testItem](&buf)
	tbl.AddField("ID", func(v testItem) string { return v.ID })
	tbl.AddField("NAME", func(v testItem) string { return v.Name })
	tbl.Write([]testItem{
		{ID: "1", Name: "alpha"},
	})

	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "1")
	assert.Contains(t, out, "alpha")
}
