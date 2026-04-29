package upload

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
	if err := client.Mkdir(context.Background(), "/drive/Home/Documents/Backups"); err != nil {
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
	if err := client.Mkdir(context.Background(), "/drive/Home/Existing"); err != nil {
		t.Errorf("Mkdir on 409 should be nil, got %v", err)
	}
}

func TestMkdir_OtherErrorReturns(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	err := client.Mkdir(context.Background(), "/drive/Home/Bad")
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

// --- WaitCloudTask: stage-2 cloud-transfer polling ---

// TestWaitCloudTask_RunningThenCompleted: the happy path. Server
// reports `running` for the first poll, then `completed`. Make sure
// WaitCloudTask returns nil, the onUpdate callback fires for the
// running poll (but NOT for completed — terminal status arrives via
// return), and the request URL carries the right node + task_id.
func TestWaitCloudTask_RunningThenCompleted(t *testing.T) {
	var hits int32
	var gotPath, gotQuery string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		n := atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			fmt.Fprint(w, `{"task":{"id":"t-1","status":"running","progress":42}}`)
			return
		}
		fmt.Fprint(w, `{"task":{"id":"t-1","status":"completed","progress":100}}`)
	}))

	var updates []CloudTaskUpdate
	err := client.WaitCloudTask(
		context.Background(), "node-a", "t-1",
		// 1ms keeps the test fast; the real default is 2s.
		time.Millisecond,
		func(u CloudTaskUpdate) {
			updates = append(updates, u)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := atomic.LoadInt32(&hits), int32(2); got != want {
		t.Errorf("hits = %d, want %d", got, want)
	}
	if want := "/api/task/node-a/"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if !strings.Contains(gotQuery, "task_id=t-1") {
		t.Errorf("query missing task_id=t-1: %q", gotQuery)
	}
	if len(updates) != 1 {
		t.Fatalf("got %d update callbacks, want 1 (only the running poll)", len(updates))
	}
	if updates[0].Status != "running" {
		t.Errorf("update status = %q, want %q", updates[0].Status, "running")
	}
	if updates[0].Progress != 42 {
		t.Errorf("update progress = %v, want 42", updates[0].Progress)
	}
}

// TestWaitCloudTask_Failed: server reports `failed` with a
// failed_reason. WaitCloudTask must surface the reason in the
// returned error so the cobra layer can render a useful CTA without
// the user having to dig into server logs.
func TestWaitCloudTask_Failed(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"task":{"id":"t-2","status":"failed","failed_reason":"bucket quota exceeded"}}`)
	}))
	err := client.WaitCloudTask(context.Background(), "n", "t-2", time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected error for failed task, got nil")
	}
	if !strings.Contains(err.Error(), "bucket quota exceeded") {
		t.Errorf("err = %v; want it to mention failed_reason", err)
	}
	if !strings.Contains(err.Error(), "t-2") {
		t.Errorf("err = %v; want it to mention task id", err)
	}
}

// TestWaitCloudTask_FailedWithoutReason: if the server reports
// `failed` but doesn't fill in `failed_reason`, the error message
// should still be self-describing — silently swallowing the failure
// would let the cobra layer print "uploaded" for a file that's not
// actually in the bucket.
func TestWaitCloudTask_FailedWithoutReason(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"task":{"id":"t-3","status":"failed"}}`)
	}))
	err := client.WaitCloudTask(context.Background(), "n", "t-3", time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected error for failed task, got nil")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("err = %v; want it to indicate failure", err)
	}
}

// TestWaitCloudTask_Cancelled: the server-side `cancelled` /
// `canceled` (both spellings — see OlaresTaskStatus enum) translate
// to a returned error. We test the British spelling here; the
// American spelling is covered by the same switch arm.
func TestWaitCloudTask_Cancelled(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"task":{"id":"t-4","status":"cancelled"}}`)
	}))
	err := client.WaitCloudTask(context.Background(), "n", "t-4", time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected error for cancelled task, got nil")
	}
	if !strings.Contains(err.Error(), "cancel") {
		t.Errorf("err = %v; want it to mention cancellation", err)
	}
}

// TestWaitCloudTask_CtxCancelBetweenPolls: while the task is still
// in flight (status: running) cancelling ctx must bail out promptly
// rather than hammering the server.
func TestWaitCloudTask_CtxCancelBetweenPolls(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"task":{"id":"t-5","status":"running","progress":1}}`)
	}))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := client.WaitCloudTask(ctx, "n", "t-5", 5*time.Millisecond, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v; want context.Canceled", err)
	}
}

// TestWaitCloudTask_EmptyArgs: WaitCloudTask must reject empty
// node / taskID up-front — both are URL segments / required query
// values, and silently accepting "" would produce a malformed
// request URL like /api/task//?task_id= that the server would
// either reject with a confusing 404 or (worse) match against a
// different task than the user expected.
func TestWaitCloudTask_EmptyArgs(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("server should not be hit for empty-args case")
		w.WriteHeader(http.StatusOK)
	}))
	if err := client.WaitCloudTask(context.Background(), "n", "", 0, nil); err == nil {
		t.Error("expected error for empty taskID, got nil")
	}
	if err := client.WaitCloudTask(context.Background(), "", "t", 0, nil); err == nil {
		t.Error("expected error for empty node, got nil")
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
