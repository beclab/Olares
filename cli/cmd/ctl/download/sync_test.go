package download

import (
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	// Empty means a full drain (zero time).
	if got, err := parseSince(""); err != nil || !got.IsZero() {
		t.Fatalf("empty: got %v err %v", got, err)
	}

	// A zoned RFC3339 value is honoured exactly (this is how the printed
	// next-cursor round-trips).
	wantUTC := time.Date(2026, 7, 15, 15, 3, 0, 0, time.UTC)
	if got, err := parseSince("2026-07-15T15:03:00Z"); err != nil || !got.Equal(wantUTC) {
		t.Fatalf("rfc3339: got %v err %v", got, err)
	}

	// Zone-less inputs are read in the local timezone, so they match the
	// table's local-time column regardless of the machine's zone.
	wantLocal := time.Date(2026, 7, 15, 23, 3, 0, 0, time.Local)
	for _, in := range []string{
		"2026-07-15T23:03",
		"2026-07-15T23:03:00",
		"2026-07-15 23:03",
		"2026-07-15 23:03:00",
	} {
		got, err := parseSince(in)
		if err != nil {
			t.Fatalf("%q: unexpected err %v", in, err)
		}
		if !got.Equal(wantLocal) {
			t.Fatalf("%q: got %v, want %v", in, got, wantLocal)
		}
	}

	// A bare date is local midnight.
	wantMidnight := time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)
	if got, err := parseSince("2026-07-15"); err != nil || !got.Equal(wantMidnight) {
		t.Fatalf("date: got %v err %v", got, err)
	}

	// Garbage is rejected.
	if _, err := parseSince("not-a-time"); err == nil {
		t.Fatal("expected error for garbage input")
	}
}

func TestShouldPrintNextCursor(t *testing.T) {
	rows := []DownloadTask{{ID: 7, UpdatedAt: time.Now()}}
	cases := []struct {
		name string
		res  SyncResult
		want bool
	}{
		{name: "more with rows hints", res: SyncResult{HasMore: true, Items: rows}, want: true},
		// has_more=true with an empty list must NOT hint: NextCursor would be
		// the zero-value cursor and stall paging.
		{name: "more but empty does not hint", res: SyncResult{HasMore: true, Items: nil}, want: false},
		{name: "no more does not hint", res: SyncResult{HasMore: false, Items: rows}, want: false},
	}
	for _, tc := range cases {
		if got := shouldPrintNextCursor(tc.res); got != tc.want {
			t.Fatalf("%s: shouldPrintNextCursor=%v want %v", tc.name, got, tc.want)
		}
	}
}
