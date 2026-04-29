package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// OutputFormat is the on-the-wire choice for `--output / -o`. Default is
// `table` (human-readable, fixed columns) — agents pass `json` for the
// canonical envelope defined in envelope.go.
type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
)

// ValidOutputFormats returns the values cobra uses for ValidArgs / shell
// completion of the `--output` flag.
func ValidOutputFormats() []string {
	return []string{string(OutputTable), string(OutputJSON)}
}

// ParseOutputFormat normalises and validates a user-supplied `--output`
// value. Empty defaults to `table` (matches the SPA's "human view" default
// and what `olares-cli files ls` already does).
func ParseOutputFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "table":
		return OutputTable, nil
	case "json":
		return OutputJSON, nil
	default:
		return "", fmt.Errorf("unknown output format %q (valid: table, json)", s)
	}
}

// ----------------------------------------------------------------------------
// Renderers
// ----------------------------------------------------------------------------

// WriteJSON marshals env as a single-line JSON document terminated by `\n`.
// Used for both one-shot output and individual NDJSON lines in `--watch`
// mode (the iteration / Error fields handle the per-line state).
func WriteJSON(w io.Writer, env Envelope) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(env)
}

// TableColumn names a table column and how to extract its value. Each leaf
// command supplies its own []TableColumn so the renderer stays agnostic.
type TableColumn struct {
	Header string
	// Get pulls the cell value out of an Item — typically by reading
	// item.Display[<key>]. Should never return nil; render "-" instead.
	Get func(Item) string
}

// WriteTable emits a tabwriter-based table for items. Empty inputs emit a
// single "-" row so the user has a visible signal that the call succeeded
// but produced nothing.
//
// Header / footer matching the SPA aesthetic (two-space gutter, no border
// chars) is intentional — bash piping (`| awk '{print $2}'`) stays simple.
func WriteTable(w io.Writer, columns []TableColumn, items []Item) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.Header
	}
	if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
		return err
	}

	if len(items) == 0 {
		dash := make([]string, len(columns))
		for i := range dash {
			dash[i] = "-"
		}
		if _, err := fmt.Fprintln(tw, strings.Join(dash, "\t")); err != nil {
			return err
		}
		return tw.Flush()
	}

	for _, it := range items {
		row := make([]string, len(columns))
		for i, c := range columns {
			row[i] = c.Get(it)
		}
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// DisplayString is a small helper for table column getters: pull `key` out
// of item.Display and stringify it, falling back to "-" when missing or
// empty. Centralises the rendering of `nil` / "" / 0-length so callers stay
// declarative.
func DisplayString(it Item, key string) string {
	if it.Display == nil {
		return "-"
	}
	v, ok := it.Display[key]
	if !ok || v == nil {
		return "-"
	}
	switch x := v.(type) {
	case string:
		if x == "" {
			return "-"
		}
		return x
	default:
		return fmt.Sprintf("%v", v)
	}
}

// HeadItems truncates items to at most `n`. n<=0 means "no truncation".
// Mirrors --head's "first-N rows" semantics on top of any sort order the
// command established.
func HeadItems(items []Item, n int) []Item {
	if n <= 0 || n >= len(items) {
		return items
	}
	return items[:n]
}
