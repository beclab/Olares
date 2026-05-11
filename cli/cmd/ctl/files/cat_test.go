package files

import "testing"

// TestIsCloudDriveType locks in the set of namespaces that cat
// dispatches to /drive/download_sync_stream rather than /api/raw/.
// Adding a new cloud-bridge-backed namespace (or removing one)
// should be an obvious code-review signal — that's why the predicate
// + test live next to the cat command rather than as a generic
// FrontendPath helper.
func TestIsCloudDriveType(t *testing.T) {
	cases := []struct {
		fileType string
		want     bool
	}{
		{"awss3", true},
		{"google", true},
		{"dropbox", true},
		// Tencent uses the same /drive/download_sync_stream endpoint
		// for downloads (see TencentDataAPI / utils.ts in the web
		// app), even though its UPLOAD path is the octet
		// /drive/direct_upload_file flow that the CLI doesn't yet
		// support. Cat is therefore safe to enable here.
		{"tencent", true},

		{"drive", false},
		{"sync", false},
		{"cache", false},
		{"external", false},
		{"share", false},
		{"", false},
		{"unknown", false},
	}
	for _, c := range cases {
		if got := isCloudDriveType(c.fileType); got != c.want {
			t.Errorf("isCloudDriveType(%q) = %v, want %v", c.fileType, got, c.want)
		}
	}
}
