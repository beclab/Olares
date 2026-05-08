package upload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fixtureFile creates a temp file of the given size containing
// deterministic byte content (i % 256). Useful for tests that want to
// assert that the server received the right bytes for chunk N.
func fixtureFile(t *testing.T, size int64) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "fixture.bin")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := make([]byte, 4096)
	written := int64(0)
	for written < size {
		toWrite := int64(len(buf))
		if size-written < toWrite {
			toWrite = size - written
		}
		for i := int64(0); i < toWrite; i++ {
			buf[i] = byte((written + i) & 0xff)
		}
		n, err := f.Write(buf[:toWrite])
		if err != nil {
			t.Fatal(err)
		}
		written += int64(n)
	}
	return p
}

// recordedChunk captures everything we want to assert about a single
// chunk POST. Tests inspect a slice of these to verify the wire shape.
//
// X-Authorization is no longer captured here: the upload Client itself no
// longer injects the header (that responsibility moved to the factory's
// refreshingTransport), so asserting it from a stock httptest *http.Client
// would always fail.
type recordedChunk struct {
	contentRange string
	contentDisp  string
	chunkBytes   []byte
	form         map[string]string
}

// chunkRecorder is the upload-link target handler used by tests. It
// parses each multipart POST, records the relevant headers + form
// fields + chunk bytes, and (by default) returns 200 to accept the
// chunk.
//
// Tests that want to simulate retryable / permanent failures wrap
// chunkRecorder with extra logic — see TestUploadFile_Retries /
// TestUploadFile_PermanentError below.
type chunkRecorder struct {
	mu     sync.Mutex
	chunks []recordedChunk
}

func (cr *chunkRecorder) record(r *http.Request) (*recordedChunk, error) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		return nil, err
	}
	form := map[string]string{}
	for k, vs := range r.MultipartForm.Value {
		form[k] = vs[0]
	}
	var chunkBytes []byte
	if files, ok := r.MultipartForm.File["file"]; ok && len(files) > 0 {
		f, err := files[0].Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()
		chunkBytes, err = io.ReadAll(f)
		if err != nil {
			return nil, err
		}
	}
	rc := recordedChunk{
		contentRange: r.Header.Get("Content-Range"),
		contentDisp:  r.Header.Get("Content-Disposition"),
		form:         form,
		chunkBytes:   chunkBytes,
	}
	cr.mu.Lock()
	cr.chunks = append(cr.chunks, rc)
	cr.mu.Unlock()
	return &rc, nil
}

// uploadServerOpts plumbs per-test customization into uploadServer
// without proliferating constructor variants.
type uploadServerOpts struct {
	uploadedBytes  int64                                                    // probe response
	uploadHandler  func(*chunkRecorder, http.ResponseWriter, *http.Request) // override chunk POST
	uploadLinkPath string                                                   // override default link path
}

// uploadServer wires up an httptest.Server that knows how to answer
// the three endpoints UploadFile depends on:
//
//   - GET /api/nodes/                     → returns one node ("n")
//   - GET /upload/upload-link/<n>/        → returns a path the chunk
//     POSTs target
//   - GET /upload/file-uploaded-bytes/<n>/ → returns opts.uploadedBytes
//   - POST <upload-link-path>             → routed through opts.uploadHandler
//
// The recorded chunks are returned via the *chunkRecorder so tests
// can assert on them post-hoc.
func uploadServer(t *testing.T, opts uploadServerOpts) (*httptest.Server, *chunkRecorder) {
	t.Helper()
	cr := &chunkRecorder{}
	uploadLink := opts.uploadLinkPath
	if uploadLink == "" {
		uploadLink = "/seafhttp/upload-aj/repo-1/"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/nodes/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"data":{"nodes":[{"name":"n"}]}}`)
	})
	mux.HandleFunc("/upload/upload-link/n/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, uploadLink)
	})
	mux.HandleFunc("/upload/file-uploaded-bytes/n/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"uploadedBytes":%d}`, opts.uploadedBytes)
	})
	mux.HandleFunc(uploadLink, func(w http.ResponseWriter, r *http.Request) {
		if opts.uploadHandler != nil {
			opts.uploadHandler(cr, w, r)
			return
		}
		if _, err := cr.record(r); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, cr
}

// TestUploadFile_Multichunk: the happy path with a multi-chunk file.
// Asserts that:
//
//   - chunk count + chunk bytes are right
//   - resumableChunkNumber is 1-indexed
//   - resumableTotalChunks / resumableTotalSize / resumableCurrentChunkSize
//     all line up with the file size + chunk size
//   - Content-Range uses INCLUSIVE end-byte semantics (matches
//     resumejs.ts setHeaders())
//   - the file's exact bytes round-trip
func TestUploadFile_Multichunk(t *testing.T) {
	const chunkSize = 1024
	// 2.5 chunks: covers full + full + partial.
	fileSize := int64(2*chunkSize + chunkSize/2)
	local := fixtureFile(t, fileSize)
	srv, recorder := uploadServer(t, uploadServerOpts{})

	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath:    local,
		Node:         "n",
		ParentDir:    "/drive/Home/Docs/",
		RemoteName:   "f.bin",
		RelativePath: "f.bin",
		ChunkSize:    chunkSize,
	}, nil); err != nil {
		t.Fatal(err)
	}

	if got, want := len(recorder.chunks), 3; got != want {
		t.Fatalf("got %d chunks, want %d", got, want)
	}
	expectedRanges := []string{
		fmt.Sprintf("bytes 0-%d/%d", chunkSize-1, fileSize),
		fmt.Sprintf("bytes %d-%d/%d", chunkSize, 2*chunkSize-1, fileSize),
		fmt.Sprintf("bytes %d-%d/%d", 2*chunkSize, fileSize-1, fileSize),
	}
	wantSizes := []int64{chunkSize, chunkSize, chunkSize / 2}
	wholeFile := readFile(t, local)
	gotFile := []byte{}
	for i, ck := range recorder.chunks {
		if ck.contentRange != expectedRanges[i] {
			t.Errorf("chunk %d: Content-Range = %q, want %q",
				i, ck.contentRange, expectedRanges[i])
		}
		if int64(len(ck.chunkBytes)) != wantSizes[i] {
			t.Errorf("chunk %d: byte len = %d, want %d",
				i, len(ck.chunkBytes), wantSizes[i])
		}
		// resumableChunkNumber is 1-indexed.
		if got := ck.form["resumableChunkNumber"]; got != strconv.Itoa(i+1) {
			t.Errorf("chunk %d: resumableChunkNumber = %q", i, got)
		}
		if got := ck.form["resumableCurrentChunkSize"]; got != strconv.FormatInt(wantSizes[i], 10) {
			t.Errorf("chunk %d: resumableCurrentChunkSize = %q", i, got)
		}
		if got := ck.form["resumableTotalSize"]; got != strconv.FormatInt(fileSize, 10) {
			t.Errorf("chunk %d: resumableTotalSize = %q", i, got)
		}
		if got := ck.form["resumableTotalChunks"]; got != "3" {
			t.Errorf("chunk %d: resumableTotalChunks = %q", i, got)
		}
		if got := ck.form["resumableFilename"]; got != "f.bin" {
			t.Errorf("chunk %d: resumableFilename = %q", i, got)
		}
		if got := ck.form["resumableRelativePath"]; got != "f.bin" {
			t.Errorf("chunk %d: resumableRelativePath = %q", i, got)
		}
		if got := ck.form["parent_dir"]; got != "/drive/Home/Docs/" {
			t.Errorf("chunk %d: parent_dir = %q", i, got)
		}
		if got := ck.form["driveType"]; got != "Drive" {
			t.Errorf("chunk %d: driveType = %q", i, got)
		}
		gotFile = append(gotFile, ck.chunkBytes...)
	}
	if !bytes.Equal(gotFile, wholeFile) {
		t.Errorf("reassembled bytes don't match source file")
	}
}

// TestUploadFile_ResumesFromServerOffset: when /file-uploaded-bytes/
// reports a non-zero count, UploadFile must skip the
// already-uploaded-chunks and start from the next one. The bytes
// reported by the server are floored to a chunk boundary (matches the
// web app's Math.floor(uploadedBytes / chunkSize) trick: it's safe to
// re-upload the unaligned tail because chunks are deterministic).
func TestUploadFile_ResumesFromServerOffset(t *testing.T) {
	const chunkSize = 1024
	fileSize := int64(3 * chunkSize)
	local := fixtureFile(t, fileSize)
	srv, recorder := uploadServer(t, uploadServerOpts{
		// Server reports "I have 1.5 chunks worth of bytes." The
		// floor-to-chunk-boundary logic should still send chunks 2 + 3
		// (i.e. resumableChunkNumber 2 + 3, 0-based offsets 1 + 2).
		uploadedBytes: int64(chunkSize + chunkSize/2),
	})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath: local, Node: "n",
		ParentDir: "/drive/Home/", RemoteName: "f.bin", RelativePath: "f.bin",
		ChunkSize: chunkSize,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if got := len(recorder.chunks); got != 2 {
		t.Fatalf("got %d chunks, want 2 (resume should skip chunk 1)", got)
	}
	want := []string{
		fmt.Sprintf("bytes %d-%d/%d", chunkSize, 2*chunkSize-1, fileSize),
		fmt.Sprintf("bytes %d-%d/%d", 2*chunkSize, 3*chunkSize-1, fileSize),
	}
	for i, ck := range recorder.chunks {
		if ck.contentRange != want[i] {
			t.Errorf("chunk %d Content-Range = %q, want %q", i, ck.contentRange, want[i])
		}
		if got := ck.form["resumableChunkNumber"]; got != strconv.Itoa(int(int64(i)+2)) {
			t.Errorf("chunk %d resumableChunkNumber = %q, want %d", i, got, i+2)
		}
	}
}

// TestUploadFile_ServerHasFullFile: when uploadedBytes >= fileSize,
// nothing needs to be sent. UploadFile should return nil without any
// chunk POST.
func TestUploadFile_ServerHasFullFile(t *testing.T) {
	const chunkSize = 1024
	fileSize := int64(2 * chunkSize)
	local := fixtureFile(t, fileSize)
	srv, recorder := uploadServer(t, uploadServerOpts{uploadedBytes: fileSize})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath: local, Node: "n",
		ParentDir: "/drive/Home/", RemoteName: "f.bin", RelativePath: "f.bin",
		ChunkSize: chunkSize,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if got := len(recorder.chunks); got != 0 {
		t.Errorf("expected 0 chunks, got %d", got)
	}
}

// TestUploadFile_Retries: a transient 502 should retry, eventually
// succeed, and not abort the upload.
func TestUploadFile_Retries(t *testing.T) {
	const chunkSize = 512
	fileSize := int64(chunkSize)
	local := fixtureFile(t, fileSize)
	var attempts int32
	srv, _ := uploadServer(t, uploadServerOpts{
		uploadHandler: func(cr *chunkRecorder, w http.ResponseWriter, r *http.Request) {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				http.Error(w, "transient", http.StatusBadGateway)
				return
			}
			if _, err := cr.record(r); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(http.StatusOK)
		},
	})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath: local, Node: "n",
		ParentDir: "/drive/Home/", RemoteName: "f.bin", RelativePath: "f.bin",
		ChunkSize:    chunkSize,
		MaxRetries:   3,
		RetryBackoff: time.Millisecond, // keep the test fast
	}, nil); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("attempts = %d, want 3", got)
	}
}

// TestUploadFile_PermanentError: a 400 should NOT trigger a retry, so
// the chunk handler should be hit exactly once before UploadFile
// surfaces the permanent error.
func TestUploadFile_PermanentError(t *testing.T) {
	const chunkSize = 512
	fileSize := int64(chunkSize)
	local := fixtureFile(t, fileSize)
	var attempts int32
	srv, _ := uploadServer(t, uploadServerOpts{
		uploadHandler: func(_ *chunkRecorder, w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&attempts, 1)
			http.Error(w, "bad", http.StatusBadRequest)
		},
	})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	_, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath: local, Node: "n",
		ParentDir: "/drive/Home/", RemoteName: "f.bin", RelativePath: "f.bin",
		ChunkSize:    chunkSize,
		MaxRetries:   3,
		RetryBackoff: time.Millisecond,
	}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("expected *HTTPError, got %T (%v)", err, err)
	}
	if hErr.Status != 400 {
		t.Errorf("status = %d, want 400", hErr.Status)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("attempts = %d, want 1 (permanent error must not retry)", got)
	}
}

// TestUploadFile_EmptyFile: 0-byte files go through the
// /api/resources POST (CreateEmptyFile), NOT through the chunk
// pipeline. Resumable.js can't represent a 0-byte chunk so we have to
// take the same detour as the web app.
func TestUploadFile_EmptyFile(t *testing.T) {
	local := fixtureFile(t, 0)
	chunkHits := int32(0)
	emptyHit := int32(0)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/nodes/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"data":{"nodes":[{"name":"n"}]}}`)
	})
	mux.HandleFunc("/upload/upload-link/n/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "/upload-target/")
	})
	mux.HandleFunc("/upload/file-uploaded-bytes/n/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"uploadedBytes":0}`)
	})
	mux.HandleFunc("/upload-target/", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&chunkHits, 1)
		w.WriteHeader(200)
	})
	// /api/resources/drive/Home/.../empty.bin (no trailing slash → empty file create).
	mux.HandleFunc("/api/resources/drive/Home/Docs/empty.bin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("empty file create: method = %s", r.Method)
		}
		atomic.AddInt32(&emptyHit, 1)
		w.WriteHeader(200)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath: local, Node: "n",
		ParentDir: "/drive/Home/Docs/", RemoteName: "empty.bin", RelativePath: "empty.bin",
		ChunkSize: 1024,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&emptyHit); got != 1 {
		t.Errorf("empty-file create hit = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&chunkHits); got != 0 {
		t.Errorf("chunk handler hit = %d, want 0 (empty file should bypass)", got)
	}
}

// TestUploadFile_SyncMultichunk: end-to-end check that the chunk POST
// for a Sync upload sends the inside-repo `parent_dir` (NOT the
// /sync/<repo>/... API form). This is a regression test for the
// HTTP 500 we got from seafhttp/upload-aj when the CLI sent the API
// form as the chunk's parent_dir form field — Seafile resolves
// parent_dir relative to the repo root pinned by the upload token,
// so the API prefix would map to a non-existent dir inside the repo.
func TestUploadFile_SyncMultichunk(t *testing.T) {
	const chunkSize = 1024
	fileSize := int64(2 * chunkSize)
	local := fixtureFile(t, fileSize)
	srv, recorder := uploadServer(t, uploadServerOpts{})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath:      local,
		Node:           "n",
		DriveType:      "Sync",
		ParentDir:      "/sync/repo-1/docs/",
		ChunkParentDir: "/docs/",
		RemoteName:     "f.bin",
		RelativePath:   "f.bin",
		ChunkSize:      chunkSize,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if got := len(recorder.chunks); got != 2 {
		t.Fatalf("got %d chunks, want 2", got)
	}
	for i, ck := range recorder.chunks {
		if got := ck.form["parent_dir"]; got != "/docs/" {
			t.Errorf("chunk %d: parent_dir = %q, want %q (inside-repo path)",
				i, got, "/docs/")
		}
		if got := ck.form["driveType"]; got != "Sync" {
			t.Errorf("chunk %d: driveType = %q, want %q", i, got, "Sync")
		}
	}
}

func TestUploadFile_EmptyFile_SyncPath(t *testing.T) {
	local := fixtureFile(t, 0)
	emptyHit := int32(0)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/resources/sync/repo-1/docs/empty.bin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("empty file create: method = %s", r.Method)
		}
		atomic.AddInt32(&emptyHit, 1)
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath: local, Node: "n",
		DriveType: "Sync",
		ParentDir: "/sync/repo-1/docs/", RemoteName: "empty.bin", RelativePath: "empty.bin",
		ChunkSize: 1024,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&emptyHit); got != 1 {
		t.Errorf("empty-file create hit = %d, want 1", got)
	}
}

// TestUploadFile_FilesBackendNamespaces is the table-driven smoke
// test for the namespaces that share the files-backend chunk
// protocol (Drive/Data/Cache/External + the cloud drives whose v2
// dataAPIs inherit DriveDataAPI's chunk pipeline without overrides:
// Awss3, Google, Dropbox). Unlike Sync — whose chunk endpoint is
// Seafile's seafhttp/upload-aj and demands an inside-repo
// `parent_dir` (covered separately by
// TestUploadFile_SyncMultichunk) — these all let
// ChunkParentDir == ParentDir.
//
// What this test pins down:
//
//   - The chunk POST's `parent_dir` form field equals the API-form
//     ParentDir verbatim (no Sync-style stripping). That's what
//     keeps GET /upload/file-uploaded-bytes/ and the chunk POST in
//     agreement, which is the requirement for resume to find an
//     existing partial upload.
//   - The chunk POST's `driveType` form field carries the
//     web-app-style namespace literal. The server side is permissive
//     here, but mirroring the wire output makes the CLI sessions
//     visually indistinguishable from web-app sessions in server
//     logs — useful for debugging and for any future server change
//     that does start branching on driveType.
//
// Adding a 5th files-backend-managed namespace? Add a row here and
// to TestUploadRootAndDriveType in cmd/ctl/files; both tests should
// fail until the dispatcher is wired up.
func TestUploadFile_FilesBackendNamespaces(t *testing.T) {
	const chunkSize = 1024
	fileSize := int64(2 * chunkSize)

	cases := []struct {
		name      string
		parentDir string
		driveType string
	}{
		{
			name:      "drive Data",
			parentDir: "/drive/Data/Backups/",
			driveType: "Data",
		},
		{
			name:      "cache",
			parentDir: "/cache/node-1/AppName/data/",
			driveType: "Cache",
		},
		{
			name:      "external",
			parentDir: "/external/node-1/hdd1/Movies/",
			driveType: "External",
		},
		{
			name:      "awss3 (cloud drive, regular multipart pipeline)",
			parentDir: "/awss3/account-x/bucket/Backups/",
			driveType: "Awss3",
		},
		{
			name:      "google (cloud drive, regular multipart pipeline)",
			parentDir: "/google/account-x/Documents/",
			driveType: "Google",
		},
		{
			name:      "dropbox (cloud drive, regular multipart pipeline)",
			parentDir: "/dropbox/account-x/Notes/",
			driveType: "Dropbox",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			local := fixtureFile(t, fileSize)
			srv, recorder := uploadServer(t, uploadServerOpts{})
			c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
			if _, err := c.UploadFile(context.Background(), UploadOpts{
				LocalPath: local,
				Node:      "n",
				DriveType: tc.driveType,
				// ChunkParentDir intentionally omitted — normalize()
				// defaults it to ParentDir for the
				// files-backend-managed namespaces, and the test
				// asserts that default is what hits the wire.
				ParentDir:    tc.parentDir,
				RemoteName:   "f.bin",
				RelativePath: "f.bin",
				ChunkSize:    chunkSize,
			}, nil); err != nil {
				t.Fatal(err)
			}
			if got := len(recorder.chunks); got != 2 {
				t.Fatalf("got %d chunks, want 2", got)
			}
			for i, ck := range recorder.chunks {
				if got := ck.form["parent_dir"]; got != tc.parentDir {
					t.Errorf("chunk %d: parent_dir = %q, want %q",
						i, got, tc.parentDir)
				}
				if got := ck.form["driveType"]; got != tc.driveType {
					t.Errorf("chunk %d: driveType = %q, want %q",
						i, got, tc.driveType)
				}
			}
		})
	}
}

// TestUploadFile_CloudTaskIDFromFinalChunk: when the destination is a
// cloud drive (awss3 / google / dropbox) the LAST chunk's response
// body carries the stage-2 cloud-transfer taskId as
// `[{"taskId":"<id>"}]`. UploadFile must parse it out and surface it
// in UploadResult so the cobra layer can drive WaitCloudTask. This
// is the regression test for the two-stage upload protocol; without
// it a future refactor that drops the response body could silently
// break cloud uploads (the chunked POST would still succeed, but
// the file would never land in the user's actual bucket).
//
// Mirrors resumejs.ts onFileUploadSuccess L591-606's parsing
// behavior: `JSON.parse(message)` → expect array → first element
// `taskId` is the handle. Anything that doesn't fit collapses to ""
// (covered by TestParseFinalChunkTaskID below).
func TestUploadFile_CloudTaskIDFromFinalChunk(t *testing.T) {
	const chunkSize = 1024
	fileSize := int64(2 * chunkSize)
	local := fixtureFile(t, fileSize)
	const wantTaskID = "task-abc-123"

	var chunkCount int32
	srv, _ := uploadServer(t, uploadServerOpts{
		uploadHandler: func(cr *chunkRecorder, w http.ResponseWriter, r *http.Request) {
			n := atomic.AddInt32(&chunkCount, 1)
			if _, err := cr.record(r); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			// Drive emits an empty body; cloud drives stuff a JSON
			// array with the stage-2 taskId on the FINAL chunk.
			// The test fixture has totalChunks == 2, so write the
			// taskId body on chunk 2.
			if n == 2 {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `[{"taskId":"%s","node":"n"}]`, wantTaskID)
				return
			}
			w.WriteHeader(http.StatusOK)
		},
	})

	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	res, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath:    local,
		Node:         "n",
		DriveType:    "Awss3",
		ParentDir:    "/awss3/account-x/bucket/Backups/",
		RemoteName:   "f.bin",
		RelativePath: "f.bin",
		ChunkSize:    chunkSize,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.CloudTaskID != wantTaskID {
		t.Errorf("CloudTaskID = %q, want %q", res.CloudTaskID, wantTaskID)
	}
}

// TestUploadFile_NoCloudTaskIDForDrive: Drive's chunk endpoint
// returns an empty body. Make sure we don't synthesise a phantom
// taskId from "" — the cobra layer's WaitCloudTask call is gated
// on CloudTaskID != "", so a false positive here would cause
// drive uploads to start polling /api/task/ for nothing.
func TestUploadFile_NoCloudTaskIDForDrive(t *testing.T) {
	const chunkSize = 1024
	fileSize := int64(chunkSize)
	local := fixtureFile(t, fileSize)
	srv, _ := uploadServer(t, uploadServerOpts{})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	res, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath: local, Node: "n",
		ParentDir: "/drive/Home/", RemoteName: "f.bin", RelativePath: "f.bin",
		ChunkSize: chunkSize,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.CloudTaskID != "" {
		t.Errorf("CloudTaskID = %q, want \"\" (Drive uploads are single-stage)",
			res.CloudTaskID)
	}
}

// TestParseFinalChunkTaskID covers the ways the final-chunk response
// body can be malformed-or-empty without us making the upload look
// "failed". The principle is: the chunk POST already returned 2xx,
// so the upload itself succeeded; if we can't extract a stage-2
// taskId from the body, return "" and let the caller skip
// WaitCloudTask. Erroring out here would turn a successful upload
// into a CLI failure for the user.
func TestParseFinalChunkTaskID(t *testing.T) {
	cases := []struct {
		name string
		body []byte
		want string
	}{
		{"happy path", []byte(`[{"taskId":"abc-123","node":"n"}]`), "abc-123"},
		{"happy with whitespace", []byte("\n  [{\"taskId\":\"x\"}]\n"), "x"},
		{"empty body (Drive)", []byte(``), ""},
		{"empty array", []byte(`[]`), ""},
		{"object instead of array", []byte(`{"taskId":"x"}`), ""},
		{"missing taskId key", []byte(`[{"foo":"bar"}]`), ""},
		{"empty taskId value", []byte(`[{"taskId":""}]`), ""},
		{"malformed JSON", []byte(`[{"taskId":`), ""},
		{"non-JSON text", []byte(`OK`), ""},
		{"first element wins", []byte(`[{"taskId":"first"},{"taskId":"second"}]`), "first"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseFinalChunkTaskID(tc.body); got != tc.want {
				t.Errorf("parseFinalChunkTaskID(%q) = %q, want %q",
					tc.body, got, tc.want)
			}
		})
	}
}

// TestUploadFile_FolderRelativePath: when a file lives in a subdir
// of the upload root (RelativePath has '/'), the per-chunk
// `relative_path` form field carries the directory prefix WITH a
// trailing slash — same shape resumejs.ts onChunkingComplete emits.
func TestUploadFile_FolderRelativePath(t *testing.T) {
	const chunkSize = 256
	local := fixtureFile(t, chunkSize)
	srv, recorder := uploadServer(t, uploadServerOpts{})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	if _, err := c.UploadFile(context.Background(), UploadOpts{
		LocalPath:    local,
		Node:         "n",
		ParentDir:    "/drive/Home/X/",
		RemoteName:   "foo.txt",
		RelativePath: "mydir/sub/foo.txt",
		ChunkSize:    chunkSize,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if len(recorder.chunks) != 1 {
		t.Fatalf("got %d chunks, want 1", len(recorder.chunks))
	}
	ck := recorder.chunks[0]
	if got := ck.form["relative_path"]; got != "mydir/sub/" {
		t.Errorf("relative_path = %q, want %q", got, "mydir/sub/")
	}
	if got := ck.form["resumableRelativePath"]; got != "mydir/sub/foo.txt" {
		t.Errorf("resumableRelativePath = %q, want %q", got, "mydir/sub/foo.txt")
	}
}

// TestUploadFile_ContextCancel: cancelling ctx mid-retry should bail
// out promptly with ctx.Err(), NOT keep grinding through the retry
// budget.
func TestUploadFile_ContextCancel(t *testing.T) {
	const chunkSize = 256
	local := fixtureFile(t, chunkSize)
	srv, _ := uploadServer(t, uploadServerOpts{
		uploadHandler: func(_ *chunkRecorder, w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "transient", http.StatusBadGateway)
		},
	})
	c := &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	_, err := c.UploadFile(ctx, UploadOpts{
		LocalPath: local, Node: "n",
		ParentDir: "/drive/Home/", RemoteName: "f.bin", RelativePath: "f.bin",
		ChunkSize:    chunkSize,
		MaxRetries:   1000,
		RetryBackoff: 50 * time.Millisecond,
	}, nil)
	if err == nil {
		t.Fatal("expected error after cancel")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v; want context.Canceled", err)
	}
}

// readFile reads a fixture back so tests can compare the round-trip.
func readFile(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Sanity check that ParseMultipartForm sees the chunk under field
// name "file" — guards against accidental rename of fileParameterName.
func TestBuildChunkBody_FileFieldNameIsFile(t *testing.T) {
	rdr, ct, err := buildChunkBody(UploadOpts{
		ChunkSize: 1024, RemoteName: "x.bin", RelativePath: "x.bin",
		ParentDir: "/drive/Home/",
		DriveType: "Sync",
	}, chunkUploadCtx{
		ChunkIndex: 0, TotalChunks: 1, ChunkLen: 4, StartByte: 0, FileSize: 4,
		Identifier: "id", MimeType: "application/octet-stream",
		ChunkContents: []byte{1, 2, 3, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mediaType(ct)
	if err != nil {
		t.Fatal(err)
	}
	mr := multipart.NewReader(rdr, params["boundary"])
	sawFile := false
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if p.FormName() == "file" {
			sawFile = true
			b, _ := io.ReadAll(p)
			if !bytes.Equal(b, []byte{1, 2, 3, 4}) {
				t.Errorf("file part = %v", b)
			}
		}
	}
	if !sawFile {
		t.Error("did not find a part named 'file' in the multipart body")
	}

	// Build and parse again to assert custom query for a Sync upload:
	//   - driveType form field is "Sync"
	//   - parent_dir form field is the inside-repo path
	//     ("/docs/"), NOT the API-form "/sync/<repo>/docs/" — Seafile's
	//     seafhttp/upload-aj resolves parent_dir relative to the repo
	//     root the upload token already pinned, so sending the
	//     `/sync/<repo>/...` prefix would 500 with "parent dir doesn't
	//     exist" inside the repo.
	rdr2, ct2, err := buildChunkBody(UploadOpts{
		ChunkSize: 1024, RemoteName: "x.bin", RelativePath: "x.bin",
		ParentDir:      "/sync/repo-1/docs/",
		ChunkParentDir: "/docs/",
		DriveType:      "Sync",
	}, chunkUploadCtx{
		ChunkIndex: 0, TotalChunks: 1, ChunkLen: 4, StartByte: 0, FileSize: 4,
		Identifier: "id", MimeType: "application/octet-stream",
		ChunkContents: []byte{1, 2, 3, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, params2, err := mediaType(ct2)
	if err != nil {
		t.Fatal(err)
	}
	mr2 := multipart.NewReader(rdr2, params2["boundary"])
	var gotDriveType, gotParentDir string
	for {
		p, err := mr2.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch p.FormName() {
		case "driveType":
			b, _ := io.ReadAll(p)
			gotDriveType = string(b)
		case "parent_dir":
			b, _ := io.ReadAll(p)
			gotParentDir = string(b)
		}
	}
	if gotDriveType != "Sync" {
		t.Errorf("driveType = %q, want %q", gotDriveType, "Sync")
	}
	if gotParentDir != "/docs/" {
		t.Errorf("parent_dir = %q, want %q (inside-repo path; NOT the /sync/<repo>/... API form)",
			gotParentDir, "/docs/")
	}
}

// mediaType is a tiny stand-in for mime.ParseMediaType to avoid the
// extra import in the test (we only need the boundary).
func mediaType(s string) (string, map[string]string, error) {
	idx := strings.Index(s, ";")
	if idx < 0 {
		return s, map[string]string{}, nil
	}
	out := map[string]string{}
	for _, kv := range strings.Split(s[idx+1:], ";") {
		kv = strings.TrimSpace(kv)
		eq := strings.Index(kv, "=")
		if eq < 0 {
			continue
		}
		k := kv[:eq]
		v := strings.Trim(kv[eq+1:], "\"")
		v, _ = url.QueryUnescape(v)
		out[k] = v
	}
	return s[:idx], out, nil
}
