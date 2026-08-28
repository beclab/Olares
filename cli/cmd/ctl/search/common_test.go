package search

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func TestPaginateRaw(t *testing.T) {
	t.Parallel()

	// Each row encodes its own index so the test can assert the window is
	// sliced at the right offset, not merely that the count matches.
	rows := func(n int) []json.RawMessage {
		out := make([]json.RawMessage, n)
		for i := range out {
			out[i] = json.RawMessage(strconv.Itoa(i))
		}
		return out
	}

	cases := []struct {
		name      string
		total     int
		offset    int
		limit     int
		wantFirst int // index of the first returned row; -1 when empty
		wantLen   int
	}{
		{"limit smaller than total", 20, 0, 10, 0, 10},
		{"limit larger than total", 5, 0, 50, 0, 5},
		{"offset past end", 5, 10, 20, -1, 0},
		{"offset at end", 5, 5, 20, -1, 0},
		{"offset with remaining window", 30, 20, 20, 20, 10},
		{"offset plus limit within total", 30, 5, 10, 5, 10},
		{"empty input", 0, 0, 20, -1, 0},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := paginateRaw(rows(tc.total), tc.offset, tc.limit)
			if len(got) != tc.wantLen {
				t.Fatalf("paginateRaw(total=%d, offset=%d, limit=%d) = %d rows, want %d",
					tc.total, tc.offset, tc.limit, len(got), tc.wantLen)
			}
			for k, raw := range got {
				wantIdx := tc.wantFirst + k
				var gotIdx int
				if err := json.Unmarshal(raw, &gotIdx); err != nil {
					t.Fatalf("row %d: decode %s: %v", k, raw, err)
				}
				if gotIdx != wantIdx {
					t.Fatalf("paginateRaw(total=%d, offset=%d, limit=%d): row %d = index %d, want %d",
						tc.total, tc.offset, tc.limit, k, gotIdx, wantIdx)
				}
			}
		})
	}
}

func TestNeedsMorePage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		offset  int
		limit   int
		initLen int
		want    bool
	}{
		{"first page within init", 0, 20, 20, false},
		{"small first page", 0, 5, 20, false},
		{"window inside full init page", 5, 10, 20, false},
		{"window touches init edge", 10, 10, 20, false},
		{"limit beyond full init page", 0, 50, 20, true},
		{"offset beyond full init page", 25, 10, 20, true},
		{"short init page is exhaustive", 0, 50, 8, false},
		{"offset past short page stays local", 10, 20, 8, false},
		{"empty results never page", 0, 50, 0, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := needsMorePage(tc.offset, tc.limit, tc.initLen); got != tc.want {
				t.Fatalf("needsMorePage(offset=%d, limit=%d, initLen=%d) = %v, want %v",
					tc.offset, tc.limit, tc.initLen, got, tc.want)
			}
		})
	}
}

func TestResultItemLocationLine(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		row  string
		want string
	}{
		{
			"federated seafile hit names its library",
			`{"resource_uri":"/sync/2c768a94/asd/","meta":{"repo_name":"My Library"}}`,
			"/sync/2c768a94/asd/ (My Library)",
		},
		{
			"legacy sync row carries repo_name at the top level",
			`{"path":"/asd/","repo_name":"My Library"}`,
			"/asd/ (My Library)",
		},
		{
			"drive hit has no library to name",
			`{"resource_uri":"drive/Home/report.pdf"}`,
			"drive/Home/report.pdf",
		},
		{
			"seafile hit without a library name stays usable",
			`{"resource_uri":"/sync/2c768a94/asd/","meta":{}}`,
			"/sync/2c768a94/asd/",
		},
		{
			// `meta` is source-specific, so an unmodelled shape must degrade to
			// "no library name" instead of failing the row.
			"unexpected meta shape is ignored",
			`{"resource_uri":"drive/Home/report.pdf","meta":[]}`,
			"drive/Home/report.pdf",
		},
		{"row without a location prints nothing", `{"title":"asd"}`, ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			items, err := decodeResultRows([]json.RawMessage{json.RawMessage(tc.row)})
			if err != nil {
				t.Fatalf("decodeResultRows(%s): %v", tc.row, err)
			}
			if got := items[0].locationLine(); got != tc.want {
				t.Fatalf("locationLine() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRemainingResults(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                 string
		total, offset, shown int
		want                 int
	}{
		{"default window over a larger set", 57, 0, 20, 37},
		{"second page", 57, 20, 20, 17},
		{"last page", 57, 40, 17, 0},
		{"window covers everything", 12, 0, 12, 0},
		{"unknown total", 0, 0, 20, 0},
		{"offset past the end", 57, 100, 0, 0},
	}
	for _, tc := range cases {
		if got := remainingResults(tc.total, tc.offset, tc.shown); got != tc.want {
			t.Fatalf("%s: remainingResults(%d, %d, %d) = %d, want %d",
				tc.name, tc.total, tc.offset, tc.shown, got, tc.want)
		}
	}
}

// A default --limit against a bigger result set is exactly the case that made
// a CLI run look like it disagreed with the Desktop dialog, so the footer has
// to name the full set.
func TestRenderResultsNamesTheFullResultSet(t *testing.T) {
	t.Parallel()

	items, err := decodeResultRows([]json.RawMessage{json.RawMessage(`{"title":"report"}`)})
	if err != nil {
		t.Fatal(err)
	}

	var partial bytes.Buffer
	if err := renderResults(&partial, searchPage{items: items, total: 57, windowed: true}); err != nil {
		t.Fatal(err)
	}
	if got := partial.String(); !strings.Contains(got, "1 of 57 result(s)") || !strings.Contains(got, "--limit") {
		t.Fatalf("partial window footer = %q", got)
	}

	// Nothing was held back on purpose here, so pointing at --limit would send
	// the reader after results that no flag can produce.
	var undelivered bytes.Buffer
	if err := renderResults(&undelivered, searchPage{items: items, total: 57}); err != nil {
		t.Fatal(err)
	}
	if got := undelivered.String(); !strings.Contains(got, "reported 56 more but never delivered") ||
		strings.Contains(got, "--limit") {
		t.Fatalf("undelivered footer = %q", got)
	}

	var complete bytes.Buffer
	if err := renderResults(&complete, searchPage{items: items, total: 1}); err != nil {
		t.Fatal(err)
	}
	if got := complete.String(); !strings.Contains(got, "\n1 result(s)\n") || strings.Contains(got, "--limit") {
		t.Fatalf("complete window footer = %q", got)
	}

	// The legacy session API cannot report a total; the footer must not invent one.
	var unknown bytes.Buffer
	if err := renderResults(&unknown, searchPage{items: items}); err != nil {
		t.Fatal(err)
	}
	if got := unknown.String(); !strings.Contains(got, "\n1 result(s)\n") {
		t.Fatalf("unknown-total footer = %q", got)
	}
}

func TestAsyncTotalHits(t *testing.T) {
	t.Parallel()

	sources := []string{appFilesV2, appSeafile}
	summaries := map[string]asyncSourceSummary{
		appFilesV2: {Status: "completed", HitCount: 40},
		appSeafile: {Status: "completed", HitCount: 17},
		// A source the caller did not ask for must not inflate the total.
		appDropbox: {Status: "completed", HitCount: 99},
	}
	if got := asyncTotalHits(sources, summaries, 57); got != 57 {
		t.Fatalf("asyncTotalHits = %d, want 57", got)
	}
	// A backend that omits the summary still owes an honest count.
	if got := asyncTotalHits(sources, nil, 12); got != 12 {
		t.Fatalf("asyncTotalHits without summary = %d, want 12", got)
	}
	// Recovery may fall short of what the job claimed; report what arrived.
	if got := asyncTotalHits(sources, summaries, 50); got != 57 {
		t.Fatalf("asyncTotalHits with a gap = %d, want 57", got)
	}
}

// The rows here are the shapes `search drive` actually printed against a
// 1.12.7 backend, including the content-only matches that came out as
// "(untitled)" with the excerpt marker showing through.
func TestResultItemTitleAndSnippet(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		row         string
		wantTitle   string
		wantSnippet string
	}{
		{
			"filename match does not repeat the title in the excerpt",
			`{"title":"test(1).py","resource_uri":"drive/Common/myxtest/test(1).py",
			  "highlight":["<hi>test</hi>(1).py","print('<hi>test</hi>')"],
			  "highlight_field":["title","content"]}`,
			"test(1).py",
			"print('test')",
		},
		{
			"content-only match recovers its name from the location",
			`{"title":"","resource_uri":"drive/Common/myxtest/market&setting-GPU 的 olares-cli 复测(1).xlsx",
			  "highlight":["running install - market.<hi>test</hi> sglangllmbasev32755#######opped stop"],
			  "highlight_field":["content"]}`,
			"market&setting-GPU 的 olares-cli 复测(1).xlsx",
			"running install - market.test sglangllmbasev32755 … opped stop",
		},
		{
			"title excerpt names a row the backend left untitled",
			`{"title":"","resource_uri":"drive/Home/notes.md",
			  "highlight":["<hi>notes</hi>.md"],"highlight_field":"title"}`,
			"notes.md",
			"",
		},
		{
			"directory hit drops the trailing slash",
			`{"title":"","resource_uri":"/sync/2c768a94/myxtest/"}`,
			"myxtest",
			"",
		},
		{
			// Rows predating highlight_field must keep rendering their excerpt.
			"row without highlight_field joins what it has",
			`{"title":"report.pdf","highlight":["a <hi>test</hi>","another <hi>test</hi>"]}`,
			"report.pdf",
			"a test … another test",
		},
		{
			"row with nothing to show",
			`{"title":""}`,
			"(untitled)",
			"",
		},
		{
			"unmodelled highlight shape is ignored",
			`{"title":"report.pdf","highlight":{"content":"x"},"highlight_field":["content"]}`,
			"report.pdf",
			"",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			items, err := decodeResultRows([]json.RawMessage{json.RawMessage(tc.row)})
			if err != nil {
				t.Fatalf("decodeResultRows: %v", err)
			}
			if got := items[0].displayTitle(); got != tc.wantTitle {
				t.Fatalf("displayTitle() = %q, want %q", got, tc.wantTitle)
			}
			if got := items[0].snippet(); got != tc.wantSnippet {
				t.Fatalf("snippet() = %q, want %q", got, tc.wantSnippet)
			}
		})
	}
}

// The elision marker is matched at its exact width, so hashes that are really
// part of a document survive.
func TestCleanHighlightKeepsGenuineHashes(t *testing.T) {
	t.Parallel()

	cases := []struct{ in, want string }{
		{"## Attachment Test #######  Different words", "## Attachment Test … Different words"},
		{"###### heading six", "###### heading six"},
		{"######## eight hashes", "######## eight hashes"},
		{"#######", ""},
	}
	for _, tc := range cases {
		if got := cleanHighlight(tc.in); got != tc.want {
			t.Fatalf("cleanHighlight(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestClampMoreLimit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in, want int
	}{
		{1, 1},
		{20, 20},
		{100, 100},
		{101, 100},
		{500, 100},
	}
	for _, tc := range cases {
		if got := clampMoreLimit(tc.in); got != tc.want {
			t.Fatalf("clampMoreLimit(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
