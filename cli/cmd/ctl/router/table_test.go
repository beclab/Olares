package router

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

var errWriteFailed = errors.New("the destination refused the write")

// Forty call sites constructed a tabwriter with the same six arguments, and the
// two that drifted are the reason this guard exists rather than a comment: a
// table built with different padding lines up differently from every other
// table the same command prints, and nobody reviewing one file sees the other
// thirty-nine.
func TestNoTabwriterIsConstructedOutsideTableGo(t *testing.T) {
	for _, path := range packageFiles(t) {
		base := filepath.Base(path)
		if base == "table.go" || strings.HasSuffix(base, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "tabwriter" {
				return true
			}
			t.Errorf("%s: tabwriter.%s is reached here; newTable in table.go owns the column layout",
				fset.Position(sel.Pos()), sel.Sel.Name)
			return true
		})
	}
}

func TestTableAlignsColumnsUnderTheHeader(t *testing.T) {
	var buf bytes.Buffer
	tbl := newTable(&buf, "NAME", "MODE")
	tbl.row("a-very-long-model-name", "chat")
	tbl.row("short", "embedding")
	if err := tbl.flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want a header and two rows:\n%s", len(lines), buf.String())
	}
	// Every cell in a column starts at the same offset, which is the only thing
	// a caller gets from this that plain Fprintf would not give it.
	want := strings.Index(lines[0], "MODE")
	for i, line := range lines[1:] {
		if got := strings.Index(line, strings.Fields(line)[1]); got != want {
			t.Errorf("row %d: second column starts at %d, header has it at %d", i, got, want)
		}
	}
}

// A detail block has no header: the left cell of every row is its own label,
// and a header would only repeat it.
func TestTableWithoutColumnsWritesNoHeader(t *testing.T) {
	var buf bytes.Buffer
	tbl := newTable(&buf)
	tbl.row("STATUS", "ready")
	if err := tbl.flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if lines := strings.Count(buf.String(), "\n"); lines != 1 {
		t.Errorf("got %d lines, want only the row:\n%s", lines, buf.String())
	}
}

// The first write error is what flush answers with, and the rows after it are
// dropped rather than retried — a caller that checked after every row behaved
// the same way, having returned at the first failure.
func TestTableReportsTheFirstWriteError(t *testing.T) {
	fw := &failingWriter{}
	tbl := newTable(fw, "NAME")
	tbl.row("one")
	tbl.row("two")
	if err := tbl.flush(); err == nil {
		t.Fatal("flush succeeded on a writer that cannot be written to")
	}
}

func TestPageFooterIsSilentOnACompleteList(t *testing.T) {
	cases := []struct {
		name                 string
		shown, total, offset int
		want                 string
	}{
		{"whole list on one page", 3, 3, 0, ""},
		{"last page", 2, 12, 10, ""},
		{"more to come", 10, 25, 0, "\nshowing 1-10 of 25; --offset 10 for the next page\n"},
		{"more to come, mid-list", 10, 25, 10, "\nshowing 11-20 of 25; --offset 20 for the next page\n"},
		// A filter can narrow the total below what a later page would start at.
		{"total behind the offset", 0, 4, 10, ""},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		if err := pageFooter(&buf, c.shown, c.total, c.offset); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := buf.String(); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

type failingWriter struct{}

func (*failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}
