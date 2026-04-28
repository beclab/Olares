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
