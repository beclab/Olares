package rm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

func TestPlan_GroupsByParent(t *testing.T) {
	targets := []Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Documents/", Name: "a.pdf"},
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Documents/", Name: "b.pdf"},
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Logs/", Name: "today.log"},
	}
	groups, err := Plan(targets, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("want 2 groups (one per parent), got %d", len(groups))
	}
	// Sorted alphabetically: Documents/ < Logs/.
	if groups[0].ParentSubPath != "/Documents/" {
		t.Errorf("groups[0].ParentSubPath = %q", groups[0].ParentSubPath)
	}
	if !equal(groups[0].Dirents, []string{"/a.pdf", "/b.pdf"}) {
		t.Errorf("groups[0].Dirents = %v", groups[0].Dirents)
	}
	if !equal(groups[1].Dirents, []string{"/today.log"}) {
		t.Errorf("groups[1].Dirents = %v", groups[1].Dirents)
	}
}

// TestPlan_DirIntentRequiresRecursive replicates Unix `rm`'s refusal:
// a trailing-slash target without -r must error, and the message must
// name the offending path.
func TestPlan_DirIntentRequiresRecursive(t *testing.T) {
	targets := []Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/", Name: "Backups", IsDirIntent: true},
	}
	_, err := Plan(targets, false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "directory") || !strings.Contains(err.Error(), "Backups") {
		t.Errorf("error should mention the directory and the -r flag, got: %v", err)
	}

	groups, err := Plan(targets, true)
	if err != nil {
		t.Fatalf("with -r the same plan should succeed: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Dirents) != 1 || groups[0].Dirents[0] != "/Backups/" {
		t.Errorf("dirent for dir target should have trailing slash, got %+v", groups)
	}
}

func TestPlan_RefusesEmptyName(t *testing.T) {
	_, err := Plan([]Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/"}, // Name missing
	}, true)
	if err == nil {
		t.Fatal("expected error for empty Name (root deletion)")
	}
	if !strings.Contains(err.Error(), "root") {
		t.Errorf("error should mention 'root', got: %v", err)
	}
}

func TestPlan_DeduplicatesDirents(t *testing.T) {
	// Same path twice on the command line — should land in one
	// dirent, not two.
	targets := []Target{
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Logs/", Name: "x.log"},
		{FileType: "drive", Extend: "Home", ParentSubPath: "/Logs/", Name: "x.log"},
	}
	groups, err := Plan(targets, false)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Dirents) != 1 {
		t.Errorf("duplicates should collapse, got groups=%+v", groups)
	}
}

func TestPlan_NoTargets(t *testing.T) {
	_, err := Plan(nil, false)
	if err == nil {
		t.Fatal("expected error for no targets")
	}
}

// TestDeleteBatch_WireShape exercises the actual HTTP DELETE: URL
// path encoding, JSON body shape, and trailing-slash on the parent.
// X-Authorization is no longer this client's responsibility — it is
// injected by the factory's refreshingTransport — so the header is not
// asserted here.
func TestDeleteBatch_WireShape(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotCType  string
		gotBody   []byte
	)
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	g := &Group{
		FileType:      "drive",
		Extend:        "Home",
		ParentSubPath: "/Documents/",
		Dirents:       []string{"/a.pdf", "/sub/"},
	}
	if err := client.DeleteBatch(context.Background(), g); err != nil {
		t.Fatalf("DeleteBatch: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("Method: want DELETE, got %s", gotMethod)
	}
	if gotPath != "/api/resources/drive/Home/Documents/" {
		t.Errorf("Path: got %q", gotPath)
	}
	if !strings.HasPrefix(gotCType, "application/json") {
		t.Errorf("Content-Type: got %q", gotCType)
	}
	var body deleteRequestBody
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("body unmarshal: %v (raw=%s)", err, gotBody)
	}
	if !equal(body.Dirents, []string{"/a.pdf", "/sub/"}) {
		t.Errorf("body.Dirents: got %v", body.Dirents)
	}
}

// TestDeleteBatch_ParentSlashEnforced confirms that a missing trailing
// slash on ParentSubPath is repaired before the wire call (the server
// requires it for the FileParam.convert split-on-/ check).
func TestDeleteBatch_ParentSlashEnforced(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	g := &Group{
		FileType:      "drive",
		Extend:        "Home",
		ParentSubPath: "/Logs", // missing trailing slash
		Dirents:       []string{"/today.log"},
	}
	if err := client.DeleteBatch(context.Background(), g); err != nil {
		t.Fatalf("DeleteBatch: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/") {
		t.Errorf("DeleteBatch should ensure trailing slash, got %q", gotPath)
	}
}

func TestDeleteBatch_NoOpOnEmpty(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("should not have hit the server for empty group")
	}))
	if err := client.DeleteBatch(context.Background(), &Group{}); err != nil {
		t.Errorf("empty group should be a no-op, got: %v", err)
	}
}

func TestDeleteBatch_HTTPError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":"nope"}`)
	}))
	g := &Group{
		FileType: "drive", Extend: "Home", ParentSubPath: "/", Dirents: []string{"/x"},
	}
	err := client.DeleteBatch(context.Background(), g)
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("want *HTTPError, got %T", err)
	}
	if hErr.Status != http.StatusForbidden {
		t.Errorf("status: want 403, got %d", hErr.Status)
	}
}

// equal is the bytes.Equal counterpart for string slices, kept local
// so the test file has no external test deps.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Compile-time check that the unused bytes import isn't a problem;
// tests above pass body bytes around.
var _ = bytes.Equal
