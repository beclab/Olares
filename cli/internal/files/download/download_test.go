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

// newTestClient wires a *Client up to an httptest.Server. The download
// Client itself no longer injects X-Authorization (that responsibility
// moved to the factory's refreshingTransport), so test handlers should
// assert wire shape but NOT the access-token header.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
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

// TestList_CloudDriveDataEnvelope: on awss3 / google / dropbox /
// tencent the server returns the directory contents under the
// top-level `data` field instead of `items`, with `fileSize`
// populated as the canonical byte count and `mode` / `modified`
// emitted as empty strings (see the awss3 sample the CLI now ships
// against). List must accept both shapes; here we verify the
// `data` path resolves to the same Entry slice the `items` path
// produces, so the walker / Stat don't need namespace branching.
func TestList_CloudDriveDataEnvelope(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
			"data":[
				{"name":"03 (1).avi","isDir":false,"isSymlink":false,"size":5788048,"fileSize":5788048,"mode":"","modified":"","path":"/03 (1).avi","type":""},
				{"name":"datasets","isDir":true,"isSymlink":false,"size":0,"fileSize":0,"mode":"","modified":"","path":"/datasets","type":""}
			],
			"fileExtend":"AKIA...","filePath":"/","fileType":"awss3","name":"","status_code":"SUCCESS"
		}`)
	}))
	entries, err := client.List(context.Background(), "awss3/AKIA.../")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d (%+v)", len(entries), entries)
	}
	if entries[0].Name != "03 (1).avi" || entries[0].IsDir || entries[0].Size != 5788048 {
		t.Errorf("entries[0] mismatch: %+v", entries[0])
	}
	if entries[1].Name != "datasets" || !entries[1].IsDir {
		t.Errorf("entries[1] mismatch: %+v", entries[1])
	}
}

// TestList_FileSizeFallback: when the cloud-drive server populates
// `fileSize` but not `size` (the shape the format() helper in
// awss3/filesFormat.ts treats as canonical), List must still surface
// the right byte count — the walker uses it for progress totals.
func TestList_FileSizeFallback(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
			"data":[
				{"name":"big.bin","isDir":false,"fileSize":17236328572}
			]
		}`)
	}))
	entries, err := client.List(context.Background(), "awss3/AKIA.../")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 || entries[0].Size != 17236328572 {
		t.Errorf("want 1 entry with Size 17236328572, got %+v", entries)
	}
}

// TestList_ItemsWinsWhenBothPresent: defensive — if a transitional
// server emits both `items` and `data` (legacy shape during a
// migration), the canonical files-backend `items` field wins.
// Otherwise we'd silently double-count or pick the wrong shape on
// hybrid backends.
func TestList_ItemsWinsWhenBothPresent(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
			"items":[{"name":"from-items","isDir":false,"size":1}],
			"data":[{"name":"from-data","isDir":false,"size":2}]
		}`)
	}))
	entries, err := client.List(context.Background(), "awss3/x/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "from-items" {
		t.Errorf("want items-wins, got %+v", entries)
	}
}

// TestStreamCloudFile_HappyPath: cat for cloud drives must hit
// /drive/download_sync_stream with `drive` / `cloud_file_path` /
// `name` query params, and stream the response body verbatim.
// Wire shape mirrors the LarePass web app's
// `generateDownloadUrl` helper (utils.ts in v2/{awss3,dropbox,google}).
func TestStreamCloudFile_HappyPath(t *testing.T) {
	body := []byte("cloud body bytes")
	var gotPath, gotQuery string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write(body)
	}))
	var buf bytes.Buffer
	n, err := client.StreamCloudFile(context.Background(), "awss3", "/photos/img.png", "AKIA...", &buf)
	if err != nil {
		t.Fatalf("StreamCloudFile: %v", err)
	}
	if n != int64(len(body)) {
		t.Errorf("returned bytes: want %d, got %d", len(body), n)
	}
	if !bytes.Equal(buf.Bytes(), body) {
		t.Errorf("body mismatch: got %q", buf.Bytes())
	}
	if gotPath != "/drive/download_sync_stream" {
		t.Errorf("path: want /drive/download_sync_stream, got %q", gotPath)
	}
	if !strings.Contains(gotQuery, "drive=awss3") {
		t.Errorf("query missing drive=awss3: %q", gotQuery)
	}
	if !strings.Contains(gotQuery, "name=AKIA") {
		t.Errorf("query missing name=AKIA...: %q", gotQuery)
	}
	// `cloud_file_path` must round-trip the leading '/' (and the
	// space + parens / unicode that real bucket keys contain). Decode
	// and compare.
	q, err := parseQueryStrict(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if q["cloud_file_path"] != "/photos/img.png" {
		t.Errorf("cloud_file_path: want /photos/img.png, got %q", q["cloud_file_path"])
	}
}

// TestStreamCloudFile_PreservesUnicodeAndSpaces: bucket keys
// regularly contain spaces, parens, and non-ASCII characters (the
// awss3 sample shipped with `测试上传`). The percent-encoder must
// preserve them verbatim end-to-end so the cloud-bridge worker
// looks up the right object.
func TestStreamCloudFile_PreservesUnicodeAndSpaces(t *testing.T) {
	var gotQuery string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte("ok"))
	}))
	const wantPath = "/03 (1).avi"
	if _, err := client.StreamCloudFile(context.Background(), "awss3", wantPath, "AKIA", io.Discard); err != nil {
		t.Fatalf("StreamCloudFile: %v", err)
	}
	q, err := parseQueryStrict(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if q["cloud_file_path"] != wantPath {
		t.Errorf("cloud_file_path round-trip mismatch: want %q, got %q", wantPath, q["cloud_file_path"])
	}
}

// TestStreamCloudFile_NotFound: 404 from the cloud-bridge worker
// (object missing in bucket / account de-authorised) must surface
// as *HTTPError so the cobra layer can render a friendly CTA.
func TestStreamCloudFile_NotFound(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"not found"}`)
	}))
	_, err := client.StreamCloudFile(context.Background(), "google", "/missing.txt", "acc-1", io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) || hErr.Status != http.StatusNotFound {
		t.Errorf("want *HTTPError(404), got %T: %v", err, err)
	}
}

// TestStreamCloudFile_EmptyArgs: empty driveType / cloudPath / name
// each fail fast without touching the network — those are URL query
// values the server requires, and silently sending "" would either
// 400 or (worse) match against an unrelated default.
func TestStreamCloudFile_EmptyArgs(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be hit for empty-args case")
		w.WriteHeader(http.StatusOK)
	}))
	if _, err := client.StreamCloudFile(context.Background(), "", "/p", "n", io.Discard); err == nil {
		t.Error("expected error for empty driveType")
	}
	if _, err := client.StreamCloudFile(context.Background(), "awss3", "", "n", io.Discard); err == nil {
		t.Error("expected error for empty cloudPath")
	}
	if _, err := client.StreamCloudFile(context.Background(), "awss3", "/p", "", io.Discard); err == nil {
		t.Error("expected error for empty name")
	}
}

// parseQueryStrict is a tiny wrapper over net/url's QueryUnescape
// that returns a flat map[string]string for assertion purposes; the
// values it surfaces are already URL-decoded so test cases can write
// the un-encoded form they expect.
func parseQueryStrict(raw string) (map[string]string, error) {
	out := map[string]string{}
	for _, kv := range strings.Split(raw, "&") {
		if kv == "" {
			continue
		}
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			out[kv] = ""
			continue
		}
		k := kv[:eq]
		v := kv[eq+1:]
		dv, err := decodeQueryValue(v)
		if err != nil {
			return nil, err
		}
		out[k] = dv
	}
	return out, nil
}

func decodeQueryValue(s string) (string, error) {
	// net/url is intentionally not imported here: we only need
	// percent-decoding, no query splitting; this keeps the test
	// helper free of indirect dependencies.
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '+':
			b.WriteByte(' ')
		case c == '%' && i+2 < len(s):
			h, err := strconv.ParseUint(s[i+1:i+3], 16, 8)
			if err != nil {
				return "", err
			}
			b.WriteByte(byte(h))
			i += 2
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), nil
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
