package download

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient wires a *Client up to an httptest.Server. We always
// inject a recognisable token so each test can assert that
// X-Authorization made it onto the wire (the most common refactor
// regression in this CLI is dropping the header).
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient:  srv.Client(),
		BaseURL:     srv.URL,
		AccessToken: "test-token-XYZ",
	}, srv
}

// TestStat_LeafLookupInParent: Stat must list the *parent* and find
// the leaf there, NOT probe /api/resources/<encFilePath> directly.
// The backend's single-file GET path returns HTTP 500 for many real
// files (it tries to read content into the response — see stat.go's
// statByParentListing comment for why), so the parent-listing
// strategy is the only one the wire actually supports.
func TestStat_LeafLookupInParent(t *testing.T) {
	for _, tc := range []struct {
		name        string
		path        string
		wantParent  string // expected URL path on the wire, including trailing slash
		wantLeaf    string
		wantIsDir   bool
		wantSize    int64
		entriesJSON string
	}{
		{
			name:       "file under nested parent",
			path:       "drive/Home/Documents/foo.pdf",
			wantParent: "/api/resources/drive/Home/Documents/",
			wantLeaf:   "foo.pdf",
			wantSize:   1234,
			entriesJSON: `[
				{"name":"other.pdf","isDir":false,"size":1},
				{"name":"foo.pdf","isDir":false,"size":1234}
			]`,
		},
		{
			name:       "directory under volume root",
			path:       "drive/Home/Backups",
			wantParent: "/api/resources/drive/Home/",
			wantLeaf:   "Backups",
			wantIsDir:  true,
			entriesJSON: `[
				{"name":"Backups","isDir":true,"size":0}
			]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Authorization") != "test-token-XYZ" {
					t.Fatalf("missing/wrong X-Authorization")
				}
				if r.Method != http.MethodGet {
					t.Fatalf("want GET, got %s", r.Method)
				}
				if r.URL.Path != tc.wantParent {
					t.Fatalf("Stat should hit the parent listing URL %q, got %q",
						tc.wantParent, r.URL.Path)
				}
				_, _ = io.WriteString(w, `{"items":`+tc.entriesJSON+`}`)
			}))
			info, err := client.Stat(context.Background(), tc.path)
			if err != nil {
				t.Fatalf("Stat: %v", err)
			}
			if info.Name != tc.wantLeaf {
				t.Errorf("Name: want %q, got %q", tc.wantLeaf, info.Name)
			}
			if info.IsDir != tc.wantIsDir {
				t.Errorf("IsDir: want %v, got %v", tc.wantIsDir, info.IsDir)
			}
			if !info.IsDir && info.Size != tc.wantSize {
				t.Errorf("Size: want %d, got %d", tc.wantSize, info.Size)
			}
		})
	}
}

// TestStat_VolumeRoot: paths that are themselves the root of a
// `<fileType>/<extend>` tuple (e.g. "drive/Home") have no parent to
// list. Stat returns a synthetic dir record without touching the
// network.
func TestStat_VolumeRoot(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Stat at volume root should not hit the network, got %s %s", r.Method, r.URL.Path)
	}))
	info, err := client.Stat(context.Background(), "drive/Home")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir || info.Name != "Home" {
		t.Errorf("volume-root Stat: want {Name:Home, IsDir:true}, got %+v", info)
	}
}

// TestStat_NotFound covers two paths that should both surface as
// IsNotFound: the parent listing returning 404, and the parent
// existing but the leaf not being in it (synthetic 404).
func TestStat_NotFound(t *testing.T) {
	t.Run("parent 404", func(t *testing.T) {
		client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"error":"not found"}`)
		}))
		_, err := client.Stat(context.Background(), "drive/Home/missing-dir/foo")
		if !IsNotFound(err) {
			t.Errorf("IsNotFound should be true for parent 404, got: %v", err)
		}
	})
	t.Run("leaf not in parent listing", func(t *testing.T) {
		client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"items":[{"name":"other.txt","isDir":false,"size":1}]}`)
		}))
		_, err := client.Stat(context.Background(), "drive/Home/Documents/missing.txt")
		if !IsNotFound(err) {
			t.Errorf("IsNotFound should be true for missing leaf, got: %v", err)
		}
	})
}

func TestList_HappyPath(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/") {
			t.Errorf("List should hit a trailing-slash URL, got %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{
			"items":[
				{"name":"a.txt","isDir":false,"size":10},
				{"name":"sub","isDir":true,"size":0}
			]
		}`)
	}))
	entries, err := client.List(context.Background(), "drive/Home/Documents")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "a.txt" || entries[0].IsDir || entries[0].Size != 10 {
		t.Errorf("entries[0] mismatch: %+v", entries[0])
	}
	if entries[1].Name != "sub" || !entries[1].IsDir {
		t.Errorf("entries[1] mismatch: %+v", entries[1])
	}
}

// TestDownloadFile_Fresh exercises the "no local file yet" path: full
// 200, written via tmp+rename. Asserts the local file ends up with the
// exact body bytes and the .tmp file does NOT linger.
func TestDownloadFile_Fresh(t *testing.T) {
	body := []byte("hello world payload")
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			t.Fatalf("fresh download should not send Range, got %q", r.Header.Get("Range"))
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
	}))
	dst := filepath.Join(t.TempDir(), "fresh.bin")

	written, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{}, nil)
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if written != int64(len(body)) {
		t.Errorf("written: want %d, got %d", len(body), written)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("body mismatch: got %q want %q", got, body)
	}
	if _, err := os.Stat(dst + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf(".tmp file should be cleaned up after rename, stat err = %v", err)
	}
}

// TestDownloadFile_RefuseExisting confirms the safety policy: without
// --overwrite or --resume, an existing local file blocks the download.
func TestDownloadFile_RefuseExisting(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("should not have hit the server")
	}))
	dir := t.TempDir()
	dst := filepath.Join(dir, "exists.bin")
	if err := os.WriteFile(dst, []byte("existing"), 0o644); err != nil {
		t.Fatalf("seed dst: %v", err)
	}
	_, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{}, nil)
	if err == nil {
		t.Fatal("expected refusal error")
	}
	if !strings.Contains(err.Error(), "--overwrite") || !strings.Contains(err.Error(), "--resume") {
		t.Errorf("error should mention both flags, got: %v", err)
	}
}

// TestDownloadFile_Resume sends a Range header and an appended body;
// the file should grow from local size to local+payload.
func TestDownloadFile_Resume(t *testing.T) {
	prefix := []byte("first half-")
	tail := []byte("second half!")

	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rng := r.Header.Get("Range")
		wantRng := fmt.Sprintf("bytes=%d-", len(prefix))
		if rng != wantRng {
			t.Fatalf("Range header: want %q, got %q", wantRng, rng)
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d",
			len(prefix), len(prefix)+len(tail)-1, len(prefix)+len(tail)))
		w.Header().Set("Content-Length", strconv.Itoa(len(tail)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(tail)
	}))
	dir := t.TempDir()
	dst := filepath.Join(dir, "resume.bin")
	if err := os.WriteFile(dst, prefix, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	written, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{Resume: true}, nil)
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if written != int64(len(tail)) {
		t.Errorf("written should be the tail only: want %d, got %d", len(tail), written)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	want := append(append([]byte(nil), prefix...), tail...)
	if !bytes.Equal(got, want) {
		t.Errorf("body mismatch: got %q want %q", got, want)
	}
}

// errCut is an io.Reader that fails immediately — used after a prefix
// in io.MultiReader so the server sends a short 206 body then stops.
type errCut struct{}

func (errCut) Read([]byte) (int, error) {
	return 0, errors.New("simulated transport cut")
}

// TestDownloadFile_ResumeRetryRefreshesRange: a partial 206 append
// followed by a read error must not duplicate bytes on retry; the
// second GET must send Range: bytes=<current file size>-.
func TestDownloadFile_ResumeRetryRefreshesRange(t *testing.T) {
	prefix := []byte("AAAA") // 4 bytes on disk; full object is 8 bytes
	wantFull := []byte("AAAABBBB")
	var hits int32
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		rng := r.Header.Get("Range")
		switch n {
		case 1:
			if rng != "bytes=4-" {
				t.Fatalf("first Range: want bytes=4-, got %q", rng)
			}
			w.Header().Set("Content-Range", "bytes 4-7/8")
			w.Header().Set("Content-Length", "4")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.Copy(w, io.MultiReader(bytes.NewReader([]byte("BB")), errCut{}))
			return
		case 2:
			if rng != "bytes=6-" {
				t.Fatalf("second Range after partial append: want bytes=6-, got %q", rng)
			}
			w.Header().Set("Content-Range", "bytes 6-7/8")
			w.Header().Set("Content-Length", "2")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte("BB"))
			return
		default:
			t.Fatalf("unexpected request #%d", n)
		}
	}))

	dst := filepath.Join(t.TempDir(), "partial.bin")
	if err := os.WriteFile(dst, prefix, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	written, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{
		Resume:       true,
		MaxRetries:   3,
		RetryBackoff: time.Millisecond,
	}, nil)
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if written != 4 {
		t.Errorf("written: want 4 (2+2 appended bytes this call), got %d", written)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, wantFull) {
		t.Errorf("final file: got %q want %q", got, wantFull)
	}
	if hits != 2 {
		t.Errorf("hits: want 2, got %d", hits)
	}
}

// TestDownloadFile_RangeIgnored covers the "we asked for Range but
// the server replied 200" case — falls back to a clean overwrite.
func TestDownloadFile_RangeIgnored(t *testing.T) {
	full := []byte("ABCDEFGHIJKLMNOP")
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") == "" {
			t.Fatal("client should have sent Range with --resume")
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(full)))
		_, _ = w.Write(full) // 200, ignoring Range
	}))
	dst := filepath.Join(t.TempDir(), "fallback.bin")
	if err := os.WriteFile(dst, []byte("OLD-PARTIAL"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	written, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{Resume: true}, nil)
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if written != int64(len(full)) {
		t.Errorf("written: want %d, got %d", len(full), written)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, full) {
		t.Errorf("dst should hold full body now, got %q", got)
	}
}

// TestDownloadFile_416Complete covers the "we asked for Range, server
// says 416 because we already have everything" case — return success
// with 0 written.
func TestDownloadFile_416Complete(t *testing.T) {
	full := []byte("complete")
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", len(full)))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
	}))
	dst := filepath.Join(t.TempDir(), "done.bin")
	if err := os.WriteFile(dst, full, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	written, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{Resume: true}, nil)
	if err != nil {
		t.Fatalf("416 with --resume should succeed: %v", err)
	}
	if written != 0 {
		t.Errorf("written should be 0 for already-complete, got %d", written)
	}
}

// TestDownloadFile_OverwriteUsesTmpRename confirms --overwrite goes
// through the .tmp + rename safe-write path.
func TestDownloadFile_OverwriteUsesTmpRename(t *testing.T) {
	body := []byte("new contents")
	dst := filepath.Join(t.TempDir(), "overwrite.bin")
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			t.Fatalf("--overwrite should not send Range")
		}
		_, _ = w.Write(body)
	}))
	_, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{Overwrite: true}, nil)
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("dst should be the new body: got %q", got)
	}
	if _, err := os.Stat(dst + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf(".tmp file should be cleaned up after rename")
	}
}

// TestDownloadFile_PermanentError: 4xx (non-416) should fail
// immediately, no retries.
func TestDownloadFile_PermanentError(t *testing.T) {
	var hits int32
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusForbidden)
	}))
	_, err := client.DownloadFile(context.Background(), "drive/Home/foo",
		filepath.Join(t.TempDir(), "x"), Options{MaxRetries: 5}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Errorf("4xx should not retry; hit %d times", got)
	}
}

// TestDownloadFile_TransientRetry: 503 once, then 200. The retry loop
// should swallow the transient and return success.
func TestDownloadFile_TransientRetry(t *testing.T) {
	body := []byte("ok")
	var hits int32
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write(body)
	}))
	dst := filepath.Join(t.TempDir(), "retry.bin")
	written, err := client.DownloadFile(context.Background(), "drive/Home/foo", dst, Options{
		MaxRetries:   3,
		RetryBackoff: time.Millisecond,
	}, nil)
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if written != int64(len(body)) {
		t.Errorf("written: want %d, got %d", len(body), written)
	}
	if got := atomic.LoadInt32(&hits); got != 2 {
		t.Errorf("expected 2 hits (1 transient + 1 success), got %d", got)
	}
}

// TestStreamRaw_HappyPath covers cat's wire path: GET /api/raw with
// inline=true, no Range, body streamed to stdout.
func TestStreamRaw_HappyPath(t *testing.T) {
	body := []byte("file body for cat\n")
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/raw/") {
			t.Errorf("want /api/raw/ prefix, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("inline") != "true" {
			t.Errorf("want inline=true, got %q", r.URL.Query().Get("inline"))
		}
		if r.Header.Get("X-Authorization") != "test-token-XYZ" {
			t.Errorf("missing X-Authorization")
		}
		_, _ = w.Write(body)
	}))
	var buf bytes.Buffer
	n, err := client.StreamRaw(context.Background(), "drive/Home/foo.txt", &buf)
	if err != nil {
		t.Fatalf("StreamRaw: %v", err)
	}
	if n != int64(len(body)) {
		t.Errorf("returned bytes: want %d, got %d", len(body), n)
	}
	if !bytes.Equal(buf.Bytes(), body) {
		t.Errorf("body mismatch: got %q", buf.Bytes())
	}
}

// TestStreamRaw_NonFile: server returns 400 (the raw_service.go
// "not a file" path). Surfaced as a typed *HTTPError so cat.go can
// turn it into a friendly message.
func TestStreamRaw_NonFile(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"not a file"}`)
	}))
	_, err := client.StreamRaw(context.Background(), "drive/Home/Documents", io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("want *HTTPError, got %T: %v", err, err)
	}
	if hErr.Status != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", hErr.Status)
	}
}

// TestDownloadFile_ContextCanceled confirms cancellation aborts
// without burning the retry budget.
func TestDownloadFile_ContextCanceled(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block briefly so the cancellation has time to land.
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte("x"))
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.DownloadFile(ctx, "drive/Home/foo",
		filepath.Join(t.TempDir(), "cancel.bin"), Options{}, nil)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}
