// Package edit implements the wire side of `olares-cli files edit`,
// hitting the per-user files-backend's in-place file-write endpoint:
//
//	PUT /api/resources/<fileType>/<extend><subPath>
//	    Content-Type: text/plain
//	    <body: full file contents>
//
// This mirrors the LarePass web app's per-driver `saveFile` /
// `updateFile` / `put` helpers (apps/.../api/files/v2/{drive,sync,
// cache,external}/utils.ts) — they all funnel into the same PUT
// against `/api/resources/...` with the new file contents in the
// body and `Content-Type: text/plain`. The endpoint replaces the
// file's contents wholesale; there is no patch/diff API on the
// wire.
//
// Cloud drives (awss3 / google / dropbox / tencent) are NOT
// supported. The FETCH leg is now fine — the unified
// `/api/raw/<fileType>/<extend><subPath>?inline=true` endpoint
// that `files cat` / `files download` use returns raw bytes for
// cloud namespaces too, so the historical "preview JSON envelope"
// risk that earlier drafts of this package called out is no longer
// in play. The remaining gap is the WRITE leg:
//
//   - Only `awss3/utils.ts` exports a `put()` helper that calls
//     `/api/resources/...`. `google/utils.ts` and
//     `dropbox/utils.ts` have NO save-related helper at all, so
//     PUT-ing against `/api/resources/<cloud-path>` for those
//     drivers would be against an endpoint the upstream GUI has
//     never exercised end-to-end — high risk of either silent
//     drop or partial-write corruption on the cloud bridge side.
//
// Until the PUT shape is wire-verified for each cloud driver
// (`tencent` lacks the upload helper too), the safe answer is to
// refuse cloud namespaces here and point users at the proven
// `download` + edit-locally + `upload` round-trip. The flip is one
// allow-list entry away when the wire shapes are signed off.
//
// `share` and `internal` namespaces are likewise refused: the
// LarePass UX exposes them as cross-user / read-only views with no
// "save" affordance, and we don't have a wire-shape signoff for
// either path.
//
// Same X-Authorization injection convention as the rest of olares-
// cli (see pkg/cmdutil/factory.go for why the per-user files
// surface uses `X-Authorization`, NOT `Authorization: Bearer`).
package edit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Target is the parsed view of the user-supplied remote path,
// normalized so the planner has one canonical shape to validate.
// Construct via the cobra layer's path parser; this package
// intentionally doesn't import FrontendPath so it stays free of
// cobra/cmdutil deps — same convention as the cp / rm / rename /
// mkdir packages.
type Target struct {
	// FileType + Extend + SubPath together form the wire path
	// (joined with '/' and percent-encoded per segment): e.g.
	// ("drive", "Home", "/Documents/notes.md") →
	// /api/resources/drive/Home/Documents/notes.md.
	FileType string
	Extend   string
	// SubPath always starts with '/'. A trailing '/' is rejected
	// at Plan time — `edit` is a per-FILE verb and a directory
	// path here is almost certainly user error (the GUI's edit
	// affordance is hidden on directories too).
	SubPath string
}

// Op is one PUT /api/resources/.../ call, fully resolved.
// Endpoint is the URL relative to BaseURL (already percent-encoded
// per encodepath.EncodeURL — same convention as cp / rm / rename /
// download / upload), so the http call site doesn't re-encode.
//
// DisplayPath is the human-readable 3-segment frontend form (e.g.
// `drive/Home/Documents/notes.md`) surfaced in log lines and error
// messages.
type Op struct {
	Endpoint    string
	DisplayPath string
}

// Client is the per-FilesURL handle for edit-related calls
// (download via /api/raw + upload via PUT /api/resources). It is
// cheap to construct; reuse one per `files edit` invocation.
//
// HTTPClient is expected to be the factory-provided client whose
// refreshingTransport injects `X-Authorization` (NOT `Authorization:
// Bearer`, see pkg/cmdutil/factory.go for why) and refreshes on
// 401/403 transparently.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// HTTPError carries the status + truncated body of a non-2xx
// response so the cobra layer can branch on the status code (e.g.
// 401 / 403 → re-login CTA; 404 → "did you mean ...?"). Same shape
// as the per-package HTTP errors in cp / rm / rename / mkdir /
// download / upload — keeps the error contract uniform across
// verbs.
type HTTPError struct {
	Status int
	Body   string
	URL    string
	Method string
}

func (e *HTTPError) Error() string {
	body := e.Body
	if len(body) > 500 {
		body = body[:500] + "...(truncated)"
	}
	return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.URL, e.Status, body)
}

// supportedFileTypes is the allow-list of namespaces whose BOTH
// legs (fetch via /api/raw + writeback via PUT /api/resources)
// are wire-verified end-to-end. We deliberately keep it tight:
// the LarePass GUI itself wires `onSaveFile` to these four
// (utils.ts `put` / `saveFile` / `updateFile` in
// apps/.../api/files/v2/{drive,sync,cache,external}), so both
// directions have been exercised in production.
//
// Cloud drives (awss3 / google / dropbox / tencent) are NOT in
// the allow-list — see the package docstring. The fetch leg is
// fine (the unified /api/raw/ endpoint serves cloud bytes too)
// but the writeback PUT shape is unverified per driver
// (google / dropbox / tencent have no GUI put helper). `share`
// and `internal` are likewise excluded — they're cross-user /
// read-only views in the LarePass UX.
var supportedFileTypes = map[string]struct{}{
	"drive":    {},
	"sync":     {},
	"cache":    {},
	"external": {},
}

// SupportedFileTypesList is the alphabetically-sorted comma-joined
// rendering of supportedFileTypes for error messages. Computed
// once so the (cold) refusal path doesn't allocate on every Plan
// call.
const SupportedFileTypesList = "cache, drive, external, sync"

// Plan validates the inputs and returns a single Op for one PUT
// call. Validation rules:
//
//   - FileType / Extend must be non-empty (defense in depth — the
//     cobra-layer parser shouldn't let these through, but a typed
//     error here beats a silent /api/resources//Home/... URL).
//   - FileType must be one of the supported namespaces (drive /
//     sync / cache / external). Cloud drives + share + internal
//     are rejected with a self-describing error pointing at the
//     limitation.
//   - SubPath, after trimming '/', must be non-empty — refusing to
//     "edit" the volume root mirrors the rest of the CLI's safety
//     policy (rm / rename / cp all refuse the root) and the wire
//     endpoint would have nothing meaningful to PUT bytes into
//     anyway.
//   - SubPath must NOT end with '/' — directories aren't editable
//     and the trailing slash would route the request through the
//     directory handler on the server side.
//   - No segment may be '.' or '..' — same path-traversal
//     blacklist mkdir / rename enforce on their cleaned inputs.
func Plan(t Target) (Op, error) {
	if t.FileType == "" || t.Extend == "" {
		return Op{}, fmt.Errorf("edit: empty fileType or extend (got %q/%q)", t.FileType, t.Extend)
	}
	if _, ok := supportedFileTypes[t.FileType]; !ok {
		// Cloud drives (awss3 / google / dropbox / tencent) get a
		// targeted message — the surface looks like an arbitrary
		// allow-list miss otherwise, but the actual reason is
		// concrete: the writeback (PUT /api/resources/<cloud-path>)
		// has no wire-shape signoff per driver (only awss3 has a
		// `put()` helper in its v2 utils; google / dropbox / tencent
		// have no save-related helper at all). The fetch leg is now
		// uniform (the unified /api/raw/ endpoint serves cloud bytes
		// too), so a future flip is just an allow-list entry once
		// the PUT shape is verified. Until then the recovery path
		// is the download → edit-locally → upload round-trip the
		// CLI already supports end-to-end.
		switch t.FileType {
		case "awss3", "google", "dropbox", "tencent":
			return Op{}, fmt.Errorf(
				"edit: cloud-drive namespace %q is not supported end-to-end "+
					"(PUT /api/resources/<cloud-path> is not wire-verified for cloud drivers — "+
					"only awss3's v2 utils exports a save helper, and writeback shape is unconfirmed); "+
					"safe alternative: `files download %s/<path> <local>` → edit locally → "+
					"`files upload <local> %s/<path>`",
				t.FileType, t.FileType, t.FileType)
		}
		return Op{}, fmt.Errorf(
			"edit: fileType %q is not supported (supported: %s); "+
				"the LarePass web app's edit flow is wired only for these namespaces",
			t.FileType, SupportedFileTypesList)
	}
	clean := strings.Trim(t.SubPath, "/")
	if clean == "" {
		return Op{}, fmt.Errorf(
			"refusing to edit the root of %s/%s: pick a file path (e.g. %s/%s/notes.md)",
			t.FileType, t.Extend, t.FileType, t.Extend)
	}
	if strings.HasSuffix(t.SubPath, "/") {
		return Op{}, fmt.Errorf(
			"edit: %s/%s%s is a directory path (trailing '/'); edit only works on files",
			t.FileType, t.Extend, t.SubPath)
	}
	for _, seg := range strings.Split(clean, "/") {
		if seg == "." || seg == ".." {
			return Op{}, fmt.Errorf(
				"edit: path segment %q is invalid ('.' / '..' are blocked by the path-traversal blacklist): %s",
				seg, t.FileType+"/"+t.Extend+t.SubPath)
		}
	}

	plain := t.FileType + "/" + t.Extend + "/" + clean
	return Op{
		Endpoint:    "/api/resources/" + encodepath.EncodeURL(plain),
		DisplayPath: t.FileType + "/" + t.Extend + "/" + clean,
	}, nil
}

// DefaultContentType is what the web app's saveFile / updateFile /
// put helpers send for text editing. The backend stores the bytes
// verbatim regardless of the type — this is mostly a hint for any
// content-aware caching layer between us and the storage driver.
const DefaultContentType = "text/plain"

// Put PUTs `body` against op.Endpoint with the supplied
// Content-Type. `contentLength` is the number of bytes available
// in `body`; pass -1 if you don't know (the caller should normally
// know, since edit always has the bytes in memory). When
// contentLength >= 0 we set Content-Length explicitly so the
// server doesn't fall back to chunked encoding for what's
// typically a tiny config / note file.
//
// Errors:
//   - non-2xx response → *HTTPError so the cobra layer can branch
//     on 401 / 403 / 404 / 409 with friendly CTAs.
//   - request build / network failures bubble up verbatim.
func (c *Client) Put(
	ctx context.Context,
	op Op,
	body io.Reader,
	contentLength int64,
	contentType string,
) error {
	if op.Endpoint == "" {
		return errors.New("Put: empty Endpoint (Plan should have rejected this)")
	}
	if contentType == "" {
		contentType = DefaultContentType
	}

	endpoint := c.BaseURL + op.Endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	if contentLength >= 0 {
		req.ContentLength = contentLength
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return &HTTPError{
			Status: resp.StatusCode,
			Body:   string(respBody),
			URL:    endpoint,
			Method: http.MethodPut,
		}
	}
	return nil
}

// PutBytes is a convenience wrapper around Put for the common case
// of "I have the new file contents in memory as []byte". It sets
// Content-Length from len(content) and rewinds via bytes.NewReader
// so the request body is replayable on transparent token-refresh
// retries (the factory's refreshingTransport relies on
// req.GetBody — bytes.NewReader populates that automatically via
// http.NewRequestWithContext).
func (c *Client) PutBytes(
	ctx context.Context,
	op Op,
	content []byte,
	contentType string,
) error {
	return c.Put(ctx, op, bytes.NewReader(content), int64(len(content)), contentType)
}

// rawURL returns `<BaseURL>/api/raw/<encPlainPath>`. Same wire
// endpoint download / cat use to fetch the raw bytes — see
// internal/files/download/client.go's rawURL for the rationale.
//
// We re-implement here instead of importing the download package
// because each verb under cli/internal/files/ is intentionally
// self-contained (cp / rm / rename / mkdir / share / repos all
// follow the same convention). One percent-encoded URL helper
// duplicated across 7 packages is a smaller cost than the
// cross-package coupling, and it keeps an edit package refactor
// from rippling through the rest.
func (c *Client) rawURL(plainPath string) string {
	return c.BaseURL + "/api/raw/" + encodepath.EncodeURL(plainPath)
}

// TooLargeError is returned by Fetch when the remote body's
// length exceeds the caller-supplied maxBytes ceiling. It's a
// distinct type (rather than a wrapped HTTPError) because the
// over-the-wire response was healthy — it's the size policy that
// rejected the body. The cobra layer formats this into a
// `--max-size` CTA at the verb-level call site.
//
// Limit and Read are best-effort: Read may be Limit+1 in the
// LimitReader-driven detection path (we read one byte past the
// cap to detect overflow without burning unbounded memory), so
// callers should treat Read as "at least this many bytes" rather
// than an exact body size.
type TooLargeError struct {
	Read  int64
	Limit int64
}

func (e *TooLargeError) Error() string {
	return fmt.Sprintf("response body exceeds %d bytes (read at least %d)", e.Limit, e.Read)
}

// Fetch GETs the current contents of the remote file at `plainPath`
// (an un-encoded `<fileType>/<extend>/<sub>` triple, no leading
// '/') and returns them as a buffered byte slice. We buffer
// rather than streaming because:
//
//   - The cobra layer needs the bytes whole to write the temp
//     file AND to compare them against the post-edit bytes;
//     streaming twice would double the wire cost.
//   - Edit's typical input is a config / note file (KB-MB range);
//     the buffer cost is negligible.
//
// `maxBytes > 0` activates a hard ceiling: the read is wrapped
// in `io.LimitReader(body, maxBytes+1)` so the buffer never
// grows past maxBytes+1 bytes regardless of what the server
// claims (or fails to claim) about Content-Length. If the read
// returns more than maxBytes bytes, we error out with
// *TooLargeError BEFORE the cobra layer sees the body — that
// way a misreported Stat.Size (e.g. 0 from a backend that
// didn't fill the field) can't surprise us with an unbounded
// download. `maxBytes == 0` disables the ceiling for the rare
// case where the caller really does want to slurp whatever the
// server returns.
//
// Errors:
//   - non-2xx response → *HTTPError, same as Put.
//   - 404 in particular is preserved with its original Status so
//     the cobra layer can distinguish "file genuinely missing"
//     (route to the upload CTA — `edit` is update-only) from a
//     concurrent-delete race (Stat OK then Fetch 404) via
//     IsNotFound().
//   - body length > maxBytes (when maxBytes > 0) → *TooLargeError.
func (c *Client) Fetch(ctx context.Context, plainPath string, maxBytes int64) ([]byte, error) {
	endpoint := c.rawURL(plainPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "*/*")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bodyReader io.Reader = resp.Body
	if maxBytes > 0 {
		// +1 so we can DETECT overflow: a read that returns
		// exactly maxBytes is at the boundary (OK); a read
		// that returns maxBytes+1 means the server had more
		// for us and we want to refuse rather than truncate.
		bodyReader = io.LimitReader(resp.Body, maxBytes+1)
	}
	body, _ := io.ReadAll(bodyReader)
	if resp.StatusCode/100 != 2 {
		return nil, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodGet,
		}
	}
	if maxBytes > 0 && int64(len(body)) > maxBytes {
		return nil, &TooLargeError{Read: int64(len(body)), Limit: maxBytes}
	}
	return body, nil
}

// IsNotFound reports whether `err` is a 404 from any of this
// package's wire calls. Used by the cobra layer to distinguish a
// genuinely missing file (route to the upload CTA — `edit` is
// strictly UPDATE-only) from a concurrent-delete race window
// (Stat said the file existed but Fetch came back 404, surfaced
// as a friendly conflict error rather than a silent recreate).
func IsNotFound(err error) bool {
	var hErr *HTTPError
	if errors.As(err, &hErr) {
		return hErr.Status == http.StatusNotFound
	}
	return false
}

// IsHTTPStatus is a convenience predicate the cobra layer uses to
// branch on common 4xx codes. Same shape as cp.IsHTTPStatus / etc.;
// duplicated here to keep this package self-contained.
func IsHTTPStatus(err error, status int) bool {
	var hErr *HTTPError
	return errors.As(err, &hErr) && hErr.Status == status
}
