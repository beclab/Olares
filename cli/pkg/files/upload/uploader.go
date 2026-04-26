// uploader.go: single-file chunked uploader. Drives the resumable upload
// protocol the LarePass web app uses (Resumable.js + Drive v2 endpoints):
//
//  1. probe the server for already-uploaded bytes via
//     /upload/file-uploaded-bytes/<node>/  (GetUploadedBytes)
//  2. align to a chunk boundary by flooring (matches the web app's
//     `Math.floor(uploadedBytes / chunkSize)` — re-uploading the
//     "overflow" within that chunk is harmless and identical-byte)
//  3. ask the server for an upload link via
//     /upload/upload-link/<node>/  (GetUploadLink, once per file)
//  4. POST each remaining chunk as multipart/form-data with the
//     Resumable.js parameter shape (resumableChunkNumber, ..., file=chunk)
//     plus the Drive-specific extras (parent_dir, driveType, ...)
//  5. classify each chunk response:
//     - 200 / 201        → chunk accepted, advance
//     - permanent codes  → fail fast (see permanentStatuses)
//     - everything else  → retry up to opts.MaxRetries with backoff
//
// Empty files are routed through CreateEmptyFile (the web app does the
// same — Resumable.js can't represent a 0-byte chunk).
package upload

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/files/encodepath"
)

// DefaultChunkSize is 8 MiB — the same value the web app uses
// (apps/packages/app/src/api/files/v2/drive/data.ts L55: SIZE = 8MB).
// Stick to it unless you have a very good reason; the server's
// already-uploaded-bytes accounting is keyed on the chunk size the
// previous run used, so changing this mid-stream means the resume
// boundary computation diverges (the floor() trick still gets you to a
// safe re-upload offset, but you waste bandwidth re-sending bytes the
// server already had).
const DefaultChunkSize = 8 * 1024 * 1024

// DefaultMaxRetries: per-chunk retry budget. Matches the web app's
// `maxChunkRetries: 3` (resumejs.ts init() L166).
const DefaultMaxRetries = 3

// DefaultRetryBackoff: between failed-chunk retries. Matches the web
// app's chunkRetryInterval default of 5s.
const DefaultRetryBackoff = 5 * time.Second

// UploadOpts is everything UploadFile needs to push one local file into
// Drive/Home. It's a value type so callers (the cobra command, the
// directory walker) can build one per file and tweak fields per call
// without sharing mutable state.
type UploadOpts struct {
	// LocalPath is the absolute or working-directory-relative path to
	// the file on disk that we're uploading.
	LocalPath string

	// Node is the {node} path segment for /upload/upload-link/<node>/
	// and /upload/file-uploaded-bytes/<node>/. Resolved by the cobra
	// command up-front via Client.FetchNodes.
	Node string

	// ParentDir is the destination directory on the server WITH the
	// `/drive/Home/...` prefix and a TRAILING `/`, e.g.
	// `/drive/Home/Documents/`. This is the value passed as the
	// `file_path` query for upload-link, the `parent_dir` query for
	// file-uploaded-bytes, AND the `parent_dir` form field on each
	// chunk POST. They MUST agree byte-for-byte for resume to find the
	// existing partial upload — that's why we plumb a single value
	// rather than recomputing it at each call site.
	ParentDir string

	// RemoteName is the bare filename on the server (no directory
	// components). For directory uploads, this is the leaf file name —
	// the directory components live in RelativePath.
	RemoteName string

	// RelativePath is the file's path relative to the upload root, in
	// POSIX form (forward slashes). For a single-file upload this is
	// just RemoteName; for a directory upload it includes the in-tree
	// directory components, e.g. `mydir/photos/IMG_001.jpg`. The web
	// app uses this for resumableRelativePath + the per-chunk
	// `relative_path` form field; the server uses both for sub-directory
	// auto-creation under parent_dir.
	RelativePath string

	// ChunkSize: bytes per chunk. Defaults to DefaultChunkSize when
	// zero.
	ChunkSize int64
	// MaxRetries: retries per chunk on transient failures. Defaults to
	// DefaultMaxRetries when zero. Negative disables retries.
	MaxRetries int
	// RetryBackoff: wait between retries. Defaults to
	// DefaultRetryBackoff when zero.
	RetryBackoff time.Duration
}

// ProgressFunc is the per-chunk callback the cobra command uses to
// surface a one-line text progress indicator. `uploaded` is the total
// bytes pushed (cumulative) and `total` is the file size; either may be
// reported as `(0, 0)` to indicate "an empty file just completed".
type ProgressFunc func(uploaded, total int64)

// permanentStatuses lists HTTP status codes that the web app's
// Resumable.js treats as non-retryable (resumable.js: see the
// `permanentErrors` option, which the web app passes as the array
// below). 5xx is partial: 500/501 are fatal, 502/503/504 are not.
var permanentStatuses = map[int]struct{}{
	400: {}, 401: {}, 403: {}, 404: {}, 409: {}, 415: {},
	440: {}, 441: {}, 442: {}, 443: {},
	500: {}, 501: {},
}

// UploadFile uploads `opts.LocalPath` to `opts.ParentDir` + `opts.RemoteName`,
// resuming from whatever the server already has. Empty files are routed
// to CreateEmptyFile (the chunk pipeline can't express 0-byte chunks).
//
// `progress`, if non-nil, is invoked once after the resume probe (with
// uploaded=<server bytes>, total=<file size>) and once per accepted
// chunk thereafter. It is NOT invoked per retry attempt.
func (c *Client) UploadFile(ctx context.Context, opts UploadOpts, progress ProgressFunc) error {
	if err := opts.normalize(); err != nil {
		return err
	}

	st, err := os.Stat(opts.LocalPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", opts.LocalPath, err)
	}
	if st.IsDir() {
		return fmt.Errorf("UploadFile: %s is a directory; use the walker", opts.LocalPath)
	}
	fileSize := st.Size()

	// Empty file: web app sends a separate POST to /api/resources/...
	// (uploadEmptyFile) instead of routing through Resumable.js. Mirror
	// that here so 0-byte files actually materialize on the server
	// (Resumable.js cannot generate a chunk of length 0).
	if fileSize == 0 {
		// RemotePath inside Drive/Home for the create call: strip the
		// `/drive/Home/` prefix the API client expects.
		rel := joinRemote(opts.ParentDir, opts.RemoteName)
		if err := c.CreateEmptyFile(ctx, rel); err != nil {
			return err
		}
		if progress != nil {
			progress(0, 0)
		}
		return nil
	}

	// Resume probe: how many bytes does the server already have for
	// this exact (parent_dir, file_name) pair? Web app does
	// `Math.floor(uploadedBytes / chunkSize)` to find the next chunk
	// index — anything smaller than that is "behind us", anything
	// equal-or-larger means "the chunk this byte falls into still needs
	// to be (re)sent in full". We follow the same floor convention so
	// the wire-level resume boundary matches.
	uploadedBytes, err := c.GetUploadedBytes(ctx, opts.Node, opts.ParentDir, opts.RemoteName)
	if err != nil {
		// GetUploadedBytes already swallows errors as "0", but defend
		// in depth in case that policy changes.
		uploadedBytes = 0
	}
	if uploadedBytes > fileSize {
		// Server claims more bytes than the local file has. Could be a
		// truncation on the source, or a mismatched name. Restart from
		// 0 — sending fewer bytes than promised would surface a
		// confusing "file size mismatch" error from the chunk endpoint.
		uploadedBytes = 0
	}
	startChunk := uploadedBytes / opts.ChunkSize
	totalChunks := (fileSize + opts.ChunkSize - 1) / opts.ChunkSize
	if startChunk > totalChunks {
		startChunk = totalChunks
	}

	if progress != nil {
		progress(startChunk*opts.ChunkSize, fileSize)
	}

	// File already complete on the server? Floor-aligned offset equal
	// to total chunks means there's nothing left to push — but we still
	// re-issue the last chunk to give the server a chance to finalize
	// (Resumable.js's behavior when its file-uploaded-bytes probe
	// returns the full file size is to skip the upload entirely; we
	// match that here).
	if startChunk >= totalChunks {
		if progress != nil {
			progress(fileSize, fileSize)
		}
		return nil
	}

	// Open the file once and seek per chunk. Saves opening N file
	// handles for big files; on POSIX, ReadAt is cheaper than Read+Seek
	// because it doesn't mutate the file cursor.
	f, err := os.Open(opts.LocalPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", opts.LocalPath, err)
	}
	defer f.Close()

	uploadLink, err := c.GetUploadLink(ctx, opts.Node, opts.ParentDir)
	if err != nil {
		return err
	}
	chunkURL := c.BaseURL + uploadLink

	identifier := uploadIdentifier(opts.ParentDir, opts.RelativePath)
	mimeType := guessMIME(opts.LocalPath)

	buf := make([]byte, opts.ChunkSize)
	for chunkIdx := startChunk; chunkIdx < totalChunks; chunkIdx++ {
		startByte := chunkIdx * opts.ChunkSize
		chunkLen := opts.ChunkSize
		if remaining := fileSize - startByte; remaining < chunkLen {
			chunkLen = remaining
		}
		// ReadAt may return io.EOF on the final short read along with
		// the actual byte count — that's fine, treat (n>0, EOF) as a
		// successful read of n bytes. Anything else (real errors, or
		// short reads that aren't at EOF) is a hard failure.
		n, rerr := f.ReadAt(buf[:chunkLen], startByte)
		if rerr != nil && !(rerr == io.EOF && int64(n) == chunkLen) {
			return fmt.Errorf("read %s @ %d: %w", opts.LocalPath, startByte, rerr)
		}
		if int64(n) != chunkLen {
			return fmt.Errorf("short read at chunk %d: got %d, want %d", chunkIdx+1, n, chunkLen)
		}
		chunkData := buf[:chunkLen]

		if err := c.uploadChunk(ctx, chunkURL, opts, chunkUploadCtx{
			ChunkIndex:    chunkIdx, // 0-based; we send +1 on the wire
			TotalChunks:   totalChunks,
			ChunkLen:      chunkLen,
			StartByte:     startByte,
			FileSize:      fileSize,
			Identifier:    identifier,
			MimeType:      mimeType,
			ChunkContents: chunkData,
		}); err != nil {
			return fmt.Errorf("upload chunk %d/%d of %s: %w",
				chunkIdx+1, totalChunks, opts.LocalPath, err)
		}
		if progress != nil {
			progress(startByte+chunkLen, fileSize)
		}
	}
	return nil
}

// chunkUploadCtx bundles the per-chunk state. Pulled out into a struct
// just so uploadChunk's signature doesn't grow to 10+ parameters.
type chunkUploadCtx struct {
	ChunkIndex    int64
	TotalChunks   int64
	ChunkLen      int64
	StartByte     int64
	FileSize      int64
	Identifier    string
	MimeType      string
	ChunkContents []byte
}

// uploadChunk POSTs a single chunk and applies the
// permanent / retryable / success classification. It returns nil only
// on a 2xx response (the web app accepts both 200 and 201; we follow
// suit). Permanent errors short-circuit the retry loop.
func (c *Client) uploadChunk(
	ctx context.Context,
	chunkURL string,
	opts UploadOpts,
	cu chunkUploadCtx,
) error {
	maxAttempts := opts.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		body, contentType, berr := buildChunkBody(opts, cu)
		if berr != nil {
			return berr
		}
		headers := http.Header{
			"Accept": []string{"application/json; text/javascript, */*; q=0.01"},
			"Content-Disposition": []string{
				`attachment; filename="` + encodepath.EncodeURIComponent(opts.RemoteName) + `"`,
			},
			"Content-Range": []string{fmt.Sprintf(
				"bytes %d-%d/%d",
				cu.StartByte,
				cu.StartByte+cu.ChunkLen-1,
				cu.FileSize,
			)},
		}

		_, err := c.do(ctx, http.MethodPost, chunkURL, body, headers, contentType)
		if err == nil {
			return nil
		}

		// Don't burn retries on cancellation: ctx.Err is final.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		var hErr *HTTPError
		if errors.As(err, &hErr) {
			if _, isPermanent := permanentStatuses[hErr.Status]; isPermanent {
				return err
			}
		}
		lastErr = err
		if attempt < maxAttempts {
			// Sleep but stay cancelable.
			select {
			case <-time.After(opts.RetryBackoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return fmt.Errorf("exhausted %d attempts: %w", maxAttempts, lastErr)
}

// buildChunkBody assembles the multipart/form-data body for one chunk.
// Field set + ordering matches what Resumable.js + LarePass's
// resumejs.ts setQuery() emits, so the server can't tell whether this
// came from a browser or olares-cli.
//
// Three groups of fields:
//
//   - Resumable.js core (resumableChunkNumber, ..., resumableTotalChunks):
//     match the parameter names from resumable.js's `chunkNumberParameterName`
//     etc. defaults. resumableChunkNumber is 1-indexed (offset+1).
//   - Drive customQuery (parent_dir, driveType, ..., resumableType):
//     the Drive override-set from setQuery() in resumejs.ts.
//   - file: the actual chunk bytes, sent under the multipart filename
//     of the basename (matches Resumable.js `fileParameterName: 'file'`).
//
// `relative_path` is included when RelativePath has a directory
// component (i.e. this is a folder-walk upload), matching
// resumejs.ts onChunkingComplete's `relative_path:
// relativePath.slice(0, lastIndexOf('/')+1)` semantics.
func buildChunkBody(opts UploadOpts, cu chunkUploadCtx) (io.Reader, string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	add := func(name, value string) error {
		return mw.WriteField(name, value)
	}

	// --- Resumable.js core fields (defaults from resumable.js).
	for _, kv := range []struct{ k, v string }{
		{"resumableChunkNumber", strconv.FormatInt(cu.ChunkIndex+1, 10)},
		{"resumableChunkSize", strconv.FormatInt(opts.ChunkSize, 10)},
		{"resumableCurrentChunkSize", strconv.FormatInt(cu.ChunkLen, 10)},
		{"resumableTotalSize", strconv.FormatInt(cu.FileSize, 10)},
		{"resumableType", cu.MimeType},
		{"resumableIdentifier", cu.Identifier},
		{"resumableFilename", opts.RemoteName},
		{"resumableRelativePath", opts.RelativePath},
		{"resumableTotalChunks", strconv.FormatInt(cu.TotalChunks, 10)},
	} {
		if err := add(kv.k, kv.v); err != nil {
			return nil, "", err
		}
	}

	// --- Drive customQuery (resumejs.ts setQuery + onChunkingComplete).
	if err := add("parent_dir", opts.ParentDir); err != nil {
		return nil, "", err
	}
	if err := add("driveType", "Drive"); err != nil {
		return nil, "", err
	}
	if dir := relativeDir(opts.RelativePath); dir != "" {
		if err := add("relative_path", dir); err != nil {
			return nil, "", err
		}
	}

	// --- file part. Use CreatePart instead of CreateFormFile so we can
	// set the chunk's MIME type explicitly (CreateFormFile hardcodes
	// "application/octet-stream"). This matches the web app's
	// `setChunkTypeFromFile` path, where the chunk's blob carries the
	// file's real MIME (resumable.js calls `file[func](start, end,
	// fileType)`).
	hdr := textproto.MIMEHeader{}
	hdr.Set(
		"Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename=%q`, opts.RemoteName),
	)
	hdr.Set("Content-Type", cu.MimeType)
	part, err := mw.CreatePart(hdr)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(cu.ChunkContents); err != nil {
		return nil, "", err
	}
	if err := mw.Close(); err != nil {
		return nil, "", err
	}
	return &buf, mw.FormDataContentType(), nil
}

func (o *UploadOpts) normalize() error {
	if o.LocalPath == "" {
		return errors.New("UploadOpts.LocalPath is required")
	}
	if o.Node == "" {
		return errors.New("UploadOpts.Node is required")
	}
	if o.ParentDir == "" {
		return errors.New("UploadOpts.ParentDir is required")
	}
	if !strings.HasSuffix(o.ParentDir, "/") {
		// The server-side resume probe + chunk endpoint both expect
		// parent_dir to end in '/'. Force it rather than failing — the
		// caller almost always means "this directory".
		o.ParentDir += "/"
	}
	if o.RemoteName == "" {
		o.RemoteName = filepath.Base(o.LocalPath)
	}
	if o.RelativePath == "" {
		o.RelativePath = o.RemoteName
	}
	if o.ChunkSize <= 0 {
		o.ChunkSize = DefaultChunkSize
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = DefaultMaxRetries
	}
	if o.RetryBackoff == 0 {
		o.RetryBackoff = DefaultRetryBackoff
	}
	return nil
}

// joinRemote builds the Drive/Home-relative path used by CreateEmptyFile.
// `parentDir` is the full `/drive/Home/...` form; we strip the prefix +
// trailing slash so CreateEmptyFile can rebuild the URL with the right
// percent-encoding.
func joinRemote(parentDir, name string) string {
	pd := strings.TrimSuffix(parentDir, "/")
	const prefix = "/drive/Home"
	if strings.HasPrefix(pd, prefix) {
		pd = strings.TrimPrefix(pd, prefix)
	}
	pd = strings.Trim(pd, "/")
	if pd == "" {
		return name
	}
	return pd + "/" + name
}

// relativeDir returns the directory portion of a POSIX-style relative
// path with a trailing slash, or "" if the path has no directory
// component (single-file upload). Mirrors the web app's
// `relativePath.slice(0, lastIndexOf('/') + 1)` from
// resumejs.ts onChunkingComplete.
func relativeDir(relPath string) string {
	idx := strings.LastIndex(relPath, "/")
	if idx < 0 {
		return ""
	}
	return relPath[:idx+1]
}

// uploadIdentifier picks a stable per-(parent_dir, relativePath)
// identifier so retries / re-runs of the same file produce the same
// `resumableIdentifier` form value. The server-side resume key is
// (parent_dir, file_name), so this is effectively cosmetic — but
// keeping it stable across runs makes server logs easier to follow,
// and makes it impossible for two concurrent uploads of different
// files to collide on the identifier.
func uploadIdentifier(parentDir, relativePath string) string {
	sum := md5.Sum([]byte(parentDir + relativePath)) // #nosec G401 -- not security
	return hex.EncodeToString(sum[:])
}

// guessMIME returns a best-effort MIME type for the file, mirroring
// the web app's `mime.getType(fileName) || 'application/octet-stream'`
// fallback. We only sniff the extension (no magic bytes) — same as the
// web app, which only sees the filename + browser-detected blob type.
func guessMIME(localPath string) string {
	switch strings.ToLower(filepath.Ext(localPath)) {
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	}
	return "application/octet-stream"
}
