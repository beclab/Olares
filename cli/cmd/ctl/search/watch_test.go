package search

import (
	"bytes"
	"strings"
	"testing"
)

func TestWatchResultPrinterTableStreamsWindowInArrivalOrder(t *testing.T) {
	var out bytes.Buffer
	p := newWatchResultPrinter(&out, FormatTable, 1, 1)

	if err := p.emit(asyncIndexedHit{Source: appFilesV2, Hit: []byte(`{"title":"skip"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := p.emit(asyncIndexedHit{Source: appDropbox, Hit: []byte(`{"title":"report","resource_uri":"dropbox/me/report.pdf"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := p.emit(asyncIndexedHit{Source: appGoogleDrive, Hit: []byte(`{"title":"past-limit"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := p.finish(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"1. [dropbox] report", "dropbox/me/report.pdf", "1 result(s)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "skip") || strings.Contains(got, "past-limit") {
		t.Fatalf("output ignored offset/limit: %q", got)
	}
}

func TestWatchResultPrinterJSONUsesJSONLines(t *testing.T) {
	var out bytes.Buffer
	p := newWatchResultPrinter(&out, FormatJSON, 0, 2)

	for _, title := range []string{"one", "two"} {
		if err := p.emit(asyncIndexedHit{Source: appFilesV2, Hit: []byte(`{"title":"` + title + `"}`)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.finish(); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("JSONL lines = %d, want 2; output=%q", len(lines), out.String())
	}
	if lines[0] != `{"title":"one"}` || lines[1] != `{"title":"two"}` {
		t.Fatalf("unexpected JSONL output: %q", out.String())
	}
}

func TestWatchResultPrinterEmpty(t *testing.T) {
	var out bytes.Buffer
	p := newWatchResultPrinter(&out, FormatTable, 0, 20)
	if err := p.finish(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "no results\n" {
		t.Fatalf("output = %q, want no results", got)
	}
}
