package rename

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestClient mirrors the rm / cp test harnesses: stand up a real
// httptest server, hand the caller a Client whose BaseURL points at
// it, and let the test inspect what landed on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// TestPlan_FileRename is the canonical case: rename a file in place,
// no trailing slash anywhere. Endpoint must be the per-resource path
// with a single `destination` query param.
func TestPlan_FileRename(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"}
	op, err := Plan(tgt, "bar.pdf")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if !strings.HasPrefix(op.Endpoint, "/api/resources/drive/Home/Documents/foo.pdf?destination=") {
		t.Errorf("Endpoint: got %q", op.Endpoint)
	}
	if !strings.Contains(op.Endpoint, "destination=bar.pdf") {
		t.Errorf("destination query: got %q", op.Endpoint)
	}
	if op.IsDir {
		t.Errorf("IsDir: want false")
	}
	if op.DisplaySrc != "drive/Home/Documents/foo.pdf" {
		t.Errorf("DisplaySrc: got %q", op.DisplaySrc)
	}
	if op.DisplayDst != "drive/Home/Documents/bar.pdf" {
		t.Errorf("DisplayDst: got %q", op.DisplayDst)
	}
}

// TestPlan_DirRename preserves the trailing slash on both the wire
// path AND the human-readable display destination. The slash is the
// frontend's directory marker, so dropping it here would route the
// rename through the file handler.
func TestPlan_DirRename(t *testing.T) {
	tgt := Target{
		FileType: "drive", Extend: "Home", SubPath: "/Documents/old/",
		IsDirIntent: true,
	}
	op, err := Plan(tgt, "new")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if !strings.HasPrefix(op.Endpoint, "/api/resources/drive/Home/Documents/old/?destination=") {
		t.Errorf("Endpoint: got %q (expect trailing '/' before query)", op.Endpoint)
	}
	if !op.IsDir {
		t.Errorf("IsDir: want true")
	}
	if op.DisplayDst != "drive/Home/Documents/new/" {
		t.Errorf("DisplayDst: got %q (expect trailing '/')", op.DisplayDst)
	}
}

// TestPlan_PercentEncoding confirms the wire URL uses the same
// JS-shaped encoding the rest of the CLI uses, both for the source
// path (each segment URL-encoded, '/' preserved) and for the
// destination query value (encodeURIComponent shape — space is %20,
// not '+').
func TestPlan_PercentEncoding(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/My Docs/老照片.jpg"}
	op, err := Plan(tgt, "新 photo.jpg")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if !strings.Contains(op.Endpoint, "/api/resources/drive/Home/My%20Docs/") {
		t.Errorf("path encoding: got %q", op.Endpoint)
	}
	if strings.Contains(op.Endpoint, "destination=新+photo.jpg") {
		t.Errorf("query value should be %%20-encoded, not '+': %q", op.Endpoint)
	}
	// Round-trip through net/url to confirm the destination decodes
	// back to the literal new name (UTF-8 + space preserved).
	u, err := url.Parse("http://x" + op.Endpoint)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := u.Query().Get("destination"); got != "新 photo.jpg" {
		t.Errorf("destination decoded: got %q", got)
	}
}

// TestPlan_RejectsEmpty / TestPlan_RejectsSlash / TestPlan_RejectsDots
// table-drive the input-validation contract. Each must produce an
// error that points at the offending input — error wording is part
// of the UX so we assert on substrings, not exact strings.
func TestPlan_RejectsBadNames(t *testing.T) {
	cases := []struct {
		name    string
		newName string
		expect  string // substring that must appear in the error
	}{
		{"empty", "", "empty"},
		{"whitespace-only", "   ", "empty"},
		{"contains forward slash", "sub/foo.pdf", "must not contain"},
		{"contains backslash", "sub\\foo.pdf", "must not contain"},
		{"single dot", ".", "path-traversal"},
		{"double dot", "..", "path-traversal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"}
			_, err := Plan(tgt, tc.newName)
			if err == nil {
				t.Fatalf("expected error for %q", tc.newName)
			}
			if !strings.Contains(err.Error(), tc.expect) {
				t.Errorf("error %q should contain %q", err.Error(), tc.expect)
			}
		})
	}
}

// TestPlan_RejectsRoot mirrors rm/cp: refusing to operate on the
// volume root is a CLI safety policy, not a server one. A typo here
// has very expensive consequences.
func TestPlan_RejectsRoot(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/", IsDirIntent: true}
	_, err := Plan(tgt, "anything")
	if err == nil {
		t.Fatal("expected refusal for root")
	}
	if !strings.Contains(err.Error(), "root") {
		t.Errorf("error should mention 'root', got: %v", err)
	}
}

// TestPlan_RejectsSameName catches the no-op case before it reaches
// the wire — `files rename foo.pdf foo.pdf` is virtually always a
// typo, and surfacing it as a clear error here is friendlier than
// whatever the server returns (which varies between drive/sync/etc).
func TestPlan_RejectsSameName(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"}
	_, err := Plan(tgt, "foo.pdf")
	if err == nil {
		t.Fatal("expected refusal for same-name rename")
	}
	if !strings.Contains(err.Error(), "same") {
		t.Errorf("error should mention 'same', got: %v", err)
	}
}

// TestPlan_RejectsProtectedDriveHomeChild pins the LarePass-aligned
// policy that the system-managed first-level children directly under
// drive/Home/ (Pictures / Music / Movies / Downloads / Documents /
// Code / Cache / Data / Home / Ollama / Huggingface) refuse rename
// — same effective UX as the LarePass GUI's `disableMenuItem` array
// in apps/packages/app/src/stores/operation.ts (gated by the user
// being at /Files/Home/), so a scripted `files rename` cannot
// silently produce a state the GUI couldn't reach.
//
// The matrix below encodes the full LarePass list verbatim plus a
// handful of negative cases: deeper paths under a protected name
// (user content, must NOT trip), drive/Data/<same-name> (different
// extend), and other namespaces (out of scope by construction).
func TestPlan_RejectsProtectedDriveHomeChild(t *testing.T) {
	rejectCases := []struct {
		name string
		sub  string
		dir  bool
	}{
		{"Pictures dir", "/Pictures", true},
		{"Pictures with trailing slash", "/Pictures/", true},
		{"Music", "/Music", true},
		{"Movies", "/Movies", true},
		{"Downloads", "/Downloads", true},
		{"Documents", "/Documents", true},
		{"Code", "/Code", true},
		{"Cache", "/Cache", true},
		{"Data", "/Data", true},
		{"Home nested", "/Home", true},
		{"Ollama", "/Ollama", true},
		{"Huggingface one-word", "/Huggingface", true},
	}
	for _, c := range rejectCases {
		t.Run("reject "+c.name, func(t *testing.T) {
			tgt := Target{
				FileType:    "drive",
				Extend:      "Home",
				SubPath:     c.sub,
				IsDirIntent: c.dir,
			}
			_, err := Plan(tgt, "Renamed")
			if err == nil {
				t.Fatalf("Plan: expected refusal for drive/Home%s", c.sub)
			}
			msg := err.Error()
			if !strings.Contains(msg, "system-managed Home folder") {
				t.Errorf("error should mention 'system-managed Home folder'; got: %v", err)
			}
			if !strings.Contains(msg, "Files GUI") {
				t.Errorf("error should reference the Files GUI for context; got: %v", err)
			}
			// Sanity-check that the message enumerates the
			// protected names so the user can see the full set
			// without consulting docs.
			if !strings.Contains(msg, "Pictures") || !strings.Contains(msg, "Huggingface") {
				t.Errorf("error should enumerate protected names (Pictures / Huggingface); got: %v", err)
			}
		})
	}

	// User-content paths and other namespaces / extends MUST stay
	// renamable — the policy must not over-extend.
	allowCases := []struct {
		name string
		t    Target
		dst  string
	}{
		{
			// Album/sub-folder under Pictures: pure user content.
			name: "deeper path under Pictures",
			t: Target{
				FileType: "drive", Extend: "Home",
				SubPath: "/Pictures/Trip2024/", IsDirIntent: true,
			},
			dst: "Trip2025",
		},
		{
			// File inside Documents.
			name: "file inside Documents",
			t: Target{
				FileType: "drive", Extend: "Home",
				SubPath: "/Documents/notes.md",
			},
			dst: "draft.md",
		},
		{
			// drive/Data/<same-name>: different volume root, the
			// policy is Home-only (the GUI also gates on
			// /Files/Home/, not /Files/Data/).
			name: "drive Data same name",
			t: Target{
				FileType: "drive", Extend: "Data",
				SubPath: "/Pictures", IsDirIntent: true,
			},
			dst: "PicturesArchive",
		},
		{
			// Other namespace: same-named entry but out of scope.
			name: "sync repo same name",
			t: Target{
				FileType: "sync", Extend: "abc-repo",
				SubPath: "/Pictures", IsDirIntent: true,
			},
			dst: "PicturesArchive",
		},
		{
			// User-created folder at drive/Home/<name> not in the
			// protected list.
			name: "drive Home user folder",
			t: Target{
				FileType: "drive", Extend: "Home",
				SubPath: "/MyProjects", IsDirIntent: true,
			},
			dst: "WorkProjects",
		},
		{
			// Lowercase variant — even if it happened to exist as
			// a real dir, it isn't in the case-sensitive policy
			// list, so it must not be guarded.
			name: "lowercase pictures not protected",
			t: Target{
				FileType: "drive", Extend: "Home",
				SubPath: "/pictures", IsDirIntent: true,
			},
			dst: "renamed",
		},
	}
	for _, c := range allowCases {
		t.Run("allow "+c.name, func(t *testing.T) {
			if _, err := Plan(c.t, c.dst); err != nil {
				t.Errorf("Plan: unexpected refusal for %s/%s%s: %v",
					c.t.FileType, c.t.Extend, c.t.SubPath, err)
			}
		})
	}
}

// TestPlan_RejectsProtectedExternalChild pins the LarePass-aligned
// policy that the system-managed AI mountpoint folders under
// `external/<node>/...` refuse rename: the GUI greys out rename
// via `externalFolderWhiteList` (depth-1: ai) and
// `externalAiFolderWhiteList` (depth-2: output / model / comfyui /
// ollama) in apps/packages/app/src/stores/operation.ts. The CLI
// mirrors that so a scripted `files rename` can't silently break
// the contract Ollama / ComfyUI / Huggingface readers look up by
// name.
//
// Path-shape ground truth: the GUI's `<X>` in
// `/Files/External/<X>/` is the LarePass `masterNode`
// (apps/.../external/data.ts:77), so `<X>` maps to
// FrontendPath.Extend on the CLI — the GUI's `ai/` row lives at
// `external/<node>/ai/` (SubPath="/ai/"), NOT under any nested
// volume segment.
//
// The matrix covers both layers (depth-1 + depth-2) plus negative
// cases that LOOK adjacent (case mismatches, depth-2 under a
// non-ai parent, deeper user-content paths, other namespaces
// with same-named entries).
func TestPlan_RejectsProtectedExternalChild(t *testing.T) {
	rejectCases := []struct {
		name string
		t    Target
	}{
		// ----- depth-1: external/<node>/ai -----
		{
			name: "depth-1 ai under olares node",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai", IsDirIntent: true,
			},
		},
		{
			name: "depth-1 ai with trailing slash",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/", IsDirIntent: true,
			},
		},
		{
			name: "depth-1 ai on arbitrary node name (node is opaque)",
			t: Target{
				FileType: "external", Extend: "node-1",
				SubPath: "/ai/", IsDirIntent: true,
			},
		},
		// ----- depth-2: external/<node>/ai/<name> -----
		{
			name: "depth-2 ai/output",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/output", IsDirIntent: true,
			},
		},
		{
			name: "depth-2 ai/model",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/model/", IsDirIntent: true,
			},
		},
		{
			name: "depth-2 ai/comfyui",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/comfyui", IsDirIntent: true,
			},
		},
		{
			name: "depth-2 ai/ollama",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/ollama/", IsDirIntent: true,
			},
		},
	}
	for _, c := range rejectCases {
		t.Run("reject "+c.name, func(t *testing.T) {
			_, err := Plan(c.t, "Renamed")
			if err == nil {
				t.Fatalf("Plan: expected refusal for %s/%s%s",
					c.t.FileType, c.t.Extend, c.t.SubPath)
			}
			msg := err.Error()
			if !strings.Contains(msg, "system-managed AI mountpoint folder") {
				t.Errorf("error should mention 'system-managed AI mountpoint folder'; got: %v", err)
			}
			if !strings.Contains(msg, "LarePass") {
				t.Errorf("error should reference LarePass for context; got: %v", err)
			}
			// The error should echo the offending path so the
			// user can match it against their command line.
			displayHint := c.t.FileType + "/" + c.t.Extend
			if !strings.Contains(msg, displayHint) {
				t.Errorf("error should echo path prefix %q; got: %v", displayHint, err)
			}
			// Sanity-check that both whitelists are enumerated so
			// the user sees the full policy without consulting
			// docs.
			if !strings.Contains(msg, "comfyui") || !strings.Contains(msg, "output") {
				t.Errorf("error should enumerate depth-2 whitelist (comfyui / output); got: %v", err)
			}
		})
	}

	// Negative cases: paths that LOOK adjacent must remain
	// renameable so the policy does not over-extend.
	allowCases := []struct {
		name string
		t    Target
		dst  string
	}{
		{
			// Depth-2 under a non-ai depth-1 parent — name
			// happens to match the ai-whitelist but the parent
			// isn't "ai", so this is just a regular dir inside
			// some volume.
			name: "depth-2 under non-ai parent",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/USB-0/output", IsDirIntent: true,
			},
			dst: "outputs-archive",
		},
		{
			// Depth-2 ai/<other> not in the whitelist — user
			// content under ai/, freely renameable.
			name: "depth-2 ai/<other> not whitelisted",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/my-experiments", IsDirIntent: true,
			},
			dst: "my-archive",
		},
		{
			// Depth-3 under ai/output — user runs, freely
			// renameable (only the dir itself is pinned).
			name: "depth-3 under ai/output",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/output/run-2026-05", IsDirIntent: true,
			},
			dst: "run-2026-06",
		},
		{
			// Case-sensitive mismatch — "AI" must NOT trip the
			// policy (the GUI compares lowercase string values).
			name: "case mismatch on depth-1 name",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/AI", IsDirIntent: true,
			},
			dst: "AI-archive",
		},
		{
			name: "case mismatch on depth-2 name",
			t: Target{
				FileType: "external", Extend: "olares",
				SubPath: "/ai/Output", IsDirIntent: true,
			},
			dst: "Output-archive",
		},
		{
			// Different fileType that looks similar — drive/Home
			// has its own policy (ProtectedDriveHomeChildren).
			// drive/Home/ai is NOT a system folder, so it must
			// stay renameable.
			name: "drive/Home/ai is not external",
			t: Target{
				FileType: "drive", Extend: "Home",
				SubPath: "/ai", IsDirIntent: true,
			},
			dst: "ai-archive",
		},
		{
			// sync/<repo>/ai is also not external — same name
			// happens to exist but in a different namespace.
			name: "sync/<repo>/ai is not external",
			t: Target{
				FileType: "sync", Extend: "abc-repo",
				SubPath: "/ai", IsDirIntent: true,
			},
			dst: "ai-archive",
		},
	}
	for _, c := range allowCases {
		t.Run("allow "+c.name, func(t *testing.T) {
			if _, err := Plan(c.t, c.dst); err != nil {
				t.Errorf("Plan: unexpected refusal for %s/%s%s: %v",
					c.t.FileType, c.t.Extend, c.t.SubPath, err)
			}
		})
	}
}

// TestRename_WireShape inspects the actual PATCH that lands on the
// server — this is the test that breaks loudly if the wire protocol
// drifts (verb, URL path, query encoding, no body).
func TestRename_WireShape(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotQuery  string
		gotLen    int64
		gotAccept string
	)
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotLen = r.ContentLength
		gotAccept = r.Header.Get("Accept")
		raw, _ := io.ReadAll(r.Body)
		_ = raw // body should be empty
		w.WriteHeader(http.StatusOK)
	}))
	op, err := Plan(
		Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"},
		"bar.pdf",
	)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := client.Rename(context.Background(), op); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("Method: got %s", gotMethod)
	}
	if gotPath != "/api/resources/drive/Home/Documents/foo.pdf" {
		t.Errorf("Path: got %q", gotPath)
	}
	// RawQuery preserves percent-encoding, so the literal substring
	// match matches the wire bytes the server actually parses.
	if !strings.Contains(gotQuery, "destination=bar.pdf") {
		t.Errorf("Query: got %q", gotQuery)
	}
	if gotLen > 0 {
		t.Errorf("ContentLength: got %d, want 0", gotLen)
	}
	if !strings.Contains(gotAccept, "application/json") {
		t.Errorf("Accept: got %q", gotAccept)
	}
}

// TestRename_HTTPError surfaces non-2xx as *HTTPError — same contract
// the cobra layer uses to reformat 401 / 403 / 404 / 409 with
// friendly CTAs.
func TestRename_HTTPError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":"already exists"}`)
	}))
	op, err := Plan(
		Target{FileType: "drive", Extend: "Home", SubPath: "/a"},
		"b",
	)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	err = client.Rename(context.Background(), op)
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("want *HTTPError, got %T", err)
	}
	if hErr.Status != http.StatusConflict {
		t.Errorf("status: got %d", hErr.Status)
	}
	if !strings.Contains(hErr.Body, "already exists") {
		t.Errorf("body should preserve server message, got: %q", hErr.Body)
	}
}
