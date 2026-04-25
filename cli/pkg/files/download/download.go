// download.go: single-file download with optional resume + retry.
//
// Wire-level behavior:
//
//   - GET /api/raw/<encPath>; the file-server's raw_service.go sets
//     `Accept-Ranges: bytes` and parseRangeHeader implements
//     `Range: bytes=N-` / `bytes=N-M` / `bytes=-N`. So when the user
//     passes --resume, we send `Range: bytes=<localSize>-`, expect a
//     206 Partial Content, and append to the existing local file.
//   - 200 OK in response to a Range request means the server ignored
//     the header (most often because the resource isn't a real file —
//     a redirect / cloud-backed handler — or the file changed under
//     us). We fall back to a full overwrite via the same tmp+rename
//     dance --overwrite uses, so the local file is left consistent.
//   - 416 Requested Range Not Satisfiable typically means localSize ==
//     remoteSize (we already have the whole file). We treat that as
//     success.
//
// Failure handling:
//
//   - 4xx (other than 416 above) is a permanent error — no retries.
//   - 5xx and transport errors retry with exponential backoff up to
//     opts.MaxRetries times. Same retry classification spirit as the
//     upload package's chunk POST loop.
//
// Atomicity:
//
//   - Full and overwrite paths write to `dst.tmp` and rename on
//     success. So a crash mid-download leaves the previous version of
//     dst intact (or no file at all if it was a fresh download).
//   - Resume writes directly to dst with O_APPEND. A crash mid-resume
//     leaves a partial file that the next --resume run will pick up
//     from — that's the whole point of the flag.
package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DefaultMaxRetries is the per-file retry budget on transient errors.
// Matches the upload pipeline's default chunk retries.
const DefaultMaxRetries = 3

// DefaultRetryBackoff: base wait between retries; the loop doubles up
// to a cap of 4s so a transient blip doesn't burn the whole budget on
// the first failure.
const DefaultRetryBackoff = 250 * time.Millisecond

// Options controls a single DownloadFile call. Zero-valued fields fall
// back to package defaults (see normalize()).
type Options struct {
	// Overwrite: if dst exists, replace it. Without this AND without
	// Resume, DownloadFile errors out so the user has to opt in
	// explicitly to clobber local data.
	Overwrite bool
	// Resume: if dst exists, ask the server to start at the local
	// file's current size via `Range: bytes=N-`. Implies "this is the
	// continuation of a previous attempt": the local tail bytes are
	// trusted as-is. Falls back to a full overwrite if the server
	// returns 200 (Range ignored) — see the file header for why.
	Resume bool
	// MaxRetries: transient error retries (0 means use the default;
	// negative disables retries entirely).
	MaxRetries int
	// RetryBackoff: base backoff; the loop doubles each attempt up to
	// 4s. 0 means use the default.
	RetryBackoff time.Duration
}

// ProgressFunc is the per-write progress callback. `written` is the
// total bytes flushed to disk so far for this file (cumulative,
// including any resumed prefix); `total` is the file's size as
// reported by the server's Content-Length / Content-Range, or -1 if
// the server didn't tell us. Called periodically (not per-byte) — the
// downloader throttles to ~64 KiB granularity so progress updates
// don't dominate CPU on fast loopback transfers.
type ProgressFunc func(written, total int64)

// DownloadFile fetches `plainPath` (a `<fileType>/<extend>/<subPath>`
// triple, un-encoded — the client encodes with EncodeURL internally)
// into the local file at `dst`.
//
// Returns the number of bytes WRITTEN to dst by this call (so a
// resumed download reports just the appended bytes — that matches
// the per-call "did work" semantics callers want for status lines).
//
// `progress` may be nil. When non-nil it's invoked with the cumulative
// bytes-written-this-call AND the total file size if known.
func (c *Client) DownloadFile(
	ctx context.Context,
	plainPath, dst string,
	opts Options,
	progress ProgressFunc,
) (int64, error) {
	opts.normalize()

	// Decide the strategy first — it dictates which path we open and
	// what Range header (if any) we send.
	mode, localSize, err := planLocalWrite(dst, opts)
	if err != nil {
		return 0, err
	}

	maxAttempts := opts.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	backoff := opts.RetryBackoff
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		written, err := c.attemptDownload(ctx, plainPath, dst, mode, localSize, progress)
		if err == nil {
			return written, nil
		}

		// Cancellation is always final.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return written, ctxErr
		}

		// 4xx (other than 416, handled inside attemptDownload as
		// "already complete") is permanent: no point retrying.
		var hErr *HTTPError
		if errors.As(err, &hErr) && hErr.Status >= 400 && hErr.Status < 500 {
			return written, err
		}

		lastErr = err
		if attempt < maxAttempts {
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return written, ctx.Err()
			}
			// Exponential backoff capped at 4s.
			backoff *= 2
			if backoff > 4*time.Second {
				backoff = 4 * time.Second
			}
		}
	}
	return 0, fmt.Errorf("download %s: exhausted %d attempts: %w", plainPath, maxAttempts, lastErr)
}

// writeMode encodes how attemptDownload should open the destination
// file. Pulled out as a typed value rather than a bag of bools so the
// branching inside attemptDownload reads naturally.
type writeMode int

const (
	// writeFresh: dst doesn't exist (or we don't care about its
	// previous contents). Write to dst.tmp + rename.
	writeFresh writeMode = iota
	// writeOverwrite: dst exists and Overwrite is set. Write to
	// dst.tmp + rename so the previous version stays intact until the
	// new one is fully on disk.
	writeOverwrite
	// writeResume: dst exists and Resume is set. Open dst with
	// O_APPEND and ask the server for `Range: bytes=<localSize>-`.
	writeResume
)

// planLocalWrite picks the writeMode + initial local size based on
// what's currently at `dst` and which flags the caller passed. It's
// the only thing in this file that touches `os.Stat`, so the test
// matrix lives in one place.
func planLocalWrite(dst string, opts Options) (writeMode, int64, error) {
	st, err := os.Stat(dst)
	switch {
	case err == nil && st.IsDir():
		return 0, 0, fmt.Errorf("local destination %q is an existing directory", dst)
	case err == nil:
		switch {
		case opts.Resume:
			return writeResume, st.Size(), nil
		case opts.Overwrite:
			return writeOverwrite, 0, nil
		default:
			return 0, 0, fmt.Errorf(
				"local file %q already exists; pass --overwrite to replace it or --resume to continue a previous download",
				dst,
			)
		}
	case errors.Is(err, os.ErrNotExist):
		return writeFresh, 0, nil
	default:
		return 0, 0, fmt.Errorf("stat %s: %w", dst, err)
	}
}

// attemptDownload runs one HTTP request + body copy. It is called
// from DownloadFile inside a retry loop, so it must:
//   - leave dst in a valid state on failure (tmp file is cleaned up;
//     resume mode never closes the real file in a half-flushed state);
//   - return enough information for the retry classifier (status code
//     embedded in *HTTPError for 4xx/5xx, raw transport errors for the
//     network layer).
func (c *Client) attemptDownload(
	ctx context.Context,
	plainPath, dst string,
	mode writeMode,
	localSize int64,
	progress ProgressFunc,
) (int64, error) {
	endpoint := c.rawURL(plainPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	if c.AccessToken != "" {
		req.Header.Set("X-Authorization", c.AccessToken)
	}
	if mode == writeResume && localSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", localSize))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Two cases collapse here:
		//   1. fresh / overwrite: expected.
		//   2. resume: the server ignored Range. Fall back to full
		//      overwrite — re-download the whole file via tmp+rename
		//      so we never leave dst in a torn state.
		return writeFullToDst(dst, resp, progress)
	case http.StatusPartialContent:
		if mode != writeResume {
			// We didn't ask for a range but the server sent one
			// anyway. Don't second-guess — append the partial body
			// to dst.tmp, but treat as fresh write so the on-disk
			// state stays well-defined.
			return writeFullToDst(dst, resp, progress)
		}
		return appendToDst(dst, localSize, resp, progress)
	case http.StatusRequestedRangeNotSatisfiable:
		// Almost always means localSize == remoteSize: the file is
		// already complete. Surface that as success; the user gets a
		// "0 new bytes" line in the cobra cmd's progress output.
		if mode == writeResume {
			return 0, nil
		}
		fallthrough
	default:
		body, _ := io.ReadAll(resp.Body)
		return 0, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodGet,
		}
	}
}

// writeFullToDst streams resp.Body into `dst.tmp` and renames it onto
// `dst` on success. Used by the fresh / overwrite paths AND by the
// resume-fell-back-to-200 case, so an in-flight failure can never
// corrupt a previously-good local file.
func writeFullToDst(dst string, resp *http.Response, progress ProgressFunc) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, fmt.Errorf("mkdir parent of %s: %w", dst, err)
	}
	tmp := dst + ".tmp"
	// O_TRUNC so a stale tmp from a previous failed attempt doesn't
	// concatenate with the new body. 0o644 matches `cp` defaults; the
	// user's umask still applies via the OS layer.
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", tmp, err)
	}
	total := contentLengthOrTotal(resp)
	written, copyErr := copyWithProgress(f, resp.Body, total, progress)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return written, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("close %s: %w", tmp, closeErr)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("rename %s -> %s: %w", tmp, dst, err)
	}
	return written, nil
}

// appendToDst is the resume path: open dst with O_APPEND and stream
// the partial-content body straight onto the end. We don't use a tmp
// file here because the server has already promised us "exactly the
// bytes from offset N onwards" and we want a crash mid-resume to
// leave a longer (still-resumable) prefix, not a truncated one.
func appendToDst(dst string, localSize int64, resp *http.Response, progress ProgressFunc) (int64, error) {
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return 0, fmt.Errorf("open %s for append: %w", dst, err)
	}
	defer f.Close()

	// total = current local size + remaining. Prefer the parsed
	// Content-Range total when present; fall back to local +
	// Content-Length when not.
	total := totalFromContentRange(resp)
	if total < 0 {
		if cl := resp.ContentLength; cl >= 0 {
			total = localSize + cl
		}
	}

	// Wrap the progress fn so the caller sees cumulative bytes for the
	// file (local + new), not just the new tail — matches what `wget
	// -c` / `curl -C -` show.
	var wrapped ProgressFunc
	if progress != nil {
		wrapped = func(written, t int64) {
			progress(localSize+written, t)
		}
	}
	written, copyErr := copyWithProgress(f, resp.Body, total, wrapped)
	if copyErr != nil {
		return written, copyErr
	}
	return written, nil
}

// copyWithProgress is io.Copy with a 64 KiB buffer + a per-buffer
// progress callback. Throttled to one callback per full 64 KiB read
// (or the final short read), which keeps the CLI's terminal output
// reasonable on fast networks without losing fidelity on slow ones.
func copyWithProgress(dst io.Writer, src io.Reader, total int64, progress ProgressFunc) (int64, error) {
	const bufSize = 64 * 1024
	buf := make([]byte, bufSize)
	var written int64
	for {
		n, rerr := src.Read(buf)
		if n > 0 {
			nw, werr := dst.Write(buf[:n])
			written += int64(nw)
			if werr != nil {
				return written, werr
			}
			if nw < n {
				return written, io.ErrShortWrite
			}
			if progress != nil {
				progress(written, total)
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				return written, nil
			}
			return written, rerr
		}
	}
}

// contentLengthOrTotal returns the file's expected total size, derived
// from the response. For a 200 it's just Content-Length; for a 206
// it's parsed from `Content-Range: bytes <s>-<e>/<total>`. Returns -1
// when neither is informative (chunked transfer with no length).
func contentLengthOrTotal(resp *http.Response) int64 {
	if resp.StatusCode == http.StatusPartialContent {
		if t := totalFromContentRange(resp); t >= 0 {
			return t
		}
	}
	if resp.ContentLength >= 0 {
		return resp.ContentLength
	}
	return -1
}

// totalFromContentRange parses the `/<total>` suffix of a Content-Range
// header. Returns -1 when the header is missing / malformed / `*`.
func totalFromContentRange(resp *http.Response) int64 {
	cr := resp.Header.Get("Content-Range")
	if cr == "" {
		return -1
	}
	idx := strings.LastIndex(cr, "/")
	if idx < 0 || idx == len(cr)-1 {
		return -1
	}
	tail := cr[idx+1:]
	if tail == "*" {
		return -1
	}
	n, err := strconv.ParseInt(tail, 10, 64)
	if err != nil {
		return -1
	}
	return n
}

// StreamRaw issues GET /api/raw/<encPlainPath>?inline=true and copies
// the response body to `w`. Used by `files cat` so the body lands on
// stdout without ever being fully buffered (a 4 GiB file is a valid
// `cat` target on the wire even if it's a poor choice on the human
// side).
//
// `inline=true` mirrors what the LarePass web app's
// formatFileContent / preview pipelines do (data.ts in v2/drive). It
// only changes Content-Disposition on the response, but we keep it
// because some files-backend code paths key off it for cache headers.
//
// Errors:
//   - non-2xx surfaces as *HTTPError (same shape as DownloadFile).
//     400 is the "not a file" code from raw_service.go; the caller
//     should Stat first to give the user a friendlier error than the
//     server's "not a file, path: ..." message.
func (c *Client) StreamRaw(ctx context.Context, plainPath string, w io.Writer) (int64, error) {
	endpoint := c.rawURL(plainPath) + "?inline=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	if c.AccessToken != "" {
		req.Header.Set("X-Authorization", c.AccessToken)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return 0, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodGet,
		}
	}
	return io.Copy(w, resp.Body)
}

func (o *Options) normalize() {
	if o.MaxRetries == 0 {
		o.MaxRetries = DefaultMaxRetries
	}
	if o.RetryBackoff == 0 {
		o.RetryBackoff = DefaultRetryBackoff
	}
}
