package files

import (
	"strings"
	"testing"
)

// TestParseChownUID locks in the --uid argument validator: integer
// inputs must round-trip, non-integers and negatives must fail with
// a self-describing error that tells the user what LarePass sends.
func TestParseChownUID(t *testing.T) {
	type tc struct {
		in        string
		expectVal int
		expectErr bool
		// errSubs (when expectErr) MUST appear in the error so the
		// user gets a path to recovery without re-reading docs.
		errSubs []string
	}
	cases := []tc{
		{in: "0", expectVal: 0},
		{in: "1000", expectVal: 1000},
		{in: "  42 ", expectVal: 42},
		// Empty: the cobra layer would never call us here in
		// practice (uidSet would be false), but defense in depth.
		{in: "", expectErr: true, errSubs: []string{"--uid is empty", "1000"}},
		// Garbage: surface the LarePass presets so the user has
		// a concrete fallback.
		{in: "user", expectErr: true, errSubs: []string{"--uid", "1000"}},
		{in: "1k", expectErr: true, errSubs: []string{"--uid"}},
		// Negative: reject client-side; the server casts to uint
		// and -1 would silently become a huge UID.
		{in: "-1", expectErr: true, errSubs: []string{"non-negative"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := parseChownUID(c.in)
			if c.expectErr {
				if err == nil {
					t.Fatalf("parseChownUID(%q) = %d, want error", c.in, got)
				}
				for _, s := range c.errSubs {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("parseChownUID(%q) error %q missing substring %q",
							c.in, err.Error(), s)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("parseChownUID(%q): unexpected error: %v", c.in, err)
			}
			if got != c.expectVal {
				t.Errorf("parseChownUID(%q) = %d, want %d", c.in, got, c.expectVal)
			}
		})
	}
}

// TestPrettifyUID confirms the human-readable annotation we tack onto
// the LarePass preset UIDs in stdout. Non-presets get an empty string
// so the regular case stays terse.
func TestPrettifyUID(t *testing.T) {
	cases := []struct {
		uid    int
		expect string
	}{
		{0, " (Root)"},
		{1000, " (User)"},
		{42, ""},
		{-1, ""},
	}
	for _, c := range cases {
		if got := prettifyUID(c.uid); got != c.expect {
			t.Errorf("prettifyUID(%d) = %q, want %q", c.uid, got, c.expect)
		}
	}
}

// TestFrontendPathToChownTarget covers every CLI-side gate:
//   - allowed namespaces parse cleanly,
//   - unsupported namespaces hit the per-namespace messages,
//   - volume root is refused with a self-describing error.
//
// The trailing-slash → IsDirIntent mapping is also exercised so a
// future change to that semantic doesn't silently regress.
func TestFrontendPathToChownTarget(t *testing.T) {
	t.Run("allows drive Home file", func(t *testing.T) {
		got, err := frontendPathToChownTarget("drive/Home/Documents/foo.pdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FileType != "drive" || got.Extend != "Home" || got.SubPath != "/Documents/foo.pdf" {
			t.Errorf("got %+v", got)
		}
		if got.IsDirIntent {
			t.Errorf("IsDirIntent = true for non-trailing-slash path")
		}
	})
	t.Run("allows drive Data dir with trailing slash", func(t *testing.T) {
		got, err := frontendPathToChownTarget("drive/Data/dir/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.IsDirIntent {
			t.Errorf("IsDirIntent = false for trailing-slash path")
		}
	})
	t.Run("allows cache deep path", func(t *testing.T) {
		got, err := frontendPathToChownTarget("cache/node-a/scratch/build/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FileType != "cache" || got.Extend != "node-a" {
			t.Errorf("got %+v", got)
		}
	})

	rejects := []struct {
		name    string
		raw     string
		errSubs []string
	}{
		{
			name:    "rejects sync namespace",
			raw:     "sync/repo-id-abc/foo.txt",
			errSubs: []string{"sync", "files repos"},
		},
		{
			name:    "rejects external namespace",
			raw:     "external/node-a/hdd1/foo",
			errSubs: []string{"external", "Permission tab", "drive", "cache"},
		},
		{
			name:    "rejects awss3 cloud",
			raw:     "awss3/account/bucket/key",
			errSubs: []string{"awss3", "POSIX uid", "drive"},
		},
		{
			name:    "rejects google cloud",
			raw:     "google/account/folder/file",
			errSubs: []string{"google", "POSIX uid"},
		},
		{
			name:    "rejects dropbox cloud",
			raw:     "dropbox/account/folder",
			errSubs: []string{"dropbox", "POSIX uid"},
		},
		{
			name:    "rejects tencent cloud",
			raw:     "tencent/account/bucket",
			errSubs: []string{"tencent", "POSIX uid"},
		},
		{
			name:    "refuses drive Home root",
			raw:     "drive/Home/",
			errSubs: []string{"refusing", "root", "drive/Home"},
		},
		{
			name:    "refuses drive Data root",
			raw:     "drive/Data",
			errSubs: []string{"refusing", "root", "drive/Data"},
		},
		{
			name:    "refuses cache node root",
			raw:     "cache/node-a/",
			errSubs: []string{"refusing", "root", "cache/node-a"},
		},
	}
	for _, c := range rejects {
		t.Run(c.name, func(t *testing.T) {
			_, err := frontendPathToChownTarget(c.raw)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", c.raw)
			}
			for _, s := range c.errSubs {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q missing substring %q", err.Error(), s)
				}
			}
		})
	}
}

// TestChownNamespaceError pins the per-namespace recovery hints so
// a regression that drops the recovery path from any branch is
// caught here rather than in a confused user's bug report.
func TestChownNamespaceError(t *testing.T) {
	cases := []struct {
		ft    string
		subs  []string
	}{
		{"sync", []string{"sync", "files repos"}},
		{"external", []string{"external", "Permission tab"}},
		{"awss3", []string{"awss3", "object stores"}},
		{"dropbox", []string{"dropbox", "object stores"}},
		{"google", []string{"google", "object stores"}},
		{"tencent", []string{"tencent", "object stores"}},
		// Generic / unknown fileType fallback — still must mention
		// the allow-list so the user can act.
		{"share", []string{"share", "drive", "cache"}},
	}
	for _, c := range cases {
		t.Run(c.ft, func(t *testing.T) {
			err := chownNamespaceError(c.ft)
			if err == nil {
				t.Fatalf("expected error for %q", c.ft)
			}
			for _, s := range c.subs {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q missing substring %q", err.Error(), s)
				}
			}
		})
	}
}

// TestNewChownCommand_FlagWiring covers the cobra-layer flag plumbing:
// --uid is a string (so we can distinguish "not passed" from "passed
// 0"), -r is a boolean, and -R is a hidden alias that flips -r in
// PreRunE.
func TestNewChownCommand_FlagWiring(t *testing.T) {
	cmd := NewChownCommand(nil)

	// --uid present, with a default the help can render.
	uidFlag := cmd.Flags().Lookup("uid")
	if uidFlag == nil {
		t.Fatal("--uid flag not registered")
	}
	if uidFlag.Value.Type() != "string" {
		t.Errorf("--uid type = %q, want string", uidFlag.Value.Type())
	}
	// -r aliased correctly.
	rFlag := cmd.Flags().Lookup("recursive")
	if rFlag == nil {
		t.Fatal("--recursive flag not registered")
	}
	if rFlag.Shorthand != "r" {
		t.Errorf("--recursive shorthand = %q, want r", rFlag.Shorthand)
	}
	// -R hidden alias.
	rBSD := cmd.Flags().Lookup("recursive-bsd")
	if rBSD == nil {
		t.Fatal("--recursive-bsd alias not registered")
	}
	if !rBSD.Hidden {
		t.Errorf("--recursive-bsd should be hidden")
	}
	if rBSD.Shorthand != "R" {
		t.Errorf("--recursive-bsd shorthand = %q, want R", rBSD.Shorthand)
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}
	if !strings.Contains(cmd.Long, "drive/Home") {
		t.Error("Long help should describe drive/Home support")
	}
	if !strings.Contains(cmd.Long, "GET") || !strings.Contains(cmd.Long, "PUT") {
		t.Error("Long help should describe GET vs PUT modes")
	}
}
