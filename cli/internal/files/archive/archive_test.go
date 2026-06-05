package archive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient is the same httptest harness shape every other
// per-package test in this repo uses (see cp_test.go's
// newTestClient): stand up a server, hand the caller a Client
// whose BaseURL points at it, let the test inspect what landed
// on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// ============================================================
// format.go
// ============================================================

func TestIsSupportedFormat(t *testing.T) {
	for _, f := range SupportedFormats {
		if !IsSupportedFormat(f) {
			t.Errorf("IsSupportedFormat(%q) = false; want true", f)
		}
	}
	if IsSupportedFormat("rar") {
		t.Error("IsSupportedFormat(rar): want false")
	}
	// Case-insensitive.
	if !IsSupportedFormat("ZIP") {
		t.Error("IsSupportedFormat(ZIP): want true")
	}
}

func TestSupportsPassword(t *testing.T) {
	for _, ok := range []string{"zip", "7z", "ZIP", "7Z"} {
		if !SupportsPassword(ok) {
			t.Errorf("SupportsPassword(%q) = false; want true", ok)
		}
	}
	for _, no := range []string{"tar", "tar.gz", "gzip", "xz"} {
		if SupportsPassword(no) {
			t.Errorf("SupportsPassword(%q) = true; want false", no)
		}
	}
}

func TestSupportsMultiVolume(t *testing.T) {
	if !SupportsMultiVolume("zip") || !SupportsMultiVolume("7z") {
		t.Error("zip/7z should both support multi-volume")
	}
	if SupportsMultiVolume("tar") {
		t.Error("tar should NOT support multi-volume")
	}
}

func TestParseConflict(t *testing.T) {
	cases := []struct {
		in   string
		want Conflict
		err  bool
	}{
		{"", ConflictRename, false},
		{"rename", ConflictRename, false},
		{"OVERWRITE", ConflictOverwrite, false},
		{"  skip  ", ConflictSkip, false},
		{"replace", "", true}, // not a valid value
	}
	for _, c := range cases {
		got, err := ParseConflict(c.in)
		if (err != nil) != c.err {
			t.Errorf("ParseConflict(%q): err=%v want err=%v", c.in, err, c.err)
			continue
		}
		if !c.err && got != c.want {
			t.Errorf("ParseConflict(%q): got %q want %q", c.in, got, c.want)
		}
	}
}

func TestValidateLevel(t *testing.T) {
	for _, ok := range []int{-1, 0, 1, 5, 9} {
		if err := ValidateLevel(ok); err != nil {
			t.Errorf("ValidateLevel(%d): %v", ok, err)
		}
	}
	if err := ValidateLevel(10); err == nil {
		t.Error("ValidateLevel(10): want error, got nil")
	}
}

func TestValidateFormat(t *testing.T) {
	if err := ValidateFormat("", "compress"); err == nil {
		t.Error("ValidateFormat(empty): want error")
	}
	if err := ValidateFormat("rar", "compress"); err == nil {
		t.Error("ValidateFormat(rar): want error")
	}
	if err := ValidateFormat("zip", "compress"); err != nil {
		t.Errorf("ValidateFormat(zip): %v", err)
	}
}

func TestFormatFromExtension(t *testing.T) {
	// Longest-suffix-first must dominate over shorter matches.
	cases := []struct{ name, want string }{
		{"foo.tar.gz", "tar.gz"},
		{"foo.tar.bz2", "tar.bz2"},
		{"foo.tar.xz", "tar.xz"},
		{"foo.tgz", "tgz"},
		{"foo.tar", "tar"},
		{"foo.7z", "7z"},
		{"foo.zip", "zip"},
		{"foo.gz", "gzip"},
		{"foo.bz2", "bzip2"},
		{"foo.xz", "xz"},
		{"FOO.ZIP", "zip"},
		{"foo.rar", ""}, // unknown
		{"foo", ""},
	}
	for _, c := range cases {
		if got := FormatFromExtension(c.name); got != c.want {
			t.Errorf("FormatFromExtension(%q): got %q want %q", c.name, got, c.want)
		}
	}
}

// ============================================================
// wire.go
// ============================================================

func TestBuildWirePath(t *testing.T) {
	cases := []struct {
		ft, ex, sub, want string
	}{
		{"drive", "Home", "/Documents/foo.pdf", "/drive/Home/Documents/foo.pdf"},
		{"drive", "Home", "/Photos/", "/drive/Home/Photos/"},
		{"drive", "Home", "", "/drive/Home/"},
		{"drive", "Home", "Documents", "/drive/Home/Documents"}, // missing leading slash auto-fixed
	}
	for _, c := range cases {
		if got := BuildWirePath(c.ft, c.ex, c.sub); got != c.want {
			t.Errorf("BuildWirePath(%q,%q,%q): got %q want %q",
				c.ft, c.ex, c.sub, got, c.want)
		}
	}
}

// ============================================================
// compress.go
// ============================================================

func TestCompress_Happy(t *testing.T) {
	cli, srv := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wire shape sanity.
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s want POST", r.Method)
		}
		if got, want := r.URL.Path, "/api/archive/node-a/compress"; got != want {
			t.Errorf("path: got %q want %q", got, want)
		}
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type: got %q", ct)
		}
		if pw := r.Header.Get("X-Archive-Password"); pw != "s3cret" {
			t.Errorf("password header: got %q want s3cret", pw)
		}
		var body compressRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if len(body.Sources) != 2 || body.Destination != "/drive/Home/out.zip" {
			t.Errorf("body shape: %+v", body)
		}
		if body.Format != "zip" {
			t.Errorf("format: got %q", body.Format)
		}
		if body.Level == nil || *body.Level != 5 {
			t.Errorf("level: %v", body.Level)
		}
		if body.VolumeSizeMB == nil || *body.VolumeSizeMB != 100 {
			t.Errorf("volumeSizeMB: %v", body.VolumeSizeMB)
		}
		if body.Conflict != "rename" {
			t.Errorf("conflict: %q", body.Conflict)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","task_id":"task-123"}`))
	}))
	_ = srv

	taskID, err := cli.Compress(context.Background(), CompressOptions{
		Sources:      []string{"/drive/Home/a.pdf", "/drive/Home/b.pdf"},
		Destination:  "/drive/Home/out.zip",
		Format:       "zip",
		Level:        5,
		VolumeSizeMB: 100,
		Conflict:     ConflictRename,
		Node:         "node-a",
	}, "s3cret")
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if taskID != "task-123" {
		t.Errorf("taskID: got %q want task-123", taskID)
	}
}

func TestCompress_LevelOmittedWhenSentinel(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Confirm "level" key is absent in raw JSON when Level == -1.
		if strings.Contains(string(body), `"level"`) {
			t.Errorf("level should be omitted; body=%s", body)
		}
		// VolumeSizeMB == 0 should also be omitted.
		if strings.Contains(string(body), `"volumeSizeMB"`) {
			t.Errorf("volumeSizeMB should be omitted; body=%s", body)
		}
		_, _ = w.Write([]byte(`{"code":0,"task_id":"t"}`))
	}))
	_, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/drive/Home/a.pdf"},
		Destination: "/drive/Home/out.zip",
		Format:      "zip",
		Level:       -1, // sentinel
		Node:        "node-a",
	}, "")
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
}

func TestCompress_ServerNonZeroCode(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"msg":"invalid path"}`))
	}))
	_, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/drive/Home/a.pdf"},
		Destination: "/drive/Home/out.zip",
		Format:      "zip",
		Node:        "node-a",
	}, "")
	if err == nil || !strings.Contains(err.Error(), "invalid path") {
		t.Errorf("want server-message error, got %v", err)
	}
}

func TestCompress_EmptyTaskID(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	_, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/drive/Home/a.pdf"},
		Destination: "/drive/Home/out.zip",
		Format:      "zip",
		Node:        "node-a",
	}, "")
	if err == nil || !strings.Contains(err.Error(), "no task_id") {
		t.Errorf("want missing-task_id error, got %v", err)
	}
}

func TestCompress_HTTPError(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad format"}`))
	}))
	_, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/drive/Home/a.pdf"},
		Destination: "/drive/Home/out.zip",
		Format:      "zip",
		Node:        "node-a",
	}, "")
	var h *HTTPError
	if !errors.As(err, &h) {
		t.Fatalf("want *HTTPError, got %v", err)
	}
	if h.Status != http.StatusBadRequest {
		t.Errorf("status: got %d", h.Status)
	}
}

func TestCompress_ValidationRejects(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be hit; validation must fail client-side")
	}))

	// Missing Node.
	if _, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/x"},
		Destination: "/y",
		Format:      "zip",
	}, ""); err == nil {
		t.Error("want error for empty Node")
	}
	// Missing Sources.
	if _, err := cli.Compress(context.Background(), CompressOptions{
		Destination: "/y",
		Format:      "zip",
		Node:        "n",
	}, ""); err == nil {
		t.Error("want error for missing sources")
	}
	// Unsupported format.
	if _, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/x"},
		Destination: "/y",
		Format:      "rar",
		Node:        "n",
	}, ""); err == nil {
		t.Error("want error for rar format")
	}
	// Password on non-passwordable format.
	if _, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/x"},
		Destination: "/y",
		Format:      "tar",
		Node:        "n",
	}, "pw"); err == nil {
		t.Error("want error: tar does not accept passwords")
	}
	// volumeSizeMB on non-multi-volume format.
	if _, err := cli.Compress(context.Background(), CompressOptions{
		Sources:      []string{"/x"},
		Destination:  "/y",
		Format:       "tar.gz",
		VolumeSizeMB: 100,
		Node:         "n",
	}, ""); err == nil {
		t.Error("want error: tar.gz does not accept --volume-size-mb")
	}
	// Level out of range.
	if _, err := cli.Compress(context.Background(), CompressOptions{
		Sources:     []string{"/x"},
		Destination: "/y",
		Format:      "zip",
		Level:       99,
		Node:        "n",
	}, ""); err == nil {
		t.Error("want error: level 99 out of range")
	}
}

// ============================================================
// extract.go
// ============================================================

func TestExtract_Happy(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/archive/node-a/extract" {
			t.Errorf("path: %q", r.URL.Path)
		}
		var body extractRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Source != "/drive/Home/out.zip" || body.Destination != "/drive/Home/unpacked/" {
			t.Errorf("body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"success","task_id":"task-456"}`))
	}))
	taskID, err := cli.Extract(context.Background(), ExtractOptions{
		Source:      "/drive/Home/out.zip",
		Destination: "/drive/Home/unpacked/",
		Format:      "zip",
		Node:        "node-a",
	}, "")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if taskID != "task-456" {
		t.Errorf("taskID: got %q", taskID)
	}
}

// ============================================================
// entries.go
// ============================================================

func TestStreamEntries_Happy(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/archive/node-a/entries" {
			t.Errorf("path: %q", r.URL.Path)
		}
		if r.URL.Query().Get("source") != "/drive/Home/out.zip" {
			t.Errorf("source query: %q", r.URL.Query().Get("source"))
		}
		// Format is intentionally NOT sent on the wire — the spec
		// lists only `source` as a query parameter for entries.
		// The local Format hint on EntriesOptions exists for
		// cobra-layer password-compat checks.
		if got := r.URL.Query().Get("format"); got != "" {
			t.Errorf("format must not be on the wire, got %q", got)
		}
		w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		_, _ = w.Write([]byte(`{"path":"a.txt","size":10,"modified":1716800000,"is_dir":false,"encrypted":false}` + "\n"))
		_, _ = w.Write([]byte(`{"path":"sub/","size":0,"modified":1716800000,"is_dir":true,"encrypted":false}` + "\n"))
		_, _ = w.Write([]byte(`{"path":"sub/b.txt","size":42,"modified":1716800000,"is_dir":false,"encrypted":true}` + "\n"))
		_, _ = w.Write([]byte(`{"_done":true,"total":3}` + "\n"))
	}))

	var got []Entry
	total, err := cli.StreamEntries(context.Background(), EntriesOptions{
		Source: "/drive/Home/out.zip",
		Format: "zip",
		Node:   "node-a",
	}, "", func(e Entry) error {
		got = append(got, e)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamEntries: %v", err)
	}
	if total != 3 {
		t.Errorf("total: got %d want 3", total)
	}
	if len(got) != 3 {
		t.Fatalf("entries: got %d want 3", len(got))
	}
	if !got[1].IsDir {
		t.Error("sub/ should be IsDir=true")
	}
	if !got[2].Encrypted {
		t.Error("sub/b.txt should be Encrypted=true")
	}
}

func TestStreamEntries_PasswordRequired(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(`{"_error":"need password","code":"password_required"}` + "\n"))
	}))

	count := 0
	_, err := cli.StreamEntries(context.Background(), EntriesOptions{
		Source: "/drive/Home/secret.zip",
		Format: "zip",
		Node:   "node-a",
	}, "", func(e Entry) error {
		count++
		return nil
	})
	if count != 0 {
		t.Errorf("cb invoked %d times; expected 0", count)
	}
	se, ok := IsEntriesStreamError(err)
	if !ok {
		t.Fatalf("want *EntriesStreamError, got %v", err)
	}
	if se.Code != CodePasswordRequired {
		t.Errorf("code: got %q want %q", se.Code, CodePasswordRequired)
	}
}

func TestStreamEntries_CallbackAbort(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		for i := 0; i < 100; i++ {
			line := fmt.Sprintf(`{"path":"f%d.txt","size":1,"modified":0,"is_dir":false,"encrypted":false}`+"\n", i)
			if _, err := w.Write([]byte(line)); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))

	sentinel := errors.New("user abort")
	got := 0
	_, err := cli.StreamEntries(context.Background(), EntriesOptions{
		Source: "/drive/Home/big.zip",
		Format: "zip",
		Node:   "node-a",
	}, "", func(e Entry) error {
		got++
		if got >= 5 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("want sentinel propagated, got %v", err)
	}
	if got < 5 || got > 10 {
		t.Errorf("cb invoked %d times; expected ~5", got)
	}
}

func TestStreamEntries_PreStreamHTTPError(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad format"}`))
	}))
	_, err := cli.StreamEntries(context.Background(), EntriesOptions{
		Source: "/drive/Home/out.zip",
		Format: "zip",
		Node:   "node-a",
	}, "", func(Entry) error { return nil })
	var h *HTTPError
	if !errors.As(err, &h) {
		t.Fatalf("want *HTTPError, got %v", err)
	}
	if h.Status != http.StatusBadRequest {
		t.Errorf("status: %d", h.Status)
	}
}

func TestStreamEntries_MissingDoneSentinel(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(`{"path":"a.txt","size":1,"modified":0,"is_dir":false,"encrypted":false}` + "\n"))
		// No _done line.
	}))
	count := 0
	_, err := cli.StreamEntries(context.Background(), EntriesOptions{
		Source: "/drive/Home/out.zip",
		Format: "zip",
		Node:   "node-a",
	}, "", func(Entry) error { count++; return nil })
	if err == nil {
		t.Fatal("want error for missing _done sentinel")
	}
	if !strings.Contains(err.Error(), "_done") {
		t.Errorf("err message should mention `_done`; got %v", err)
	}
}

// ============================================================
// entry.go
// ============================================================

func TestStreamEntry_Happy(t *testing.T) {
	payload := []byte("hello world")
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/archive/node-a/entry" {
			t.Errorf("path: %q", r.URL.Path)
		}
		if r.URL.Query().Get("source") != "/drive/Home/out.zip" {
			t.Errorf("source query: %q", r.URL.Query().Get("source"))
		}
		if r.URL.Query().Get("path") != "dir/file.txt" {
			t.Errorf("path query: %q", r.URL.Query().Get("path"))
		}
		if got := r.URL.Query().Get("format"); got != "" {
			t.Errorf("format must not be on the wire, got %q", got)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="file.txt"`)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		_, _ = w.Write(payload)
	}))
	var buf bytes.Buffer
	dl, err := cli.StreamEntry(context.Background(), EntryOptions{
		Source: "/drive/Home/out.zip",
		Path:   "dir/file.txt",
		Format: "zip",
		Node:   "node-a",
	}, "", &buf)
	if err != nil {
		t.Fatalf("StreamEntry: %v", err)
	}
	if buf.String() != "hello world" {
		t.Errorf("body: got %q", buf.String())
	}
	if dl.BytesWritten != int64(len(payload)) {
		t.Errorf("BytesWritten: %d", dl.BytesWritten)
	}
	if dl.Filename != "file.txt" {
		t.Errorf("Filename: %q", dl.Filename)
	}
}

func TestStreamEntry_NotFound(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"no such entry","code":"not_found"}`))
	}))
	var buf bytes.Buffer
	_, err := cli.StreamEntry(context.Background(), EntryOptions{
		Source: "/drive/Home/out.zip",
		Path:   "missing.txt",
		Format: "zip",
		Node:   "node-a",
	}, "", &buf)
	var entryErr *EntryError
	if !errors.As(err, &entryErr) {
		t.Fatalf("want *EntryError, got %v", err)
	}
	if entryErr.Code != CodeNotFound {
		t.Errorf("code: %q", entryErr.Code)
	}
	// Ensure errors.As(*HTTPError) still fires (Unwrap chain).
	var h *HTTPError
	if !errors.As(err, &h) {
		t.Error("EntryError should unwrap to *HTTPError")
	}
	if h.Status != http.StatusNotFound {
		t.Errorf("status: %d", h.Status)
	}
}

// TestStreamEntry_TooLarge exercises the spec §4 413 path: when
// a single in-archive entry exceeds the server's single-shot
// read limit, the server replies with the typed JSON body and
// HTTP 413. We assert the code constant the cobra-layer
// formatter switches on is set, and that the status surfaces
// via errors.As(*HTTPError).
func TestStreamEntry_TooLarge(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte(`{"error":"entry exceeds 2 GiB","code":"entry_too_large"}`))
	}))
	var buf bytes.Buffer
	_, err := cli.StreamEntry(context.Background(), EntryOptions{
		Source: "/drive/Home/out.zip",
		Path:   "huge.bin",
		Format: "zip",
		Node:   "node-a",
	}, "", &buf)
	var entryErr *EntryError
	if !errors.As(err, &entryErr) {
		t.Fatalf("want *EntryError, got %v", err)
	}
	if entryErr.Code != CodeEntryTooLarge {
		t.Errorf("code: got %q want %q", entryErr.Code, CodeEntryTooLarge)
	}
	var h *HTTPError
	if !errors.As(err, &h) {
		t.Fatal("EntryError should unwrap to *HTTPError")
	}
	if h.Status != http.StatusRequestEntityTooLarge {
		t.Errorf("status: got %d want 413", h.Status)
	}
}

func TestStreamEntry_BadHTMLBodyFallsBack(t *testing.T) {
	// Misconfigured server / gateway returns HTML; we still want
	// a usable HTTPError (no panic in JSON decode).
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`<html><body>502 Bad Gateway</body></html>`))
	}))
	var buf bytes.Buffer
	_, err := cli.StreamEntry(context.Background(), EntryOptions{
		Source: "/drive/Home/out.zip",
		Path:   "x",
		Format: "zip",
		Node:   "node-a",
	}, "", &buf)
	var h *HTTPError
	if !errors.As(err, &h) {
		t.Fatalf("want *HTTPError, got %v", err)
	}
	if h.Status != http.StatusBadGateway {
		t.Errorf("status: %d", h.Status)
	}
}

func TestParseContentDispositionFilename(t *testing.T) {
	cases := []struct{ in, want string }{
		{`attachment; filename="foo.txt"`, "foo.txt"},
		{`attachment; filename=foo.txt`, "foo.txt"},
		{`attachment; filename=foo.txt; other=x`, "foo.txt"},
		{`inline`, ""},
		{``, ""},
	}
	for _, c := range cases {
		if got := parseContentDispositionFilename(c.in); got != c.want {
			t.Errorf("parseContentDispositionFilename(%q): got %q want %q", c.in, got, c.want)
		}
	}
}

// ============================================================
// task.go
// ============================================================

func TestWaitTask_Completed(t *testing.T) {
	polls := 0
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/task/node-a/" {
			t.Errorf("path: %q", r.URL.Path)
		}
		if r.URL.Query().Get("task_id") != "task-123" {
			t.Errorf("task_id: %q", r.URL.Query().Get("task_id"))
		}
		polls++
		// First two polls report "running" with progress; third
		// reports "completed".
		switch polls {
		case 1:
			_, _ = w.Write([]byte(`{"task":{"id":"task-123","status":"running","progress":25}}`))
		case 2:
			_, _ = w.Write([]byte(`{"task":{"id":"task-123","status":"running","progress":75}}`))
		default:
			_, _ = w.Write([]byte(`{"task":{"id":"task-123","status":"completed","progress":100}}`))
		}
	}))

	updates := 0
	err := cli.WaitTask(context.Background(), "node-a", "task-123", 10*time.Millisecond, func(u TaskUpdate) {
		updates++
		if u.Status != "running" {
			t.Errorf("update status: %q", u.Status)
		}
	})
	if err != nil {
		t.Fatalf("WaitTask: %v", err)
	}
	if polls < 3 {
		t.Errorf("polls: got %d want >= 3", polls)
	}
	if updates < 2 {
		t.Errorf("updates: got %d want >= 2", updates)
	}
}

func TestWaitTask_Failed(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"task":{"id":"t","status":"failed","failed_reason":"disk full"}}`))
	}))
	err := cli.WaitTask(context.Background(), "node-a", "t", 1*time.Millisecond, nil)
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Errorf("want failure with reason, got %v", err)
	}
}

func TestWaitTask_Cancelled(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"task":{"id":"t","status":"cancelled"}}`))
	}))
	err := cli.WaitTask(context.Background(), "node-a", "t", 1*time.Millisecond, nil)
	if err == nil || !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("want cancelled error, got %v", err)
	}
}

func TestWaitTask_CtxCancel(t *testing.T) {
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"task":{"id":"t","status":"running","progress":0}}`))
	}))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := cli.WaitTask(ctx, "node-a", "t", 5*time.Millisecond, nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

func TestCancelTask(t *testing.T) {
	hit := false
	cli, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method: %s", r.Method)
		}
		if r.URL.Query().Get("task_id") != "task-123" {
			t.Errorf("task_id: %q", r.URL.Query().Get("task_id"))
		}
		hit = true
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok"}`))
	}))
	if err := cli.CancelTask(context.Background(), "node-a", "task-123"); err != nil {
		t.Fatalf("CancelTask: %v", err)
	}
	if !hit {
		t.Error("server not hit")
	}
}
