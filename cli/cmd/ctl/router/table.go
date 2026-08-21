package router

// One column layout, and one sentence about the rows a page left out.
//
// Every table in this tree wants the same thing — tab-separated cells aligned
// into columns — and got it by constructing a tabwriter with the same six
// arguments in forty places, then checking an error after each row. The check
// is not idle: tabwriter forwards a write when it meets a line that cannot
// affect a column width. But it is the same check forty times over, and
// writing it inline meant a row loop was four lines of plumbing around one
// line of content.
//
// table keeps the first error and answers with it at flush, which is the
// behaviour the inline version had and the only observable moment a caller
// cares about. What it buys is that adding a column is editing a list of
// strings.
//
// Cells are strings by the time they arrive. Most already were, passing
// through nonEmpty, boolStr or clip; the few that were numbers now say how
// they want to be spelled at the call site, where the unit ("14ms", "80%")
// is a fact about the column rather than about the writer.

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type table struct {
	tw  *tabwriter.Writer
	err error
}

// newTable starts a table, writing a header row when given column names.
// Called with none it is a two-column field-and-value block, which is how the
// detail views are laid out — there the left column is the label, so a header
// would only repeat what each row already says.
func newTable(dest io.Writer, columns ...string) *table {
	t := &table{tw: tabwriter.NewWriter(dest, 0, 4, 2, ' ', 0)}
	if len(columns) > 0 {
		t.row(columns...)
	}
	return t
}

func (t *table) row(cells ...string) {
	if t.err != nil {
		return
	}
	if _, err := io.WriteString(t.tw, strings.Join(cells, "\t")+"\n"); err != nil {
		t.err = err
	}
}

func (t *table) flush() error {
	if t.err != nil {
		return t.err
	}
	return t.tw.Flush()
}

// pageFooter says what a page left out, and how to ask for the rest.
//
// Silent when there is no rest: a footer under a complete list would have the
// reader looking for rows that do not exist. shown is how many rows this page
// held, which with the offset is the only way to know where the next one
// starts — total alone cannot say, because a filter narrowed both.
func pageFooter(dest io.Writer, shown, total, offset int) error {
	end := offset + shown
	if total <= end {
		return nil
	}
	_, err := fmt.Fprintf(dest, "\nshowing %d-%d of %d; --offset %d for the next page\n",
		offset+1, end, total, end)
	return err
}
