// Package output renders command results as a table or as JSON.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Mode selects the renderer. It is the value of -o and of EDU_OUTPUT.
type Mode string

// Supported modes.
const (
	Table Mode = "table"
	JSON  Mode = "json"
)

// ParseMode validates a -o value.
func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case Table, JSON:
		return Mode(s), nil
	case "":
		return Table, nil
	}
	return "", fmt.Errorf("invalid output %q; expected table or json", s)
}

// Column describes one table column over rows of type T.
type Column[T any] struct {
	Header string
	Value  func(T) string
}

// Print renders rows. Table mode prints a header line and one line per
// row. JSON mode marshals rows verbatim as an array; a nil slice renders
// as [].
func Print[T any](w io.Writer, mode Mode, rows []T, columns []Column[T]) error {
	if mode == JSON {
		if rows == nil {
			rows = []T{}
		}
		return PrintJSON(w, rows)
	}
	return printTable(w, rows, columns)
}

// PrintOne renders a single value. JSON mode marshals v as an object;
// table mode prints a one-row table.
func PrintOne[T any](w io.Writer, mode Mode, v T, columns []Column[T]) error {
	if mode == JSON {
		return PrintJSON(w, v)
	}
	return printTable(w, []T{v}, columns)
}

// PrintJSON writes v as indented JSON followed by a newline.
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printTable[T any](w io.Writer, rows []T, columns []Column[T]) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = strings.ToUpper(c.Header)
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	cells := make([]string, len(columns))
	for _, row := range rows {
		for i, c := range columns {
			cells[i] = c.Value(row)
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	return tw.Flush()
}
