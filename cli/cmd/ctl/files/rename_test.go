package files

import (
	"strings"
	"testing"
)

// TestFrontendPathToRenameTarget covers the cobra-layer adapter that
// turns the user-supplied source path into the rename package's
// Target. Two concerns are exercised here:
//
//  1. The directory-intent signal — a trailing '/' on the input must
//     round-trip into IsDirIntent / SubPath so the wire URL keeps the
//     directory marker (rename.Plan uses it to route file vs dir
//     handlers).
//
//  2. The `.` / `..` blacklist must fire BEFORE ParseFrontendPath's
//     `path.Clean` step strips those segments silently. Without the
//     pre-check `rename drive/Home/foo/.. newname` would silently
//     surface as "refusing to rename the root" (path.Clean collapses
//     `foo/..` → "/"), and `rename drive/Home/foo/./bar newname`
//     would rename an unintended entry at `drive/Home/foo/bar`.
//
// The new-name leg of the validation is unit-tested in
// internal/files/rename (rename.Plan), so this adapter test focuses
// on the source-path leg the planner cannot see by the time it runs.
func TestFrontendPathToRenameTarget(t *testing.T) {
	t.Run("happy path: file source has no trailing slash", func(t *testing.T) {
		tgt, err := frontendPathToRenameTarget("drive/Home/Documents/foo.pdf")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if tgt.FileType != "drive" || tgt.Extend != "Home" {
			t.Errorf("FileType/Extend wrong: %+v", tgt)
		}
		if tgt.SubPath != "/Documents/foo.pdf" {
			t.Errorf("SubPath: got %q, want %q", tgt.SubPath, "/Documents/foo.pdf")
		}
		if tgt.IsDirIntent {
			t.Errorf("IsDirIntent: want false for a file source")
		}
	})

	t.Run("dir source preserves trailing slash + IsDirIntent", func(t *testing.T) {
		tgt, err := frontendPathToRenameTarget("drive/Home/Pictures/old/")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !strings.HasSuffix(tgt.SubPath, "/") {
			t.Errorf("SubPath should end with '/': %q", tgt.SubPath)
		}
		if !tgt.IsDirIntent {
			t.Errorf("IsDirIntent: want true for a trailing-slash source")
		}
	})

	t.Run("volume root rejected with friendly message", func(t *testing.T) {
		_, err := frontendPathToRenameTarget("drive/Home/")
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "root of drive/Home") {
			t.Errorf("err should mention 'root of drive/Home', got: %v", err)
		}
	})

	// Regression: `path.Clean` inside ParseFrontendPath silently
	// collapses `.` / `..` segments. Without the pre-check the
	// rename source could be rewritten under the user's feet
	// (`drive/Home/foo/../bar` → `drive/Home/bar`, renaming a
	// different entry than the user typed). The blacklist guard in
	// frontendPathToRenameTarget must fire BEFORE that cleanup so
	// the user sees a targeted "path-traversal blacklist" error
	// instead.
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
			{"dir leaf '..' with trailing slash", "drive/Home/Documents/../", `".."`},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				_, err := frontendPathToRenameTarget(c.in)
				if err == nil {
					t.Fatalf("want error for %q", c.in)
				}
				if !strings.Contains(err.Error(), c.seg) {
					t.Errorf("err %q should mention segment %s", err.Error(), c.seg)
				}
				if !strings.Contains(err.Error(), "path-traversal blacklist") {
					t.Errorf("err %q should mention 'path-traversal blacklist'", err.Error())
				}
			})
		}
	})

	t.Run("malformed front-end path bubbles up", func(t *testing.T) {
		if _, err := frontendPathToRenameTarget(""); err == nil {
			t.Error("want error for empty path")
		}
	})
}
