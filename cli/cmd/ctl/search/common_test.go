package search

import (
	"encoding/json"
	"strconv"
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
