package search

import (
	"bytes"
	"strings"
	"testing"
)

func TestWatchResultPrinterTableStreamsWindowInArrivalOrder(t *testing.T) {
	var out, notes bytes.Buffer
	p := newWatchResultPrinter(&out, &notes, FormatTable, 1, 1)

	if err := p.emit(asyncIndexedHit{Source: appFilesV2, Hit: []byte(`{"title":"skip"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := p.emit(asyncIndexedHit{Source: appDropbox, Hit: []byte(`{"title":"report","resource_uri":"dropbox/me/report.pdf"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := p.emit(asyncIndexedHit{Source: appGoogleDrive, Hit: []byte(`{"title":"past-limit"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := p.finish(3); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"1. [dropbox] report", "dropbox/me/report.pdf", "1 of 3 result(s)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "skip") || strings.Contains(got, "past-limit") {
		t.Fatalf("output ignored offset/limit: %q", got)
	}
}

// A window that covered everything must not be dressed up as a partial one.
func TestWatchResultPrinterCompleteWindowReportsPlainCount(t *testing.T) {
	var out, notes bytes.Buffer
	p := newWatchResultPrinter(&out, &notes, FormatTable, 0, 20)

	for _, title := range []string{"one", "two"} {
		if err := p.emit(asyncIndexedHit{Source: appFilesV2, Hit: []byte(`{"title":"` + title + `"}`)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.finish(2); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, "\n2 result(s)\n") {
		t.Fatalf("output %q does not end with a plain count", got)
	}
	if strings.Contains(got, "--limit") {
		t.Fatalf("output suggests paging a complete result set: %q", got)
	}
}

// The point of the asynchronous search is that the job already ran; with no
// --limit, everything it produced has to reach the terminal.
func TestWatchResultPrinterUnlimitedStreamsEverything(t *testing.T) {
	var out, notes bytes.Buffer
	p := newWatchResultPrinter(&out, &notes, FormatTable, 0, 0)

	for _, title := range []string{"one", "two", "three", "four", "five"} {
		if err := p.emit(asyncIndexedHit{Source: appFilesV2, Hit: []byte(`{"title":"` + title + `"}`)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.finish(5); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"1. [files_v2] one", "5. [files_v2] five", "\n5 result(s)\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "--limit") {
		t.Fatalf("an unlimited run must not suggest paging: %q", got)
	}
}

// The job's own total can only be trusted as a floor: whatever streamed is
// already proof the result set is at least that big.
func TestWatchResultPrinterCountNeverUndercutsStreamedHits(t *testing.T) {
	var out, notes bytes.Buffer
	p := newWatchResultPrinter(&out, &notes, FormatTable, 0, 1)

	for _, title := range []string{"one", "two", "three"} {
		if err := p.emit(asyncIndexedHit{Source: appFilesV2, Hit: []byte(`{"title":"` + title + `"}`)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.finish(0); err != nil {
		t.Fatal(err)
	}

	if got := out.String(); !strings.Contains(got, "1 of 3 result(s)") {
		t.Fatalf("output %q does not report the streamed total", got)
	}
}

func TestWatchResultPrinterJSONUsesJSONLines(t *testing.T) {
	var out, notes bytes.Buffer
	p := newWatchResultPrinter(&out, &notes, FormatJSON, 0, 2)

	for _, title := range []string{"one", "two"} {
		if err := p.emit(asyncIndexedHit{Source: appFilesV2, Hit: []byte(`{"title":"` + title + `"}`)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.finish(5); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("JSONL lines = %d, want 2; output=%q", len(lines), out.String())
	}
	if lines[0] != `{"title":"one"}` || lines[1] != `{"title":"two"}` {
		t.Fatalf("unexpected JSONL output: %q", out.String())
	}
	if !strings.Contains(notes.String(), "printed 2 of 5 result(s)") {
		t.Fatalf("note = %q, want it to report the truncated total", notes.String())
	}
}

func TestWatchResultPrinterEmpty(t *testing.T) {
	var out, notes bytes.Buffer
	p := newWatchResultPrinter(&out, &notes, FormatTable, 0, 20)
	if err := p.finish(0); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "no results\n" {
		t.Fatalf("output = %q, want no results", got)
	}
}
