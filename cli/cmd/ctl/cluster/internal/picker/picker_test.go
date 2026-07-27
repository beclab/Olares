package picker

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	entries := []Entry{
		{Namespace: "user-space-a", Pod: "web-abc", Container: "nginx"},
		{Namespace: "user-space-a", Pod: "web-abc", Container: "sidecar"},
		{Namespace: "system", Pod: "db-xyz", Container: "postgres"},
	}

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"empty returns all", "", 3},
		{"case-insensitive namespace", "USER-SPACE", 2},
		{"container substring", "nginx", 1},
		{"pod substring", "db-", 1},
		{"cross-field slash", "system/db-xyz/postgres", 1},
		{"whitespace trimmed", "  postgres  ", 1},
		{"no match", "zzz", 0},
		// Substring, not subsequence: scattered chars must NOT match.
		{"subsequence container rejected", "ngx", 0},
		{"subsequence namespace rejected", "usrspc", 0},
		// Multi-token AND, order-independent.
		{"two tokens same entry", "system postgres", 1},
		{"tokens reversed", "postgres db", 1},
		{"token missing excludes", "nginx postgres", 0},
		{"partial substring token", "user post", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(entries, tc.query)
			if len(got) != tc.want {
				t.Fatalf("Filter(%q) = %d entries, want %d", tc.query, len(got), tc.want)
			}
		})
	}
}

func TestFilterEmptyReturnsSameSlice(t *testing.T) {
	entries := []Entry{{Namespace: "a", Pod: "b", Container: "c"}}
	if got := Filter(entries, "   "); len(got) != 1 {
		t.Fatalf("blank query should return all, got %d", len(got))
	}
}

func TestFilterRanksBoundaryFirst(t *testing.T) {
	// Both contain "app" as a substring, but a word-boundary hit (start of the
	// pod segment) should rank above a mid-word hit.
	entries := []Entry{
		{Namespace: "ns", Pod: "webapp-1", Container: "c"},   // "app" mid-word
		{Namespace: "ns", Pod: "app-server", Container: "c"}, // "app" at boundary
	}
	got := Filter(entries, "app")
	if len(got) != 2 {
		t.Fatalf("both contain 'app' substring, got %d", len(got))
	}
	if got[0].Pod != "app-server" {
		t.Fatalf("boundary hit should rank first, got %q", got[0].Pod)
	}
}

func TestMatchScoreTokenOrderIndependent(t *testing.T) {
	h := "user-space-yyhtest201/olares-app-deployment-xyz/olares-app"
	s1, ok1 := matchScore(h, "olares app")
	s2, ok2 := matchScore(h, "app olares")
	if !ok1 || !ok2 {
		t.Fatalf("both token orders should match: %v %v", ok1, ok2)
	}
	if s1 != s2 {
		t.Fatalf("score should be order-independent, got %d vs %d", s1, s2)
	}
}

func TestSort_RunningFirstThenAlpha(t *testing.T) {
	entries := []Entry{
		{Namespace: "b", Pod: "p", Container: "c1", Running: false},
		{Namespace: "a", Pod: "p", Container: "c2", Running: true},
		{Namespace: "a", Pod: "p", Container: "c1", Running: true},
		{Namespace: "a", Pod: "p", Container: "c0", Running: false},
	}
	Sort(entries)

	// Running entries come first, each group sorted by ns/pod/container.
	want := []Entry{
		{Namespace: "a", Pod: "p", Container: "c1", Running: true},
		{Namespace: "a", Pod: "p", Container: "c2", Running: true},
		{Namespace: "a", Pod: "p", Container: "c0", Running: false},
		{Namespace: "b", Pod: "p", Container: "c1", Running: false},
	}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("Sort mismatch:\n got %+v\nwant %+v", entries, want)
	}
}

func TestWindow(t *testing.T) {
	tests := []struct {
		name                   string
		n, cursor, height      int
		wantStart, wantEnd     int
	}{
		{"height covers all", 5, 3, 10, 0, 5},
		{"empty list", 0, 0, 10, 0, 0},
		{"zero height", 5, 2, 0, 0, 0},
		{"cursor at top", 10, 0, 4, 0, 4},
		{"cursor centered", 10, 5, 4, 3, 7},
		{"cursor at bottom clamps", 10, 9, 4, 6, 10},
		{"cursor near bottom", 10, 8, 4, 6, 10},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start, end := window(tc.n, tc.cursor, tc.height)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Fatalf("window(%d,%d,%d) = (%d,%d), want (%d,%d)",
					tc.n, tc.cursor, tc.height, start, end, tc.wantStart, tc.wantEnd)
			}
			// Invariant: the window must contain the cursor when non-empty.
			if end > start && (tc.cursor < start || tc.cursor >= end) && tc.height > 0 && tc.n > 0 {
				t.Fatalf("cursor %d not inside window [%d,%d)", tc.cursor, start, end)
			}
		})
	}
}
