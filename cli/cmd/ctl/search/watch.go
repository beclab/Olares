package search

import (
	"encoding/json"
	"fmt"
	"io"
)

// watchResultPrinter streams unique asynchronous hits in arrival order. This
// deliberately does not repaint the terminal: output remains useful when
// redirected to a file or piped into another command.
//
// notes carries the trailing count in JSON mode, where stdout has to stay a
// clean JSONL stream.
type watchResultPrinter struct {
	w       io.Writer
	notes   io.Writer
	format  Format
	offset  int
	limit   int
	seen    int
	printed int
}

func newWatchResultPrinter(w, notes io.Writer, format Format, offset, limit int) *watchResultPrinter {
	return &watchResultPrinter{w: w, notes: notes, format: format, offset: offset, limit: limit}
}

func (p *watchResultPrinter) emit(hit asyncIndexedHit) error {
	p.seen++
	if p.seen <= p.offset || (p.limit > 0 && p.printed >= p.limit) {
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

	if _, err := fmt.Fprintf(p.w, "%d. [%s] %s\n", p.printed, hit.Source, item.displayTitle()); err != nil {
		return err
	}
	if loc := item.locationLine(); loc != "" {
		if _, err := fmt.Fprintf(p.w, "   %s\n", loc); err != nil {
			return err
		}
	}
	if snippet := item.snippet(); snippet != "" {
		if _, err := fmt.Fprintf(p.w, "   %s\n", snippet); err != nil {
			return err
		}
	}
	return nil
}

// finish closes the stream with the count. total is the size of the job's full
// result set, which only the terminal frame knows; the hits that actually
// streamed are a floor under it.
func (p *watchResultPrinter) finish(total int) error {
	if total < p.seen {
		total = p.seen
	}
	remaining := remainingResults(total, p.offset, p.printed)
	windowed := p.offset > 0 || p.limit > 0
	if p.format == FormatJSON {
		return writeTruncationNote(p.notes, p.printed, total, remaining, windowed)
	}
	if p.printed == 0 {
		_, err := fmt.Fprintln(p.w, "no results")
		return err
	}
	return writeResultCount(p.w, p.printed, total, remaining, windowed)
}
