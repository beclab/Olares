package files

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/mkdir"
)

// TestFrontendPathToMkdirTarget covers the cobra-layer adapter that
// turns a user-supplied 3-segment path into the mkdir package's
// Target. The volume-root refusal is enforced both here and in
// mkdir.Plan; this test pins the cobra-side error message (which
// includes a "(e.g. drive/Home/NewFolder)" CTA the planner alone
// can't produce since it doesn't have the parsed FrontendPath).
func TestFrontendPathToMkdirTarget(t *testing.T) {
	t.Run("happy path: full subPath round-trips", func(t *testing.T) {
		tgt, err := frontendPathToMkdirTarget("drive/Home/Documents/Backups")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if tgt.FileType != "drive" || tgt.Extend != "Home" {
			t.Errorf("FileType/Extend wrong: %+v", tgt)
		}
		if tgt.SubPath != "/Documents/Backups" {
			t.Errorf("SubPath: got %q, want %q", tgt.SubPath, "/Documents/Backups")
		}
	})

	t.Run("trailing slash on input is preserved on SubPath", func(t *testing.T) {
		tgt, err := frontendPathToMkdirTarget("drive/Home/Documents/")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !strings.HasSuffix(tgt.SubPath, "/") {
			t.Errorf("trailing slash dropped: %q", tgt.SubPath)
		}
	})

	t.Run("volume root rejected with friendly CTA", func(t *testing.T) {
		_, err := frontendPathToMkdirTarget("drive/Home/")
		if err == nil {
			t.Fatal("want error, got nil")
		}
		// CTA should mention the namespace + a sample subdirectory.
		// Without that hint, the user gets a useless "refusing to
		// mkdir the root" line and has to guess what shape the path
		// should take.
		if !strings.Contains(err.Error(), "drive/Home/NewFolder") {
			t.Errorf("err should suggest a sample path, got: %v", err)
		}
	})

	t.Run("malformed front-end path bubbles up", func(t *testing.T) {
		// Empty input is the most common "I forgot the path" mistake;
		// the FrontendPath parser owns the error message but the
		// adapter must propagate it.
		if _, err := frontendPathToMkdirTarget(""); err == nil {
			t.Error("want error for empty path")
		}
	})

	// Regression: `path.Clean` inside ParseFrontendPath silently
	// collapses `.` / `..` segments, which would otherwise let
	// `mkdir drive/Home/.` slip through as a "create the root"
	// attempt (misleading error) and `mkdir drive/Home/foo/./bar`
	// silently land at `drive/Home/foo/bar`. The pre-check in
	// frontendPathToMkdirTarget must fire BEFORE that cleanup so the
	// user sees a targeted blacklist error pointing at the offending
	// segment.
	t.Run("dot-segment blacklist fires before path.Clean", func(t *testing.T) {
		cases := []struct {
			name string
			in   string
			seg  string
		}{
			{"leaf is '.'", "drive/Home/.", `"."`},
			{"leaf is '..'", "drive/Home/..", `".."`},
			{"interior './'", "drive/Home/foo/./bar", `"."`},
			{"interior '../'", "drive/Home/foo/../bar", `".."`},
			{"leaf with trailing slash '.' ", "drive/Home/Documents/./", `"."`},
			{"all-traversal subpath", "drive/Home/../../etc", `".."`},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				_, err := frontendPathToMkdirTarget(c.in)
				if err == nil {
					t.Fatalf("want error for %q", c.in)
				}
				// Error must mention the offending segment AND the
				// blacklist phrase so the user understands what's
				// being rejected. Without the segment in the
				// message a user with a deep path can't easily
				// locate the typo.
				if !strings.Contains(err.Error(), c.seg) {
					t.Errorf("err %q should mention segment %s", err.Error(), c.seg)
				}
				if !strings.Contains(err.Error(), "path-traversal blacklist") {
					t.Errorf("err %q should mention 'path-traversal blacklist'", err.Error())
				}
			})
		}
	})
}

// TestLastSegmentForHint locks in the basename derivation used by
// the post-mkdir auto-rename hint. SubPath always starts with '/';
// trailing '/' is tolerated.
func TestLastSegmentForHint(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/Documents", "Documents"},
		{"/Documents/", "Documents"},
		{"/A/B/C", "C"},
		{"/A/B/C/", "C"},
		{"/", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := lastSegmentForHint(c.in); got != c.want {
			t.Errorf("lastSegmentForHint(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParentDisplayPath locks in the parent-path derivation used by
// the post-mkdir `files ls` CTA. The volume root falls back to
// `<fileType>/<extend>/` so the suggestion always points at a
// listable path.
func TestParentDisplayPath(t *testing.T) {
	cases := []struct {
		in   mkdir.Target
		want string
	}{
		{mkdir.Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/Backups"}, "drive/Home/Documents/"},
		{mkdir.Target{FileType: "drive", Extend: "Home", SubPath: "/Solo"}, "drive/Home/"},
		{mkdir.Target{FileType: "drive", Extend: "Home", SubPath: "/Solo/"}, "drive/Home/"},
		{mkdir.Target{FileType: "sync", Extend: "abc", SubPath: "/notes/2026/Q2"}, "sync/abc/notes/2026/"},
	}
	for _, c := range cases {
		if got := parentDisplayPath(c.in); got != c.want {
			t.Errorf("parentDisplayPath(%+v) = %q, want %q", c.in, got, c.want)
		}
	}
}
