package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Table renders items as an ASCII table.
type Table[T any] struct {
	w       io.Writer
	headers []string
	fields  []func(T) string
}

// NewTable creates a new Table that writes to the given writer.
func NewTable[T any](w io.Writer) *Table[T] {
	return &Table[T]{w: w}
}

// AddField adds a column to the table with a header name and a field extraction function.
func (t *Table[T]) AddField(name string, fn func(T) string) {
	t.headers = append(t.headers, name)
	t.fields = append(t.fields, fn)
}

// Write renders the table with the given items.
func (t *Table[T]) Write(items []T) {
	tw := table.NewWriter()
	tw.SetOutputMirror(t.w)
	tw.SetStyle(table.StyleLight)

	header := make(table.Row, len(t.headers))
	for i, h := range t.headers {
		header[i] = h
	}
	tw.AppendHeader(header)

	for _, item := range items {
		row := make(table.Row, len(t.fields))
		for i, fn := range t.fields {
			row[i] = fn(item)
		}
		tw.AppendRow(row)
	}
	tw.Render()
}

// PrintJSON marshals items as JSON and writes to w.
// For proto messages, uses protojson for proper field naming.
func PrintJSON(w io.Writer, v any) error {
	if msg, ok := v.(proto.Message); ok {
		marshaler := protojson.MarshalOptions{Indent: "  "}
		b, err := marshaler.Marshal(msg)
		if err != nil {
			return err
		}
		_, _ = w.Write(b)
		fmt.Fprintln(w)
		return nil
	}
	// Fallback: standard JSON.
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
