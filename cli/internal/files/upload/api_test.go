package upload

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient wires a Client to an httptest.Server. X-Authorization
// is injected by the factory's refreshingTransport in production, not
// by upload.Client itself, so tests use a stock httptest *http.Client
// here and do NOT assert the access-token header.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

func TestFetchNodes(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/nodes/" {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}
		fmt.Fprintln(w, `{"data":{"nodes":[{"name":"node-a","master":true},{"name":"node-b"}]}}`)
	}))
	nodes, err := client.FetchNodes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(nodes))
	}
	if nodes[0].Name != "node-a" || !nodes[0].Master {
		t.Errorf("nodes[0] = %+v", nodes[0])
	}
	if nodes[1].Name != "node-b" || nodes[1].Master {
		t.Errorf("nodes[1] = %+v", nodes[1])
	}
}

func TestFetchNodes_EmptyList(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, `{"data":{"nodes":[]}}`)
	}))
	_, err := client.FetchNodes(context.Background())
	if err == nil {
		t.Fatal("expected error for empty nodes list, got nil")
	}
}

func TestGetUploadLink(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path: /upload/upload-link/<node>/
		wantPath := "/upload/upload-link/node-a/"
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, want %q", r.URL.Path, wantPath)
		}
		// Query: file_path uses encodeURIComponent semantics, so
		// '/' should be encoded as %2F (NOT '/').
		raw := r.URL.RawQuery
		if !strings.Contains(raw, "file_path=%2Fdrive%2FHome%2FDocuments%2F") {
			t.Errorf("raw query missing properly-encoded file_path: %q", raw)
		}
		if !strings.Contains(raw, "from=web") {
			t.Errorf("raw query missing from=web: %q", raw)
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "/seafhttp/upload-aj/repo-1/")
	}))
	link, err := client.GetUploadLink(context.Background(), "node-a", "/drive/Home/Documents/")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/seafhttp/upload-aj/repo-1/?ret-json=1"; link != want {
		t.Errorf("link = %q, want %q", link, want)
	}
}

func TestGetUploadLink_AppendsRetJSONOnExistingQuery(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "/seafhttp/upload-aj/repo-1/?token=abc")
	}))
	link, err := client.GetUploadLink(context.Background(), "n", "/drive/Home/")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/seafhttp/upload-aj/repo-1/?token=abc&ret-json=1"; link != want {
		t.Errorf("link = %q, want %q", link, want)
	}
}

func TestGetUploadedBytes(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("parent_dir") != "/drive/Home/Documents/" {
			t.Errorf("parent_dir = %q", r.URL.Query().Get("parent_dir"))
		}
		if r.URL.Query().Get("file_name") != "report.pdf" {
			t.Errorf("file_name = %q", r.URL.Query().Get("file_name"))
		}
		fmt.Fprint(w, `{"uploadedBytes":16777216}`)
	}))
	got, err := client.GetUploadedBytes(context.Background(), "n", "/drive/Home/Documents/", "report.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if got != 16777216 {
		t.Errorf("uploadedBytes = %d, want 16777216", got)
	}
}

// GetUploadedBytes is intentionally lenient: any error means "we
// haven't started yet, restart from 0" so a missing-file 404 doesn't
// abort a fresh upload. This is the same policy as the web app's
// silent .catch in resumejs.ts resumableUpload().
func TestGetUploadedBytes_404Treats0(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	got, err := client.GetUploadedBytes(context.Background(), "n", "/p/", "f")
	if err != nil {
		t.Fatalf("expected nil error on 404, got %v", err)
	}
	if got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestMkdir(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	if err := client.Mkdir(context.Background(), "Documents/Backups"); err != nil {
		t.Fatal(err)
	}
	if want := "/api/resources/drive/Home/Documents/Backups/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestMkdir_409Idempotent(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "exists", http.StatusConflict)
	}))
	if err := client.Mkdir(context.Background(), "Existing"); err != nil {
		t.Errorf("Mkdir on 409 should be nil, got %v", err)
	}
}

func TestMkdir_OtherErrorReturns(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	err := client.Mkdir(context.Background(), "Bad")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("expected *HTTPError, got %T (%v)", err, err)
	}
	if hErr.Status != http.StatusInternalServerError {
		t.Errorf("status = %d", hErr.Status)
	}
}

// Mkdir against root (or empty path) is a no-op — drive/Home always
// exists, and we don't want a stray POST to surprise the server.
func TestMkdir_RootIsNoop(t *testing.T) {
	hit := false
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	if err := client.Mkdir(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Errorf("Mkdir(\"\") should not hit the server")
	}
	if err := client.Mkdir(context.Background(), "/"); err != nil {
		t.Fatal(err)
	}
	if hit {
		t.Errorf("Mkdir(\"/\") should not hit the server")
	}
}
