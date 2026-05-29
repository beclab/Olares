// share_preflight_test.go: unit tests for the cobra-layer share-
// create dir-only / existence preflight (preflightShareCreate in
// share_create.go). Mirrors the cp/rm preflight test pattern: stand
// up an httptest.Server, drive download.Client.Stat against it, and
// assert that preflightShareCreate refuses the right shapes with
// the right error messages.
//
// The preflight is what enforces the LarePass GUI's per-driver
// share-menu gating on event.isDir (sharing a single file is
// rejected by both the web app and the CLI). The shape of the gate
// — Stat the target, refuse if !IsDir — is shared across all three
// flavors (internal / public / smb); the only per-flavor variance
// is the friendly flavor name baked into the error message, so the
// tests parameterize over share.Type to lock that in.
package files

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/internal/files/share"
)

// newSharePreflightClient stands up an httptest.Server with the
// supplied handler and returns a download.Client pointing at it.
// Same shape as newPreflightClient in cp_test.go — the share
// preflight reuses download.Client.Stat verbatim, so a dedicated
// helper would just be a copy; keeping a tiny share-specific
// wrapper isolates this file's intent (share-create gating).
func newSharePreflightClient(t *testing.T, h http.Handler) *download.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &download.Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}
}

// shareListingHandler is the share-preflight equivalent of
// cp_test.go's listingHandler. Routes are keyed by the GET URL
// path (e.g. `/api/resources/drive/Home/`) and the value is the
// raw `{"items":[...]}` body to write back. Unknown paths return
// 404 so the preflight's NotFound branch fires. The handler also
// asserts that the preflight only issues GETs (a POST against the
// /api/share/ surface would mean we accidentally let the create
// call leak past the gate — that's exactly what the preflight is
// supposed to PREVENT, so we want the test to scream).
func shareListingHandler(t *testing.T, routes map[string]string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("share preflight should only issue GETs, got %s %s",
				r.Method, r.URL.Path)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, ok := routes[r.URL.Path]
		if !ok {
			http.Error(w, "{}", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, body)
	})
}

// TestPreflightShareCreate_DirectoryHappyPath: target is an
// existing directory under drive/Home/. The preflight should pass
// without error after exactly one GET (the parent listing that
// resolves the directory leaf).
//
// Runs across all three flavors to confirm the dir-only gate is
// uniformly applied (the per-flavor namespace allow-list lives in
// a separate gate; this test only exercises the existence + kind
// branch).
func TestPreflightShareCreate_DirectoryHappyPath(t *testing.T) {
	for _, flavor := range []share.Type{
		share.TypeInternal,
		share.TypePublic,
		share.TypeSMB,
	} {
		t.Run(string(flavor), func(t *testing.T) {
			stat := newSharePreflightClient(t, shareListingHandler(t, map[string]string{
				"/api/resources/drive/Home/": `{"items":[
					{"name":"Backups","isDir":true,"size":0}
				]}`,
			}))
			tgt := share.Target{
				FileType:    "drive",
				Extend:      "Home",
				SubPath:     "/Backups/",
				IsDirIntent: true,
			}
			if err := preflightShareCreate(context.Background(), stat, tgt, flavor); err != nil {
				t.Errorf("directory happy path (%s): unexpected error %v", flavor, err)
			}
		})
	}
}

// TestPreflightShareCreate_RejectsFile: the central regression
// guard. The target resolves to a file on the server; preflight
// must refuse with a message that:
//
//   - names the offending path verbatim;
//   - says "is a file on the server";
//   - mentions "only supports directories";
//   - cites the GUI gating ("event.isDir") so a maintainer doing
//     code-archaeology can find this rule in the LarePass app
//     source from the error message alone;
//   - hints at the corrective workflow ("place it in a dedicated
//     directory").
//
// Runs across all three flavors to lock down the per-flavor
// friendly-name prefix ("internal" / "public" / "smb") — the wire
// type for Public is the historically confusing "external", so
// users would hunt for a non-existent verb if the error leaked
// the wire value.
func TestPreflightShareCreate_RejectsFile(t *testing.T) {
	cases := []struct {
		flavor       share.Type
		wantFriendly string
	}{
		{share.TypeInternal, "internal"},
		{share.TypePublic, "public"},
		{share.TypeSMB, "smb"},
	}
	for _, c := range cases {
		t.Run(string(c.flavor), func(t *testing.T) {
			stat := newSharePreflightClient(t, shareListingHandler(t, map[string]string{
				"/api/resources/drive/Home/": `{"items":[
					{"name":"notes.md","isDir":false,"size":42}
				]}`,
			}))
			tgt := share.Target{
				FileType: "drive",
				Extend:   "Home",
				SubPath:  "/notes.md",
			}
			err := preflightShareCreate(context.Background(), stat, tgt, c.flavor)
			if err == nil {
				t.Fatalf("flavor=%s: expected refusal for file target", c.flavor)
			}
			msg := err.Error()
			for _, want := range []string{
				"refusing to create a " + c.wantFriendly + " share",
				"drive/Home/notes.md",
				"is a file on the server",
				"only supports directories",
				"event.isDir",
				"place it in a dedicated directory",
			} {
				if !strings.Contains(msg, want) {
					t.Errorf("flavor=%s: error must contain %q; got: %v",
						c.flavor, want, err)
				}
			}
		})
	}
}

// TestPreflightShareCreate_RejectsMissingTarget: the target path
// doesn't exist on the server (parent listing exists but the leaf
// isn't there). The preflight must refuse with a "does not exist"
// message that names the offending path. This is the typo / stale-
// path branch — same shape as cp/rm's preflight, just routed
// through the share-flavor friendly-name prefix.
func TestPreflightShareCreate_RejectsMissingTarget(t *testing.T) {
	stat := newSharePreflightClient(t, shareListingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"other","isDir":true,"size":0}
		]}`,
	}))
	tgt := share.Target{
		FileType:    "drive",
		Extend:      "Home",
		SubPath:     "/Backups/",
		IsDirIntent: true,
	}
	err := preflightShareCreate(context.Background(), stat, tgt, share.TypeInternal)
	if err == nil {
		t.Fatal("expected refusal for missing target")
	}
	msg := err.Error()
	for _, want := range []string{
		"refusing to create a internal share",
		"drive/Home/Backups/",
		"does not exist on the server",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error must contain %q; got: %v", want, err)
		}
	}
}

// TestPreflightShareCreate_VolumeRootIsSyntheticDir: a share
// against the bare volume root (e.g. `drive/Home/`,
// `drive/Data/`, `sync/<repo>/`) must pass without hitting the
// wire. download.Stat short-circuits ≤2-segment paths to a
// synthetic IsDir=true record (see internal/files/download/stat.go).
//
// The handler asserts that ZERO GETs reach it — a regression here
// would mean we're paying an extra round-trip for sharing a
// volume root (and it would 404 against the test server since no
// routes are mounted).
func TestPreflightShareCreate_VolumeRootIsSyntheticDir(t *testing.T) {
	cases := []struct {
		name   string
		target share.Target
	}{
		{
			name:   "drive/Home/",
			target: share.Target{FileType: "drive", Extend: "Home", SubPath: "/", IsDirIntent: true},
		},
		{
			name:   "drive/Data/",
			target: share.Target{FileType: "drive", Extend: "Data", SubPath: "/", IsDirIntent: true},
		},
		{
			name: "sync/<repo>/",
			target: share.Target{
				FileType:    "sync",
				Extend:      "b7ffab7f-3ceb-4e36-aeb7-74d958ad0a7a",
				SubPath:     "/",
				IsDirIntent: true,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var hits int
			stat := newSharePreflightClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				t.Errorf("volume-root preflight should not hit the wire; got %s %s",
					r.Method, r.URL.Path)
				http.Error(w, "{}", http.StatusInternalServerError)
			}))
			if err := preflightShareCreate(context.Background(), stat, c.target, share.TypeInternal); err != nil {
				t.Errorf("volume root preflight: unexpected error %v", err)
			}
			if hits != 0 {
				t.Errorf("expected zero wire GETs for synthetic volume root, got %d", hits)
			}
		})
	}
}

// TestPreflightShareCreate_DeepDirectory: target is a deeper
// subdirectory (`drive/Home/Reports/2026/Q1/`) — the preflight
// must resolve through the parent listing
// (`/api/resources/drive/Home/Reports/2026/`) and accept when the
// leaf is a dir. Confirms the parent-listing strategy works at
// depth, not just the depth-1 case the happy-path test covers.
func TestPreflightShareCreate_DeepDirectory(t *testing.T) {
	stat := newSharePreflightClient(t, shareListingHandler(t, map[string]string{
		"/api/resources/drive/Home/Reports/2026/": `{"items":[
			{"name":"Q1","isDir":true,"size":0}
		]}`,
	}))
	tgt := share.Target{
		FileType:    "drive",
		Extend:      "Home",
		SubPath:     "/Reports/2026/Q1/",
		IsDirIntent: true,
	}
	if err := preflightShareCreate(context.Background(), stat, tgt, share.TypePublic); err != nil {
		t.Errorf("deep dir preflight: unexpected error %v", err)
	}
}

// TestPreflightShareCreate_DeepFileRejected: deeper variant of
// the file-rejection test — a real path like
// `drive/Home/Reports/2026/Q1.pdf`. Locks in that the dir-only
// gate fires at any depth, not just at the volume-root level.
func TestPreflightShareCreate_DeepFileRejected(t *testing.T) {
	stat := newSharePreflightClient(t, shareListingHandler(t, map[string]string{
		"/api/resources/drive/Home/Reports/2026/": `{"items":[
			{"name":"Q1.pdf","isDir":false,"size":1024}
		]}`,
	}))
	tgt := share.Target{
		FileType: "drive",
		Extend:   "Home",
		SubPath:  "/Reports/2026/Q1.pdf",
	}
	err := preflightShareCreate(context.Background(), stat, tgt, share.TypeSMB)
	if err == nil {
		t.Fatal("expected refusal for deep file target")
	}
	if !strings.Contains(err.Error(), "drive/Home/Reports/2026/Q1.pdf") {
		t.Errorf("error must name the deep file path; got: %v", err)
	}
	if !strings.Contains(err.Error(), "is a file on the server") {
		t.Errorf("error must say 'is a file on the server'; got: %v", err)
	}
}
