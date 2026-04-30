package mkdir

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// newTestClient mirrors the rename / rm / cp test harnesses: stand
// up a real httptest server, hand the caller a Client whose BaseURL
// points at it, and let each test inspect what landed on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// --- Plan: validation + endpoint shape ------------------------------------

// TestPlan_HappyPath: typical "create one new directory" case. The
// wire URL must end with '/' (the backend's "this is a directory"
// marker) and the human-readable display path must match too.
func TestPlan_HappyPath(t *testing.T) {
	op, err := Plan(Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/Backups"})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if op.Endpoint != "/api/resources/drive/Home/Documents/Backups/" {
		t.Errorf("Endpoint: got %q", op.Endpoint)
	}
	if op.DisplayPath != "drive/Home/Documents/Backups/" {
		t.Errorf("DisplayPath: got %q", op.DisplayPath)
	}
}

// TestPlan_TrailingSlashIsTolerated: mkdir is always "create a
// directory", so the user-supplied trailing '/' on SubPath is
// neither required nor harmful — Plan must produce the same Op
// either way.
func TestPlan_TrailingSlashIsTolerated(t *testing.T) {
	a, _ := Plan(Target{FileType: "drive", Extend: "Home", SubPath: "/A/B"})
	b, _ := Plan(Target{FileType: "drive", Extend: "Home", SubPath: "/A/B/"})
	if a.Endpoint != b.Endpoint || a.DisplayPath != b.DisplayPath {
		t.Errorf("trailing-slash sensitivity: %+v vs %+v", a, b)
	}
}

// TestPlan_PercentEncoding confirms the wire URL uses the same
// JS-shaped encoding the rest of the CLI uses (encodepath.EncodeURL).
// Spaces become %20, unicode is UTF-8 percent-encoded, '/' separators
// stay literal.
func TestPlan_PercentEncoding(t *testing.T) {
	op, err := Plan(Target{FileType: "drive", Extend: "Home", SubPath: "/My Docs/老照片"})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	// Path segments encoded; '/' between them preserved.
	if !strings.Contains(op.Endpoint, "/api/resources/drive/Home/My%20Docs/") {
		t.Errorf("space encoding: got %q", op.Endpoint)
	}
	if !strings.HasSuffix(op.Endpoint, "/") {
		t.Errorf("must end with '/': got %q", op.Endpoint)
	}
}

// TestPlan_RefusesVolumeRoot: mkdir on `<fileType>/<extend>` is a
// no-op (the volume root always exists) and the auto-rename quirk
// would create an "Extend (1)" sibling instead — that's never what
// the user means, so refuse up-front.
func TestPlan_RefusesVolumeRoot(t *testing.T) {
	cases := []Target{
		{FileType: "drive", Extend: "Home", SubPath: "/"},
		{FileType: "drive", Extend: "Home", SubPath: ""},
		{FileType: "sync", Extend: "abc", SubPath: "/"},
	}
	for _, tc := range cases {
		_, err := Plan(tc)
		if err == nil {
			t.Errorf("Plan(%+v): want error, got nil", tc)
			continue
		}
		if !strings.Contains(err.Error(), "root") {
			t.Errorf("Plan(%+v): want error mentioning 'root', got %v", tc, err)
		}
	}
}

// TestPlan_RejectsInvalidSegments: empty / '.' / '..' segments
// anywhere in the path are obvious typos / path-traversal grenades;
// surface a typed error instead of silently building a malformed
// URL.
func TestPlan_RejectsInvalidSegments(t *testing.T) {
	cases := []struct {
		sub  string
		want string // substring expected in the error
	}{
		{"/foo/./bar", `"."`},
		{"/foo/../bar", `".."`},
		{"/foo//bar", `""`},
		{"/.", `"."`},
		{"/..", `".."`},
	}
	for _, c := range cases {
		_, err := Plan(Target{FileType: "drive", Extend: "Home", SubPath: c.sub})
		if err == nil {
			t.Errorf("Plan(%q): want error, got nil", c.sub)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("Plan(%q): want error mentioning %s, got %v", c.sub, c.want, err)
		}
	}
}

// TestPlan_RejectsEmptyFileTypeOrExtend: defense-in-depth check —
// the cobra-layer parser shouldn't let these through, but a typed
// error here beats a `/api/resources//Home/...` silent malformed URL.
func TestPlan_RejectsEmptyFileTypeOrExtend(t *testing.T) {
	if _, err := Plan(Target{FileType: "", Extend: "Home", SubPath: "/foo"}); err == nil {
		t.Error("want error for empty FileType")
	}
	if _, err := Plan(Target{FileType: "drive", Extend: "", SubPath: "/foo"}); err == nil {
		t.Error("want error for empty Extend")
	}
}

// --- PlanRecursive: -p mode segment expansion -----------------------------

// TestPlanRecursive_ProducesPrefixOps: `-p drive/Home/A/B/C` must
// produce three Ops in left-to-right order (A, A/B, A/B/C). The
// cobra layer interleaves Exists checks; PlanRecursive itself is
// pure validation + materialisation, no I/O.
func TestPlanRecursive_ProducesPrefixOps(t *testing.T) {
	ops, err := PlanRecursive(Target{FileType: "drive", Extend: "Home", SubPath: "/A/B/C"})
	if err != nil {
		t.Fatalf("PlanRecursive: %v", err)
	}
	want := []string{
		"drive/Home/A/",
		"drive/Home/A/B/",
		"drive/Home/A/B/C/",
	}
	if len(ops) != len(want) {
		t.Fatalf("got %d ops, want %d (%+v)", len(ops), len(want), ops)
	}
	for i, op := range ops {
		if op.DisplayPath != want[i] {
			t.Errorf("op[%d] DisplayPath = %q, want %q", i, op.DisplayPath, want[i])
		}
		if !strings.HasSuffix(op.Endpoint, "/") {
			t.Errorf("op[%d] Endpoint must end with '/': %q", i, op.Endpoint)
		}
	}
}

// TestPlanRecursive_SingleSegmentDegradesToSingleOp: the `-p` flag
// on a leaf-only path is a no-op semantic-wise (one Op, one POST);
// PlanRecursive must still return a single Op rather than special-
// casing.
func TestPlanRecursive_SingleSegmentDegradesToSingleOp(t *testing.T) {
	ops, err := PlanRecursive(Target{FileType: "drive", Extend: "Home", SubPath: "/Solo"})
	if err != nil {
		t.Fatalf("PlanRecursive: %v", err)
	}
	if len(ops) != 1 || ops[0].DisplayPath != "drive/Home/Solo/" {
		t.Errorf("got %+v, want one Op for drive/Home/Solo/", ops)
	}
}

// TestPlanRecursive_SharesPlanValidation: bad inputs that Plan
// rejects must also surface from PlanRecursive — we don't want the
// `-p` flag to bypass the volume-root or invalid-segment guards.
func TestPlanRecursive_SharesPlanValidation(t *testing.T) {
	if _, err := PlanRecursive(Target{FileType: "drive", Extend: "Home", SubPath: "/"}); err == nil {
		t.Error("PlanRecursive on volume root: want error")
	}
	if _, err := PlanRecursive(Target{FileType: "drive", Extend: "Home", SubPath: "/A/../B"}); err == nil {
		t.Error("PlanRecursive on '..' segment: want error")
	}
}

// --- Client.Mkdir: wire shape, success, errors ----------------------------

// TestMkdir_HappyPath: one POST against the planned endpoint, empty
// body, 200 OK from the server. The client must surface no error
// and the URL must arrive verbatim (trailing slash included).
func TestMkdir_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))

	op := Op{Endpoint: "/api/resources/drive/Home/Foo/", DisplayPath: "drive/Home/Foo/"}
	if err := client.Mkdir(context.Background(), op); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/resources/drive/Home/Foo/" {
		t.Errorf("path = %q", gotPath)
	}
}

// TestMkdir_409IsIdempotent: the server's rare 409 path (a few
// deployments / namespaces still return it instead of auto-renaming)
// must surface as nil — that's exactly what `-p` mode wants when
// another client created the same dir between our Exists probe and
// our POST.
func TestMkdir_409IsIdempotent(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "exists", http.StatusConflict)
	}))
	op := Op{Endpoint: "/api/resources/drive/Home/Existing/", DisplayPath: "drive/Home/Existing/"}
	if err := client.Mkdir(context.Background(), op); err != nil {
		t.Errorf("409 should be silenced, got %v", err)
	}
}

// TestMkdir_4xxSurfacesAsHTTPError: 401 / 403 / 404 must surface as
// typed *HTTPError so the cobra layer can branch on the status code
// for friendly CTAs.
func TestMkdir_4xxSurfacesAsHTTPError(t *testing.T) {
	for _, status := range []int{
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
	} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "rejected", status)
			}))
			op := Op{Endpoint: "/api/resources/drive/Home/Foo/", DisplayPath: "drive/Home/Foo/"}
			err := client.Mkdir(context.Background(), op)
			if err == nil {
				t.Fatalf("status %d: want error, got nil", status)
			}
			var hErr *HTTPError
			if !errors.As(err, &hErr) {
				t.Fatalf("status %d: want *HTTPError, got %T: %v", status, err, err)
			}
			if hErr.Status != status {
				t.Errorf("status: want %d, got %d", status, hErr.Status)
			}
		})
	}
}

// TestMkdir_RejectsEmptyEndpoint: defense-in-depth — Mkdir is
// supposed to receive Ops produced by Plan / PlanRecursive (which
// always populate Endpoint). A zero-value Op should fail fast
// instead of POSTing to BaseURL itself.
func TestMkdir_RejectsEmptyEndpoint(t *testing.T) {
	client := &Client{HTTPClient: http.DefaultClient, BaseURL: "http://unused"}
	if err := client.Mkdir(context.Background(), Op{}); err == nil {
		t.Error("empty Op: want error, got nil")
	}
}

// --- Client.Exists: parent-listing existence check ------------------------

// TestExists_FoundDirInDriveEnvelope: `items` envelope (Drive /
// Sync / Cache / External / Share) — the canonical case the web
// app's navigation uses.
func TestExists_FoundDirInDriveEnvelope(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parent listing GET with trailing slash.
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		_, _ = io.WriteString(w, `{
			"items":[
				{"name":"Other","isDir":true},
				{"name":"Backups","isDir":true}
			]
		}`)
	}))
	found, isDir, err := client.Exists(context.Background(), "drive/Home/Documents/Backups")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !found {
		t.Errorf("want found=true")
	}
	if !isDir {
		t.Errorf("want isDir=true")
	}
}

// TestExists_FoundInCloudDataEnvelope: cloud drives (awss3 / google /
// dropbox / tencent) put listings under top-level `data` instead of
// `items`. Exists must accept both shapes for `-p` to work uniformly
// across all namespaces.
func TestExists_FoundInCloudDataEnvelope(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{
			"data":[
				{"name":"Backups","isDir":true,"path":"/Backups"}
			],
			"fileExtend":"AKIA...","filePath":"/","fileType":"awss3"
		}`)
	}))
	found, isDir, err := client.Exists(context.Background(), "awss3/AKIA.../Backups")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !found || !isDir {
		t.Errorf("want found=true isDir=true, got found=%v isDir=%v", found, isDir)
	}
}

// TestExists_NotFoundInListing: parent listing succeeds but the
// basename isn't in it — that's "this leaf doesn't exist", and `-p`
// mode must POST to create it. (NOT an error.)
func TestExists_NotFoundInListing(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"items":[{"name":"Other","isDir":true}]}`)
	}))
	found, _, err := client.Exists(context.Background(), "drive/Home/Documents/Backups")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if found {
		t.Errorf("want found=false")
	}
}

// TestExists_ParentMissing404: when the parent itself doesn't exist
// (404 on the parent listing) Exists must return found=false (the
// outer `-p` walk will create the parent in a previous iteration).
// We do NOT surface 404 here as an error — that would conflate
// "user typo" with "haven't gotten to this prefix yet".
func TestExists_ParentMissing404(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	found, _, err := client.Exists(context.Background(), "drive/Home/Missing/Leaf")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if found {
		t.Errorf("want found=false on parent 404")
	}
}

// TestExists_VolumeRootIsAlwaysDir: paths that resolve to
// `<fileType>/<extend>` are the volume root; they always exist as
// directories without any HTTP traffic. Same convention as
// download.Stat.
func TestExists_VolumeRootIsAlwaysDir(t *testing.T) {
	var hits int32
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	found, isDir, err := client.Exists(context.Background(), "drive/Home")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !found || !isDir {
		t.Errorf("volume root: want found=true isDir=true")
	}
	if h := atomic.LoadInt32(&hits); h != 0 {
		t.Errorf("volume root should not hit the network, got %d hit(s)", h)
	}
}

// TestExists_FoundButNotDir: when the basename exists as a file
// (not a directory) we must report isDir=false; the cobra layer
// translates this into a hard error so `-p` doesn't auto-rename
// around the conflict.
func TestExists_FoundButNotDir(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"items":[{"name":"clash","isDir":false,"size":42}]}`)
	}))
	found, isDir, err := client.Exists(context.Background(), "drive/Home/clash")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !found {
		t.Errorf("want found=true")
	}
	if isDir {
		t.Errorf("want isDir=false (clash is a file)")
	}
}

// TestExists_NonNotFoundErrorBubblesUp: a non-404 4xx/5xx from the
// parent listing isn't "this path doesn't exist" — it's a server
// problem the user needs to know about (auth, permission, internal
// error). Surface as *HTTPError so the cobra layer can reformat it.
func TestExists_NonNotFoundErrorBubblesUp(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	}))
	_, _, err := client.Exists(context.Background(), "drive/Home/Foo/Bar")
	if err == nil {
		t.Fatal("want error for 403, got nil")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) || hErr.Status != http.StatusForbidden {
		t.Errorf("want *HTTPError(403), got %T: %v", err, err)
	}
}

// TestIsHTTPStatus is a tiny sanity check for the predicate the
// cobra layer uses to branch on common 4xx codes.
func TestIsHTTPStatus(t *testing.T) {
	hErr := &HTTPError{Status: http.StatusConflict}
	if !IsHTTPStatus(hErr, http.StatusConflict) {
		t.Error("want true for matching status")
	}
	if IsHTTPStatus(hErr, http.StatusNotFound) {
		t.Error("want false for non-matching status")
	}
	if IsHTTPStatus(errors.New("plain"), http.StatusConflict) {
		t.Error("want false for non-HTTPError")
	}
}
