// Package printer renders resources in the formats selected by the -o flag.
package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format enumerates supported -o values.
type Format string

const (
	FormatTable Format = "table"
	FormatWide  Format = "wide"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// Parse normalizes a raw -o flag string. An empty value defaults to FormatTable.
func Parse(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "table":
		return FormatTable, nil
	case "wide":
		return FormatWide, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	}
	return "", fmt.Errorf("unsupported output format %q (want: table|wide|json|yaml)", s)
}

// Table is a set of rows ready for tabular rendering.
type Table struct {
	Headers []string
	Rows    [][]string
}

// WriteTable emits t through a text/tabwriter to w.
func WriteTable(w io.Writer, t Table) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if len(t.Headers) > 0 {
		fmt.Fprintln(tw, strings.Join(t.Headers, "\t"))
	}
	for _, row := range t.Rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

// WriteJSON pretty-prints v as JSON.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteYAML emits v as YAML.
func WriteYAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(v)
}

// WriteStructured writes v when format is JSON or YAML. The returned handled
// value is false for table-oriented formats so callers can apply their own
// fallback renderer.
func WriteStructured(w io.Writer, format Format, v any) (handled bool, err error) {
	switch format {
	case FormatJSON:
		return true, WriteJSON(w, v)
	case FormatYAML:
		return true, WriteYAML(w, v)
	default:
		return false, nil
	}
}
