// cp_test.go: unit tests for the cobra-layer cp/mv glue —
// specifically the preflight existence/kind checks that run BEFORE
// any state-changing PATCH /api/paste/<node>/ goes out.
//
// Why this lives in the cobra layer rather than internal/files/cp:
// the planner (cp.Plan) is pure — it operates on already-typed
// Source/Destination shapes and never touches the wire. The
// preflight, by contrast, is a network step (one Stat per source +
// one Stat for the destination side) that the cobra glue
// orchestrates by wiring a download.Client into preflightCpMv. The
// tests follow the same parent-listing mock pattern the download
// package's stat tests use (see internal/files/download/stat.go).
package files

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/cp"
	"github.com/beclab/Olares/cli/internal/files/download"
)

// newPreflightClient stands up an httptest.Server with the supplied
// handler and returns a download.Client pointing at it. The cp
// preflight only consumes Stat, so a single handler is enough — no
// need to instantiate the cp.Client.
func newPreflightClient(t *testing.T, h http.Handler) *download.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &download.Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}
}

// listingHandler is a tiny test helper: routes are keyed by
// `<URL.Path>` (e.g. `/api/resources/drive/Home/`) and the value is
// the raw JSON body to write back. Unknown paths return 404 so the
// preflight's NotFound branch fires; entries shape mirrors the
// LarePass web app's items[] envelope (the cp preflight reuses
// download.Client.Stat which decodes that shape).
func listingHandler(t *testing.T, routes map[string]string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("preflight should only issue GETs, got %s %s",
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

// TestParentSubPath pins the small helper that derives the parent
// directory of a `/...`-prefixed subpath. The cases cover the three
// shapes that show up on the wire: nested file, nested directory
// (input itself ending in `/`), depth-1 leaf (parent is the extend
// root), and the extend root itself (its own parent — degenerate
// guard so callers never have to special-case it).
func TestParentSubPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"nested file", "/Documents/foo.pdf", "/Documents/"},
		{"nested dir with trailing slash", "/Documents/sub/", "/Documents/"},
		{"depth-1 file", "/foo.pdf", "/"},
		{"depth-1 dir", "/Backups/", "/"},
		{"extend root", "/", "/"},
		// Defensive: an empty input should NOT panic. The cp glue
		// always passes a `/`-prefixed string from ParseFrontendPath
		// so this is unreachable in practice, but the helper guards
		// it explicitly and we lock that down.
		{"empty input", "", "/"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parentSubPath(tc.in); got != tc.want {
				t.Errorf("parentSubPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestPreflightCpMv_HappyPath: every source exists with the right
// kind, the destination is an existing directory. The handler
// returns the parent listings that make all those lookups succeed.
//
// `cp drive/Home/notes.md drive/Home/Archive/` — single file source
// under `drive/Home/`, dropped into the existing `drive/Home/Archive/`
// directory. The preflight should issue exactly two GETs:
//
//   - GET /api/resources/drive/Home/      → find "notes.md" (file) and "Archive" (dir)
//   - GET /api/resources/drive/Home/      → find "Archive" (dir) for the dst stat
//
// We mount both lookups on the same `/api/resources/drive/Home/`
// URL so the handler returns the same body twice (download.Stat
// doesn't cache across calls — that's a documented future
// optimization, not the current contract).
func TestPreflightCpMv_HappyPath(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42},
			{"name":"Archive","isDir":true,"size":0}
		]}`,
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/Archive/", IsDirIntent: true,
	}
	if err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy); err != nil {
		t.Fatalf("happy-path preflight: unexpected error %v", err)
	}
}

// TestPreflightCpMv_SourceNotFound covers the most common
// user-facing failure: a typo in the src path. The handler omits
// the leaf from the parent listing, so download.Stat synthesises a
// 404 and the preflight should turn it into "source ... does not
// exist on the server" with the offending path named verbatim.
func TestPreflightCpMv_SourceNotFound(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"other.md","isDir":false,"size":1}
		]}`,
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/", IsDirIntent: true,
	}
	err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if !strings.Contains(err.Error(), "drive/Home/notes.md") ||
		!strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should name the missing source and say 'does not exist', got: %v", err)
	}
}

// TestPreflightCpMv_SrcKind_DirNoTrailingSlash: the user typed
// `cp drive/Home/Photos drive/Home/Archive/` (no trailing slash on
// Photos), but Photos is actually a directory on the server. The
// planner wouldn't catch this — it trusts the user's IsDirIntent
// signal — so the preflight surfaces the kind mismatch with a
// targeted message that includes the corrective command shape
// ("add a trailing '/' and pass -r").
func TestPreflightCpMv_SrcKind_DirNoTrailingSlash(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"Photos","isDir":true,"size":0},
			{"name":"Archive","isDir":true,"size":0}
		]}`,
	}))
	srcs := []cp.Source{
		// IsDirIntent=false on purpose: user did NOT type trailing slash.
		{FileType: "drive", Extend: "Home", SubPath: "/Photos"},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/Archive/", IsDirIntent: true,
	}
	err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy)
	if err == nil {
		t.Fatal("expected error for dir-source-without-trailing-slash")
	}
	if !strings.Contains(err.Error(), "is a directory") ||
		!strings.Contains(err.Error(), "Photos") ||
		!strings.Contains(err.Error(), "-r") {
		t.Errorf("error should name the dir, mention '-r', got: %v", err)
	}
}

// TestPreflightCpMv_SrcKind_FileWithTrailingSlash: inverse of the
// above. The user typed `cp -r drive/Home/notes.md/ drive/Home/Archive/`
// (trailing slash on what is actually a file). The planner would
// happily proceed (the slash + -r flag pass the IsDirIntent +
// recursive guard), and the backend would reject it confusingly.
// The preflight catches it with "is a file on the server, not a
// directory; drop the trailing '/'".
func TestPreflightCpMv_SrcKind_FileWithTrailingSlash(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42},
			{"name":"Archive","isDir":true,"size":0}
		]}`,
	}))
	srcs := []cp.Source{
		// IsDirIntent=true (user typed `notes.md/`).
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md/", IsDirIntent: true},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/Archive/", IsDirIntent: true,
	}
	err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy)
	if err == nil {
		t.Fatal("expected error for file-source-with-trailing-slash")
	}
	if !strings.Contains(err.Error(), "is a file") ||
		!strings.Contains(err.Error(), "notes.md") ||
		!strings.Contains(err.Error(), "drop the trailing") {
		t.Errorf("error should name the file and tell the user to drop the slash, got: %v", err)
	}
}

// TestPreflightCpMv_DstDirMissing covers the "you forgot to mkdir
// the target" case. drop-into-dir mode REQUIRES the dst directory
// to already exist on the server (the backend's POST-on-collision
// auto-rename quirk would otherwise create weird sibling shapes —
// see SKILL.md's "Server-side quirks" section). The preflight
// refuses early with a `mkdir` CTA.
func TestPreflightCpMv_DstDirMissing(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42}
		]}`,
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		// `Archive` is NOT in the parent listing above.
		FileType: "drive", Extend: "Home", SubPath: "/Archive/", IsDirIntent: true,
	}
	err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy)
	if err == nil {
		t.Fatal("expected error for missing dst dir")
	}
	if !strings.Contains(err.Error(), "drive/Home/Archive/") ||
		!strings.Contains(err.Error(), "does not exist") ||
		!strings.Contains(err.Error(), "mkdir") {
		t.Errorf("error should name the missing dir + suggest mkdir, got: %v", err)
	}
}

// TestPreflightCpMv_DstIsFile catches `cp foo bar/` where `bar` is
// actually a file on the server (not a directory). Drop-into-dir
// mode would land foo at `bar/foo`, which the backend can't
// represent because bar is a file. The preflight refuses with
// "destination ... is a file ... drop the trailing '/'".
func TestPreflightCpMv_DstIsFile(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42},
			{"name":"bar","isDir":false,"size":7}
		]}`,
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/bar/", IsDirIntent: true,
	}
	err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy)
	if err == nil {
		t.Fatal("expected error for dst-that-is-file")
	}
	if !strings.Contains(err.Error(), "bar") ||
		!strings.Contains(err.Error(), "is a file") {
		t.Errorf("error should name the file dst, got: %v", err)
	}
}

// TestPreflightCpMv_VolumeRootDst: dst is the bare volume root
// (`drive/Home/`). download.Stat short-circuits to a synthetic dir
// for ≤2-segment paths so the preflight should pass WITHOUT hitting
// the wire for the dst leg — the handler asserts that the only GET
// we see is the source-side parent listing.
func TestPreflightCpMv_VolumeRootDst(t *testing.T) {
	var dstStats int
	stat := newPreflightClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("want GET, got %s", r.Method)
		}
		// The only legitimate GET is the source's parent listing.
		// Any other URL means download.Stat hit the network for
		// the synthetic-dir volume root — a regression.
		if r.URL.Path != "/api/resources/drive/Home/" {
			t.Errorf("unexpected wire call %s — synthetic dst stat should not hit the network",
				r.URL.Path)
		}
		dstStats++
		_, _ = io.WriteString(w, `{"items":[
			{"name":"notes.md","isDir":false,"size":42}
		]}`)
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/", IsDirIntent: true,
	}
	if err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy); err != nil {
		t.Fatalf("volume-root dst: unexpected error %v", err)
	}
	// Exactly one GET for the source's parent listing — no
	// extra round-trip for the synthetic dst.
	if dstStats != 1 {
		t.Errorf("want 1 wire GET (source parent), got %d", dstStats)
	}
}

// TestPreflightCpMv_ExactTargetParentExists covers the undocumented
// single-source + non-trailing-slash `<dst>` mode that the planner
// still accepts for backwards compatibility. The preflight should
// verify the PARENT of dst (not dst itself) — the leaf doesn't exist
// yet by definition. Happy path: parent exists as a dir.
func TestPreflightCpMv_ExactTargetParentExists(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		// Source's parent — contains notes.md.
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42},
			{"name":"Documents","isDir":true,"size":0}
		]}`,
		// Dst's parent — `/Documents/`. The leaf
		// `notes-2026.md` is intentionally absent: exact-target
		// mode tolerates that (the backend will create it or
		// auto-rename on collision).
		"/api/resources/drive/Home/Documents/": `{"items":[]}`,
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		// IsDirIntent=false: this is the undocumented exact-target shape.
		FileType: "drive", Extend: "Home", SubPath: "/Documents/notes-2026.md",
	}
	if err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy); err != nil {
		t.Fatalf("exact-target happy path: unexpected error %v", err)
	}
}

// TestPreflightCpMv_ExactTargetParentMissing: exact-target mode
// where the dst's parent dir is itself a typo. Without this check
// the backend would either 404 or create a weird placeholder; the
// preflight refuses early and names the missing parent.
func TestPreflightCpMv_ExactTargetParentMissing(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		// Source's parent — notes.md is here.
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42}
		]}`,
		// Dst's parent (`/Documents/`) is intentionally NOT in
		// the routes map → returns 404 → preflight should refuse.
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/Documents/notes-2026.md",
	}
	err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionCopy)
	if err == nil {
		t.Fatal("expected error for missing exact-target parent")
	}
	if !strings.Contains(err.Error(), "parent") ||
		!strings.Contains(err.Error(), "Documents") {
		t.Errorf("error should mention the missing parent dir, got: %v", err)
	}
}

// TestPreflightCpMv_MvActionLabelsErrors: the error messages should
// say `mv:` (not `copy:`) when action is move, so the user can match
// the failure to the command they actually typed. Cheap to verify
// alongside one of the existing rejection arms.
func TestPreflightCpMv_MvActionLabelsErrors(t *testing.T) {
	stat := newPreflightClient(t, listingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[]}`,
	}))
	srcs := []cp.Source{
		{FileType: "drive", Extend: "Home", SubPath: "/notes.md"},
	}
	dst := cp.Destination{
		FileType: "drive", Extend: "Home", SubPath: "/", IsDirIntent: true,
	}
	err := preflightCpMv(context.Background(), stat, srcs, dst, cp.ActionMove)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if !strings.HasPrefix(err.Error(), "move:") {
		t.Errorf("error should be labelled with the action verb, got: %v", err)
	}
}
