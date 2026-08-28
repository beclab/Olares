package search

import (
	"testing"
)

// `--limit 0` means "print every result", which the session API cannot do: it
// pages blind over a server-side cache. Left unresolved it reads as a
// zero-width window -- one that the init page always already satisfies, so a
// truncated first page is served as if it were the whole result set, and that
// asks /search/more for `limit: 0`, outside the 1-100 the backend accepts.
func TestResolveSessionWindowServesUnlimitedAsAPage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		offset, limit int
		initLen       int
		wantOffset    int
		wantLimit     int
		wantNeedsMore bool
	}{
		{"unlimited over a full init page", 0, 0, initPageSize, 0, initPageSize, false},
		{"unlimited over a short init page", 0, 0, 8, 0, initPageSize, false},
		{"unlimited past the init page", 30, 0, initPageSize, 30, initPageSize, true},
		{"explicit window inside the init page", 0, 5, initPageSize, 0, 5, false},
		{"explicit window past the init page", 20, 20, initPageSize, 20, 20, true},
		{"explicit window larger than the init page", 0, 50, initPageSize, 0, 50, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveSessionWindow(&pagingOptions{offset: tc.offset, limit: tc.limit}, tc.initLen)
			if got.offset != tc.wantOffset || got.limit != tc.wantLimit || got.needsMore != tc.wantNeedsMore {
				t.Fatalf("resolveSessionWindow(offset=%d, limit=%d, initLen=%d) = %+v, want {offset:%d limit:%d needsMore:%v}",
					tc.offset, tc.limit, tc.initLen, got, tc.wantOffset, tc.wantLimit, tc.wantNeedsMore)
			}
			if !got.needsMore {
				return
			}
			if sent := clampMoreLimit(got.limit); sent < 1 || sent > moreMaxLimit {
				t.Fatalf("/search/more would be asked for limit=%d, outside the backend's 1-%d range",
					sent, moreMaxLimit)
			}
		})
	}
}

// TestSessionAppConstants pins the search3 source/app wire strings. These must
// stay aligned with TermiPass ServiceType and the federated search protocol.
func TestSessionAppConstants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"files_v2", appFilesV2, "files_v2"},
		{"google_drive", appGoogleDrive, "google_drive"},
		{"dropbox", appDropbox, "dropbox"},
		{"knowledge", appKnowledge, "knowledge"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestSearchSessionAppMinVersion(t *testing.T) {
	t.Parallel()
	if searchSessionAppMinOlaresVersion != "1.12.7" {
		t.Fatalf("searchSessionAppMinOlaresVersion = %q, want 1.12.7", searchSessionAppMinOlaresVersion)
	}
}
