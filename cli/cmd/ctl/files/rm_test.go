// rm_test.go: unit tests for the cobra-layer rm preflight existence
// + kind check. Same shape as cp_test.go — uses httptest to mock
// the parent-listing endpoint download.Client.Stat hits, and asserts
// the error surface on each refusal arm.
//
// Why the test lives at the cobra layer rather than internal/files/rm:
// the planner (rm.Plan) is pure — it operates on already-typed Target
// shapes and never touches the wire. The preflight, by contrast, is
// a network step (one Stat per target) that the cobra glue
// orchestrates by wiring a download.Client into preflightRm. The
// planner tests in internal/files/rm/rm_test.go stay focused on the
// pure validation arms (root refusal, protected names, group
// dedup, ...); preflight semantics live here.
package files

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/internal/files/download"
	"github.com/beclab/Olares/cli/internal/files/rm"
)

// newRmPreflightClient stands up an httptest.Server with the supplied
// handler and returns a download.Client pointing at it. Mirrors the
// cp_test.go helper — the rm preflight only consumes Stat, so a
// single handler suffices (no need to instantiate rm.Client).
func newRmPreflightClient(t *testing.T, h http.Handler) *download.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &download.Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}
}

// rmListingHandler keys routes by URL path (e.g.
// `/api/resources/drive/Home/`) and returns the raw JSON body for
// download.Client.Stat to decode. Unknown paths return 404, which
// download.Stat surfaces as the synthetic "leaf not in listing" 404
// branch — i.e. our preflight's "does not exist" arm.
func rmListingHandler(t *testing.T, routes map[string]string) http.Handler {
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

// TestPreflightRm_HappyPath: target exists with the matching kind.
// One file under `drive/Home/Documents/`, no `-r` — should pass the
// preflight clean without touching anything else.
func TestPreflightRm_HappyPath(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/Documents/": `{"items":[
			{"name":"foo.pdf","isDir":false,"size":1234}
		]}`,
	}))
	targets := []rm.Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Documents/", Name: "foo.pdf"},
	}
	if err := preflightRm(context.Background(), stat, targets, false); err != nil {
		t.Fatalf("happy-path preflight: unexpected error %v", err)
	}
}

// TestPreflightRm_TargetNotFound: the most common typo. Stat's
// parent listing succeeds but the target leaf isn't in it; the
// synthetic 404 maps to "does not exist on the server" with the
// offending path named verbatim (no -r flag in play, so the display
// path stays without a trailing slash).
func TestPreflightRm_TargetNotFound(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"other.md","isDir":false,"size":1}
		]}`,
	}))
	targets := []rm.Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/", Name: "notes.md"},
	}
	err := preflightRm(context.Background(), stat, targets, false)
	if err == nil {
		t.Fatal("expected error for missing target")
	}
	if !strings.Contains(err.Error(), "drive/Home/notes.md") ||
		!strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should name the target and say 'does not exist', got: %v", err)
	}
}

// TestPreflightRm_DirWithoutRecursive: user typed `rm drive/Home/Photos`
// (no trailing slash) and DIDN'T pass -r, but Photos is actually a
// directory on the server. The planner would happily send a
// file-shape dirent `/Photos` which the backend refuses; preflight
// catches it earlier with the same "pass -r/-R" CTA.
func TestPreflightRm_DirWithoutRecursive(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"Photos","isDir":true,"size":0}
		]}`,
	}))
	targets := []rm.Target{
		// IsDirIntent=false: no trailing slash, planner has not seen `-r` yet.
		{FileType: "drive", Extend: "Home", ParentSubPath: "/", Name: "Photos"},
	}
	err := preflightRm(context.Background(), stat, targets, false)
	if err == nil {
		t.Fatal("expected error for dir-target-without-r")
	}
	if !strings.Contains(err.Error(), "Photos") ||
		!strings.Contains(err.Error(), "is a directory") ||
		!strings.Contains(err.Error(), "-r") {
		t.Errorf("error should name the dir + suggest -r, got: %v", err)
	}
}

// TestPreflightRm_FileWithRecursive: inverse of the above. User
// typed `rm -r drive/Home/notes.md` — the planner promotes this to
// a dir-shape dirent `/notes.md/` based on the `-r` flag alone,
// which would hit the backend's directory-removal path against a
// file and fail. The preflight catches the kind mismatch with
// "drop the -r/-R flag".
func TestPreflightRm_FileWithRecursive(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42}
		]}`,
	}))
	targets := []rm.Target{
		// IsDirIntent=false (no trailing slash typed by user), but
		// recursive=true below promotes effective intent to dir.
		{FileType: "drive", Extend: "Home", ParentSubPath: "/", Name: "notes.md"},
	}
	err := preflightRm(context.Background(), stat, targets, true)
	if err == nil {
		t.Fatal("expected error for file-target-with-r")
	}
	if !strings.Contains(err.Error(), "notes.md") ||
		!strings.Contains(err.Error(), "is a file") {
		t.Errorf("error should name the file and reject the recursive intent, got: %v", err)
	}
	// The user did NOT type a trailing slash here — only -r promoted
	// dir intent. Telling them to "drop the trailing '/'" would send
	// them hunting for one in their command line that isn't there.
	// Lock the CTA down to the -r/-R-only branch.
	if !strings.Contains(err.Error(), "drop the -r/-R flag") {
		t.Errorf("CTA must say 'drop the -r/-R flag' when only -r promoted dir intent; got: %v", err)
	}
	if strings.Contains(err.Error(), "trailing '/'") {
		t.Errorf("CTA must NOT mention 'trailing /' when the user didn't type one; got: %v", err)
	}
}

// TestPreflightRm_FileWithTrailingSlash: user typed `rm
// drive/Home/notes.md/` (trailing slash) without -r — but
// IsDirIntent=true triggers the dir-without-recursive planner
// rejection, which is structural and doesn't reach the preflight.
// The interesting case at the preflight layer is `rm -r notes.md/`:
// IsDirIntent=true AND recursive=true, but the target is a file on
// the server. The preflight should refuse with the same "is a file"
// message — directory intent doesn't matter, the actual kind does.
func TestPreflightRm_DirIntentButActuallyFile(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"notes.md","isDir":false,"size":42}
		]}`,
	}))
	targets := []rm.Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/", Name: "notes.md", IsDirIntent: true},
	}
	err := preflightRm(context.Background(), stat, targets, true)
	if err == nil {
		t.Fatal("expected error for dir-intent target that is a file on the server")
	}
	// Display includes the trailing slash (IsDirIntent preserved on the wire path).
	if !strings.Contains(err.Error(), "notes.md/") ||
		!strings.Contains(err.Error(), "is a file") {
		t.Errorf("error should preserve trailing slash on display and call it a file, got: %v", err)
	}
	// Both IsDirIntent AND recursive triggered the effective dir
	// promotion, so the CTA must mention BOTH corrective actions —
	// the user has two flags to undo (or two inputs to reconsider).
	if !strings.Contains(err.Error(), "drop the trailing '/' and the -r/-R flag") {
		t.Errorf("CTA must mention both trailing slash and -r/-R flag when both triggered dir intent; got: %v", err)
	}
}

// TestPreflightRm_RecursiveDirHappyPath: `rm -r drive/Home/Photos/`
// — dir intent (user-typed slash AND -r) matches the actual dir
// kind on the server. The preflight should pass; the planner's
// effective dir-intent rule is in lockstep with the preflight's, so
// `IsDirIntent=true, recursive=true, info.IsDir=true` is the happy
// case.
func TestPreflightRm_RecursiveDirHappyPath(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[
			{"name":"Photos","isDir":true,"size":0}
		]}`,
	}))
	targets := []rm.Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/", Name: "Photos", IsDirIntent: true},
	}
	if err := preflightRm(context.Background(), stat, targets, true); err != nil {
		t.Fatalf("recursive-dir happy path: unexpected error %v", err)
	}
}

// TestPreflightRm_MultipleTargets_OneMissing: fail-fast contract.
// Three targets, the SECOND one is missing — the preflight should
// stop at the second and surface its path. The third target
// (which exists) is NOT consulted, so the test handler asserts that
// the second target's parent is the last URL visited.
//
// We don't try to assert "third URL never hit" because Stat reuses
// the same parent-listing URL for siblings, and we'd be coupling
// the test to download.Stat's internal call order. Naming the
// failing path in the error is the user-facing contract that
// matters here.
func TestPreflightRm_MultipleTargets_OneMissing(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/Documents/": `{"items":[
			{"name":"a.pdf","isDir":false,"size":1},
			{"name":"c.pdf","isDir":false,"size":1}
		]}`,
	}))
	targets := []rm.Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Documents/", Name: "a.pdf"},
		// b.pdf does NOT exist in the listing.
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Documents/", Name: "b.pdf"},
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Documents/", Name: "c.pdf"},
	}
	err := preflightRm(context.Background(), stat, targets, false)
	if err == nil {
		t.Fatal("expected error for the middle missing target")
	}
	if !strings.Contains(err.Error(), "b.pdf") ||
		!strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should name the missing target verbatim, got: %v", err)
	}
}

// TestPreflightRm_RmErrorPrefix: every preflight error should be
// prefixed with `rm:` so the user can match it to the command they
// typed (and so a future log-grep can filter rm-specific failures
// without ambiguity).
func TestPreflightRm_RmErrorPrefix(t *testing.T) {
	stat := newRmPreflightClient(t, rmListingHandler(t, map[string]string{
		"/api/resources/drive/Home/": `{"items":[]}`,
	}))
	targets := []rm.Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/", Name: "ghost"},
	}
	err := preflightRm(context.Background(), stat, targets, false)
	if err == nil {
		t.Fatal("expected error for missing target")
	}
	if !strings.HasPrefix(err.Error(), "rm:") {
		t.Errorf("error should be labelled with `rm:`, got: %v", err)
	}
}
