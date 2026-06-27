package search

import (
	"encoding/json"
	"testing"
)

func TestPaginateRaw(t *testing.T) {
	t.Parallel()

	rows := func(n int) []json.RawMessage {
		out := make([]json.RawMessage, n)
		for i := range out {
			out[i] = json.RawMessage(`{}`)
		}
		return out
	}

	cases := []struct {
		name   string
		total  int
		offset int
		limit  int
		want   int
	}{
		{"limit smaller than total", 20, 0, 10, 10},
		{"limit larger than total", 5, 0, 50, 5},
		{"offset past end", 5, 10, 20, 0},
		{"offset at end", 5, 5, 20, 0},
		{"offset with remaining window", 30, 20, 20, 10},
		{"offset plus limit within total", 30, 5, 10, 10},
		{"empty input", 0, 0, 20, 0},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := paginateRaw(rows(tc.total), tc.offset, tc.limit)
			if len(got) != tc.want {
				t.Fatalf("paginateRaw(total=%d, offset=%d, limit=%d) = %d rows, want %d",
					tc.total, tc.offset, tc.limit, len(got), tc.want)
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
