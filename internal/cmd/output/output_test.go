package output_test

import (
	"bytes"
	"testing"
	"time"

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

func TestDuration(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{24 * time.Hour, "1 day"},
		{30 * 24 * time.Hour, "30 days"},
		{36 * time.Hour, "1 day 12 hours"},
		{400 * 24 * time.Hour, "400 days"},
		{90 * time.Minute, "1 hour 30 minutes"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, output.Duration(tt.in))
	}
}

func TestUnlimitedIfZero(t *testing.T) {
	assert.Equal(t, "unlimited", output.UnlimitedIfZero("0", 0))
	assert.Equal(t, "10", output.UnlimitedIfZero("10", 10))
}

func TestCompactDuration(t *testing.T) {
	cases := map[time.Duration]string{
		0:                          "0s",
		10 * time.Second:           "10s",
		90 * time.Second:           "1m30s",
		2 * time.Minute:            "2m",
		10 * time.Minute:           "10m",
		time.Hour:                  "1h",
		24 * time.Hour:             "24h",
		time.Hour + 30*time.Minute: "1h30m",
		time.Hour + 30*time.Second: "1h0m30s",
		500 * time.Millisecond:     "500ms",
	}
	for d, want := range cases {
		assert.Equal(t, want, output.CompactDuration(d), d.String())
	}
}

func TestRate(t *testing.T) {
	assert.Equal(t, "0.00/s", output.Rate(0))
	assert.Equal(t, "1.25/s", output.Rate(1.2499))
}

func TestMilliseconds(t *testing.T) {
	assert.Equal(t, "0.0ms", output.Milliseconds(0))
	assert.Equal(t, "12.3ms", output.Milliseconds(12.34))
}

func TestBytes(t *testing.T) {
	assert.Equal(t, "0 B", output.Bytes(0))
	assert.Equal(t, "1.5 GiB", output.Bytes(3<<29))
}
