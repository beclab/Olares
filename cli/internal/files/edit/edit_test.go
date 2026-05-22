package edit

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient mirrors the rm / cp / rename / mkdir test
// harnesses: stand up a real httptest server, hand the caller a
// Client whose BaseURL points at it, and let the test inspect
// what landed on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// TestPlan_FileEdit is the canonical case: edit a file in place.
// Endpoint must be the per-resource path with NO trailing slash
// (a trailing slash would route the PUT through the directory
// handler).
func TestPlan_FileEdit(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/notes.md"}
	op, err := Plan(tgt)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if op.Endpoint != "/api/resources/drive/Home/Documents/notes.md" {
		t.Errorf("Endpoint: got %q", op.Endpoint)
	}
	if strings.HasSuffix(op.Endpoint, "/") {
		t.Errorf("Endpoint must NOT end with '/' for files: %q", op.Endpoint)
	}
	if op.DisplayPath != "drive/Home/Documents/notes.md" {
		t.Errorf("DisplayPath: got %q", op.DisplayPath)
	}
}

// TestPlan_PercentEncoding: the wire URL must use the same
// JS-shaped encoding the rest of the CLI uses (encodepath.EncodeURL).
// Verify a path with spaces and unicode survives unchanged on the
// display side and is percent-encoded on the wire side.
func TestPlan_PercentEncoding(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/My Docs/老照片.txt"}
	op, err := Plan(tgt)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if !strings.Contains(op.Endpoint, "/api/resources/drive/Home/My%20Docs/") {
		t.Errorf("path encoding: got %q", op.Endpoint)
	}
	// Display path keeps the human-readable form (so log lines
	// don't show users a URL-encoded mess).
	if op.DisplayPath != "drive/Home/My Docs/老照片.txt" {
		t.Errorf("DisplayPath: got %q", op.DisplayPath)
	}
}

// TestPlan_RejectsBadInput table-drives the input-validation
// contract. Each must produce an error that points at the
// offending input — error wording is part of the UX so we assert
// on substrings, not exact strings.
func TestPlan_RejectsBadInput(t *testing.T) {
	cases := []struct {
		name   string
		tgt    Target
		expect string // substring that must appear in the error
	}{
		{
			name:   "root of volume",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/"},
			expect: "root of",
		},
		{
			name:   "root via empty SubPath",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: ""},
			expect: "root of",
		},
		{
			name:   "trailing slash (directory path)",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/"},
			expect: "directory path",
		},
		{
			name:   "single-dot segment",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/foo/./bar"},
			expect: "path-traversal",
		},
		{
			name:   "double-dot segment",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/foo/../bar"},
			expect: "path-traversal",
		},
		{
			name:   "empty fileType",
			tgt:    Target{FileType: "", Extend: "Home", SubPath: "/foo"},
			expect: "empty fileType",
		},
		{
			name:   "empty extend",
			tgt:    Target{FileType: "drive", Extend: "", SubPath: "/foo"},
			expect: "empty fileType",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Plan(c.tgt)
			if err == nil {
				t.Fatalf("Plan(%+v): want error containing %q, got nil", c.tgt, c.expect)
			}
			if !strings.Contains(err.Error(), c.expect) {
				t.Errorf("Plan(%+v): error %q does not contain %q", c.tgt, err.Error(), c.expect)
			}
		})
	}
}

// TestPlan_NamespaceAllowlist locks in the allow-list from the
// package docstring:
//
//   - drive / sync / cache / external are supported (the LarePass
//     GUI's `onSaveFile` natively wires these to /api/resources
//     PUT).
//   - awss3 / google / dropbox are supported on the wire even
//     though the GUI's `onSaveFile` plumbing has known wiring
//     bugs there — the underlying PUT endpoint is uniform across
//     every namespace the resources handler covers, and the CLI
//     hits it directly.
//   - tencent / share / internal stay rejected — see the package
//     docstring for the upload-protocol divergence (tencent) and
//     the read-only / cross-user nature of share / internal.
//
// Adding a namespace to either side should be an obvious code-
// review signal — that's why this lives next to Plan and not as
// a generic FrontendPath helper.
func TestPlan_NamespaceAllowlist(t *testing.T) {
	supported := []string{"drive", "sync", "cache", "external", "awss3", "google", "dropbox"}
	for _, ft := range supported {
		t.Run("supported/"+ft, func(t *testing.T) {
			tgt := Target{FileType: ft, Extend: "x", SubPath: "/y.txt"}
			if _, err := Plan(tgt); err != nil {
				t.Errorf("Plan(%s): unexpected error: %v", ft, err)
			}
		})
	}
	rejected := []string{"tencent", "share", "internal", "unknown"}
	for _, ft := range rejected {
		t.Run("rejected/"+ft, func(t *testing.T) {
			tgt := Target{FileType: ft, Extend: "x", SubPath: "/y.txt"}
			_, err := Plan(tgt)
			if err == nil {
				t.Fatalf("Plan(%s): want error, got nil", ft)
			}
			if !strings.Contains(err.Error(), "not supported") {
				t.Errorf("Plan(%s): error %q does not say 'not supported'", ft, err.Error())
			}
		})
	}
}

// TestPlan_CloudDriveWireShape pins the wire URL we build for
// each cloud namespace. The LarePass web app's per-cloud
// utils.ts builds `/api/resources/<fileType><path>`; we mirror
// the same shape exactly so a future `files cat` / `files ls`
// against the same path lands on the same wire bytes the edit
// PUT sends.
func TestPlan_CloudDriveWireShape(t *testing.T) {
	cases := []struct {
		fileType string
		extend   string
		sub      string
		want     string
	}{
		{"awss3", "myacc", "/bucket/file.json", "/api/resources/awss3/myacc/bucket/file.json"},
		{"google", "myacc", "/Documents/draft.md", "/api/resources/google/myacc/Documents/draft.md"},
		{"dropbox", "myacc", "/Notes/idea.txt", "/api/resources/dropbox/myacc/Notes/idea.txt"},
	}
	for _, c := range cases {
		t.Run(c.fileType, func(t *testing.T) {
			op, err := Plan(Target{FileType: c.fileType, Extend: c.extend, SubPath: c.sub})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}
			if op.Endpoint != c.want {
				t.Errorf("Endpoint: got %q, want %q", op.Endpoint, c.want)
			}
		})
	}
}

// TestClient_PutBytes_Success: the client sends a PUT against the
// computed endpoint with the supplied body + Content-Type, and
// surfaces a 2xx as nil error. Inspect the captured request to
// confirm the wire shape matches the web app's saveFile call.
func TestClient_PutBytes_Success(t *testing.T) {
	var (
		gotMethod      string
		gotPath        string
		gotContentType string
		gotBody        []byte
	)
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	op := Op{Endpoint: "/api/resources/drive/Home/notes.md", DisplayPath: "drive/Home/notes.md"}
	want := []byte("hello world\n")
	if err := c.PutBytes(context.Background(), op, want, ""); err != nil {
		t.Fatalf("PutBytes: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method: got %q, want PUT", gotMethod)
	}
	if gotPath != "/api/resources/drive/Home/notes.md" {
		t.Errorf("path: got %q", gotPath)
	}
	if gotContentType != DefaultContentType {
		t.Errorf("Content-Type: got %q, want %q", gotContentType, DefaultContentType)
	}
	if string(gotBody) != string(want) {
		t.Errorf("body: got %q, want %q", gotBody, want)
	}
}

// TestClient_PutBytes_CustomContentType: when a non-empty content
// type is passed, it must be threaded through verbatim (so a user
// can save JSON / YAML / markdown with the right server-side
// hint).
func TestClient_PutBytes_CustomContentType(t *testing.T) {
	var gotCT string
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	op := Op{Endpoint: "/api/resources/drive/Home/data.json"}
	if err := c.PutBytes(context.Background(), op, []byte(`{"k":1}`), "application/json"); err != nil {
		t.Fatalf("PutBytes: %v", err)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", gotCT)
	}
}

// TestClient_PutBytes_HTTPError: a non-2xx surfaces as *HTTPError
// with the status / URL / method preserved, so the cobra layer's
// reformatter can branch on Status without stringly-typed parsing.
func TestClient_PutBytes_HTTPError(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	}))
	op := Op{Endpoint: "/api/resources/drive/Home/x.md"}
	err := c.PutBytes(context.Background(), op, []byte("ignored"), "")
	if err == nil {
		t.Fatalf("want error, got nil")
	}
	if !IsHTTPStatus(err, http.StatusForbidden) {
		t.Errorf("IsHTTPStatus(403): got false; err=%v", err)
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("error is not *HTTPError: %T", err)
	}
	if hErr.Method != http.MethodPut {
		t.Errorf("method: got %q", hErr.Method)
	}
	if !strings.Contains(hErr.Body, "forbidden") {
		t.Errorf("body preserved: got %q", hErr.Body)
	}
}

// TestClient_Fetch_Success: GET /api/raw/<encPath> returns the
// file contents verbatim (no envelope unwrapping). Confirm we
// pass through the bytes the server sent.
func TestClient_Fetch_Success(t *testing.T) {
	want := []byte("# Notes\n\nhello world\n")
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: got %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/raw/drive/Home/Documents/notes.md" {
			t.Errorf("path: got %q", r.URL.Path)
		}
		_, _ = w.Write(want)
	}))
	got, err := c.Fetch(context.Background(), "drive/Home/Documents/notes.md")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("body: got %q, want %q", got, want)
	}
}

// TestClient_Fetch_NotFound: a 404 surfaces as *HTTPError and
// IsNotFound returns true, so the cobra layer can branch on
// `--create` to start with an empty buffer instead of failing.
func TestClient_Fetch_NotFound(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	_, err := c.Fetch(context.Background(), "drive/Home/missing.md")
	if err == nil {
		t.Fatalf("want error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound: got false; err=%v", err)
	}
}

// TestClient_Put_EmptyEndpoint: defensive guard. Plan should never
// emit an empty Endpoint, but if a future caller forgets to call
// Plan we want a typed error rather than a silent malformed URL.
func TestClient_Put_EmptyEndpoint(t *testing.T) {
	c := &Client{HTTPClient: http.DefaultClient, BaseURL: "http://x"}
	err := c.PutBytes(context.Background(), Op{}, []byte("x"), "")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !strings.Contains(err.Error(), "empty Endpoint") {
		t.Errorf("error: got %q", err.Error())
	}
}
