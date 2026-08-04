package search

import (
	"testing"
)

// TestSessionAppConstants pins the search3 app partition strings that the
// new gdrive / dropbox / knowledge subcommands send as the "app" field on
// /api/search/init. These must stay aligned with TermiPass ServiceType.
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
