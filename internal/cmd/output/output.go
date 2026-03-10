package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Table renders items as an ASCII table.
type Table struct {
	w       io.Writer
	headers []string
	fields  []func(any) string
}

// NewTable creates a new Table that writes to stdout.
func NewTable() *Table {
	return &Table{w: os.Stdout}
}

// AddField adds a column to the table with a header name and a field extraction function.
func (t *Table) AddField(name string, fn func(any) string) {
	t.headers = append(t.headers, name)
	t.fields = append(t.fields, fn)
}

// Write renders the table with the given items.
func (t *Table) Write(items []any) {
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

// PrintJSON marshals items as JSON and prints to stdout.
// For proto messages, uses protojson for proper field naming.
func PrintJSON(items any) error {
	// Handle slice of proto messages.
	if msgs, ok := items.([]proto.Message); ok {
		return printProtoSlice(msgs)
	}
	// Handle single proto message.
	if msg, ok := items.(proto.Message); ok {
		return printProtoSingle(msg)
	}
	// Fallback: standard JSON.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func printProtoSlice(msgs []proto.Message) error {
	marshaler := protojson.MarshalOptions{Indent: "  "}
	fmt.Print("[")
	for i, msg := range msgs {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Println()
		b, err := marshaler.Marshal(msg)
		if err != nil {
			return err
		}
		fmt.Print("  ")
		os.Stdout.Write(b)
	}
	if len(msgs) > 0 {
		fmt.Println()
	}
	fmt.Println("]")
	return nil
}

func printProtoSingle(msg proto.Message) error {
	marshaler := protojson.MarshalOptions{Indent: "  "}
	b, err := marshaler.Marshal(msg)
	if err != nil {
		return err
	}
	os.Stdout.Write(b)
	fmt.Println()
	return nil
}
