package search

import (
	"encoding/json"
	"fmt"
	"io"
)

// watchResultPrinter streams unique asynchronous hits in arrival order. This
// deliberately does not repaint the terminal: output remains useful when
// redirected to a file or piped into another command.
type watchResultPrinter struct {
	w       io.Writer
	format  Format
	offset  int
	limit   int
	seen    int
	printed int
}

func newWatchResultPrinter(w io.Writer, format Format, offset, limit int) *watchResultPrinter {
	return &watchResultPrinter{w: w, format: format, offset: offset, limit: limit}
}

func (p *watchResultPrinter) emit(hit asyncIndexedHit) error {
	p.seen++
	if p.seen <= p.offset || p.printed >= p.limit {
		return nil
	}

	var item resultItem
	if err := json.Unmarshal(hit.Hit, &item); err != nil {
		return fmt.Errorf("decode asynchronous search result: %w", err)
	}
	item.Raw = hit.Hit
	p.printed++

	if p.format == FormatJSON {
		return json.NewEncoder(p.w).Encode(item.Raw)
	}

	title := item.Title
	if title == "" {
		title = "(untitled)"
	}
	if _, err := fmt.Fprintf(p.w, "%d. [%s] %s\n", p.printed, hit.Source, title); err != nil {
		return err
	}
	if loc := item.location(); loc != "" {
		if _, err := fmt.Fprintf(p.w, "   %s\n", loc); err != nil {
			return err
		}
	}
	if snippet := highlightSnippet(item.Highlight); snippet != "" {
		if _, err := fmt.Fprintf(p.w, "   %s\n", snippet); err != nil {
			return err
		}
	}
	return nil
}

func (p *watchResultPrinter) finish() error {
	if p.format == FormatJSON {
		return nil
	}
	if p.printed == 0 {
		_, err := fmt.Fprintln(p.w, "no results")
		return err
	}
	_, err := fmt.Fprintf(p.w, "\n%d result(s)\n", p.printed)
	return err
}
