// archive_test.go: client-side gates for the four archive verbs
// (compress / extract / archive entries / archive cat).
//
// The interesting surface here is the namespace allow-list:
// compress / extract / archive entries / archive cat only accept
// `drive/Home`, `drive/Data`, `cache/<node>`, `external/<node>/...`,
// and the parsers must refuse everything else with an actionable,
// per-namespace recovery hint instead of letting the request reach
// the backend and surface as an opaque error.
//
// These tests pin both the allow set (positive cases) and the
// rejection text (negative cases): regressions in either dimension
// are loud rather than subtle.
package files

import (
	"strings"
	"testing"
)

// TestValidateArchiveNamespace pins the allow-list and reject text.
// Splitting positive / negative into one test (subtests) keeps the
// matrix legible — if the allowed set ever changes both halves stay
// nearby in diffs.
func TestValidateArchiveNamespace(t *testing.T) {
	t.Run("allowed", func(t *testing.T) {
		cases := []struct {
			fileType string
			extend   string
		}{
			{"drive", "Home"},
			{"drive", "Data"},
			// Common is the app common data area (Olares >= 1.12.6);
			// TermiPass's archiveSupportedDriveTypes includes it.
			{"drive", "Common"},
			// cache: any extend (=any node) is allowed.
			{"cache", "node-a"},
			{"cache", "olares"},
			{"cache", ""},
			// external: same shape — any extend allowed.
			{"external", "node-a"},
			{"external", "olares"},
			{"external", ""},
		}
		for _, c := range cases {
			c := c
			t.Run(c.fileType+"/"+c.extend, func(t *testing.T) {
				if err := validateArchiveNamespace("compress", c.fileType, c.extend); err != nil {
					t.Errorf("expected %s/%s allowed, got %v", c.fileType, c.extend, err)
				}
			})
		}
	})

	t.Run("rejected", func(t *testing.T) {
		cases := []struct {
			name     string
			fileType string
			extend   string
			subs     []string // substrings the error message must contain
		}{
			{
				name:     "sync rejected with repos hint",
				fileType: "sync",
				extend:   "abc123",
				subs:     []string{"compress", "sync", "Seafile", "files repos"},
			},
			{
				name:     "awss3 rejected with LarePass hint",
				fileType: "awss3",
				extend:   "my-bucket",
				subs:     []string{"compress", "awss3", "cloud", "LarePass"},
			},
			{
				name:     "dropbox rejected",
				fileType: "dropbox",
				extend:   "primary",
				subs:     []string{"dropbox", "cloud"},
			},
			{
				name:     "google rejected",
				fileType: "google",
				extend:   "primary",
				subs:     []string{"google", "cloud"},
			},
			{
				name:     "tencent rejected",
				fileType: "tencent",
				extend:   "primary",
				subs:     []string{"tencent", "cloud"},
			},
			{
				name:     "drive with bad extend rejected",
				fileType: "drive",
				extend:   "Shared",
				subs:     []string{"compress", "drive", "Home", "Data"},
			},
			{
				name:     "unknown fileType rejected with allow-list",
				fileType: "share",
				extend:   "x",
				subs:     []string{"share", "drive/Home", "drive/Data", "cache", "external"},
			},
		}
		for _, c := range cases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				err := validateArchiveNamespace("compress", c.fileType, c.extend)
				if err == nil {
					t.Fatalf("expected error for %s/%s", c.fileType, c.extend)
				}
				for _, s := range c.subs {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("error %q missing substring %q", err.Error(), s)
					}
				}
			})
		}
	})
}

// TestArchiveNamespaceError_VerbAppears makes sure the user-facing
// verb name is rendered into every namespace-error branch — the
// whole point of threading `verb` through validateArchiveNamespace
// is so a multi-command pipeline tells you which command tripped.
func TestArchiveNamespaceError_VerbAppears(t *testing.T) {
	verbs := []string{"compress", "extract", "archive entries", "archive cat"}
	fileTypes := []string{"sync", "awss3", "drive", "share"}
	for _, v := range verbs {
		for _, ft := range fileTypes {
			extend := "Bad" // for drive: triggers the Home/Data branch
			err := archiveNamespaceError(v, ft, extend)
			if err == nil {
				t.Fatalf("expected error: verb=%q fileType=%q", v, ft)
			}
			if !strings.Contains(err.Error(), v) {
				t.Errorf("verb %q missing from error: %q", v, err.Error())
			}
		}
	}
}

// TestParseArchiveSource_NamespaceGate covers the integration:
// parseArchiveSource (used by extract / entries / cat) must funnel
// the rejection through the allow-list before its file/dir-intent
// checks, so the user gets the namespace hint first.
//
// Note: `drive/<bad-extend>` is intentionally NOT exercised here —
// ParseFrontendPath itself already restricts `drive` to Home / Data
// and emits a clear "drive extend must be Home or Data" message,
// so my validator is defence-in-depth only for that case. The
// allow-list still has the drive/Home/Data row for clarity, so a
// future ParseFrontendPath relaxation doesn't accidentally let
// `drive/Trash` slip through.
func TestParseArchiveSource_NamespaceGate(t *testing.T) {
	cases := []struct {
		raw       string
		ok        bool
		verbMatch bool     // require "archive entries" in the error
		subs      []string // other substrings the error must contain
	}{
		// Positive: standard supported namespaces.
		{raw: "drive/Home/Documents/out.zip", ok: true},
		{raw: "drive/Data/dumps/x.tar.gz", ok: true},
		{raw: "cache/node-a/scratch.zip", ok: true},
		{raw: "external/olares/usb1/back.7z", ok: true},

		// Negative: namespace allow-list rejects (my validator).
		{raw: "sync/repo123/notes.zip", ok: false, verbMatch: true, subs: []string{"sync"}},
		{raw: "awss3/primary/x.zip", ok: false, verbMatch: true, subs: []string{"awss3", "cloud"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.raw, func(t *testing.T) {
			_, _, err := parseArchiveSource(c.raw, "archive entries")
			if c.ok {
				if err != nil {
					t.Fatalf("expected ok for %q, got %v", c.raw, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected reject for %q", c.raw)
			}
			for _, s := range c.subs {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q missing substring %q", err.Error(), s)
				}
			}
			if c.verbMatch && !strings.Contains(err.Error(), "archive entries") {
				t.Errorf("error %q missing verb 'archive entries'", err.Error())
			}
		})
	}
}

// TestParseArchiveSource_UpstreamDriveExtendRejection pins the
// upstream rejection path: even though my validator would also
// catch `drive/Trash`, ParseFrontendPath gets there first and
// emits the "Home or Data" hint. Asserting that here keeps the
// regression diagnostic crisp — if FrontendPath's allow-list is
// ever loosened, this test fails before users do.
func TestParseArchiveSource_UpstreamDriveExtendRejection(t *testing.T) {
	_, _, err := parseArchiveSource("drive/Shared/x.zip", "archive entries")
	if err == nil {
		t.Fatal("expected reject for drive/Shared/x.zip")
	}
	for _, s := range []string{"drive", "Home", "Data"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("error %q missing substring %q", err.Error(), s)
		}
	}
}

// TestParseCompressSources_NamespaceGate exercises the multi-source
// guard: even one bad source aborts the batch with the namespace
// rejection BEFORE any wire request goes out.
func TestParseCompressSources_NamespaceGate(t *testing.T) {
	// All sources legal — no error.
	if _, _, err := parseCompressSources([]string{
		"drive/Home/a.pdf",
		"cache/node-a/b.txt",
	}); err != nil {
		t.Errorf("expected ok, got %v", err)
	}

	// One bad source in the middle: the helper rejects on first
	// non-allowed namespace, with the verb threaded into the message.
	_, _, err := parseCompressSources([]string{
		"drive/Home/a.pdf",
		"sync/repo123/b.txt",
		"drive/Data/c.csv",
	})
	if err == nil {
		t.Fatal("expected reject for sync source")
	}
	if !strings.Contains(err.Error(), "compress") {
		t.Errorf("missing verb 'compress' in: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "sync") {
		t.Errorf("missing namespace 'sync' in: %q", err.Error())
	}
}

// TestParseCompressDestination_NamespaceGate guards the destination
// side: writing an archive into sync / cloud is rejected before the
// task is queued.
func TestParseCompressDestination_NamespaceGate(t *testing.T) {
	if _, _, err := parseCompressDestination("drive/Home/out.zip"); err != nil {
		t.Errorf("drive/Home/out.zip should be ok, got %v", err)
	}
	_, _, err := parseCompressDestination("google/primary/out.zip")
	if err == nil {
		t.Fatal("expected reject for google destination")
	}
	if !strings.Contains(err.Error(), "compress") || !strings.Contains(err.Error(), "google") {
		t.Errorf("missing verb/ns in: %q", err.Error())
	}
}

// TestParseExtractDestination_NamespaceGate guards extracting INTO
// an unsupported namespace.
func TestParseExtractDestination_NamespaceGate(t *testing.T) {
	if _, _, err := parseExtractDestination("drive/Home/unpack/"); err != nil {
		t.Errorf("drive/Home/unpack/ should be ok, got %v", err)
	}
	_, _, err := parseExtractDestination("dropbox/primary/unpack/")
	if err == nil {
		t.Fatal("expected reject for dropbox destination")
	}
	if !strings.Contains(err.Error(), "extract") || !strings.Contains(err.Error(), "dropbox") {
		t.Errorf("missing verb/ns in: %q", err.Error())
	}
}

// TestResolveVolumeSizeMB pins the two-flag reconciliation:
// --volume-size (unit-aware) wins, --volume-size-mb is the
// back-compat alias, and passing both is a usage error.
func TestResolveVolumeSizeMB(t *testing.T) {
	cases := []struct {
		name       string
		volumeSize string
		volumeMB   int
		want       int
		err        bool
	}{
		{"neither", "", 0, 0, false},
		{"mb-only", "", 100, 100, false},
		{"size-only-bare", "100", 0, 100, false},
		{"size-only-unit", "1.5GB", 0, 1536, false},
		{"size-sub-mib-floors", "500KB", 0, 1, false},
		{"both-set", "100MB", 50, 0, true},
		{"negative-mb", "", -1, 0, true},
		{"bad-size", "abc", 0, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := resolveVolumeSizeMB(c.volumeSize, c.volumeMB)
			if (err != nil) != c.err {
				t.Fatalf("resolveVolumeSizeMB(%q, %d): err=%v want err=%v", c.volumeSize, c.volumeMB, err, c.err)
			}
			if !c.err && got != c.want {
				t.Errorf("resolveVolumeSizeMB(%q, %d): got %d want %d", c.volumeSize, c.volumeMB, got, c.want)
			}
		})
	}
}
