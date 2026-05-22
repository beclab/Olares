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
// Cloud drives awss3 / google / dropbox ARE supported here even
// though the LarePass web app's GUI doesn't currently wire them
// through to a working save. The per-driver API classes
// (Awss3DataAPI / GoogleDataAPI / DropboxDataAPI) extend
// DriveDataAPI without overriding `onSaveFile`, so a "save" in
// the GUI mistakenly routes through `drive.saveFile()` and
// targets `/api/resources/drive/Home<...>` instead of the cloud
// bucket — that's a GUI wiring bug, not a wire-shape limitation.
// On the wire, `PUT /api/resources/<fileType><subPath>` is the
// uniform write endpoint for every namespace the backend's
// resources handler covers, and awss3/utils.ts already exports a
// `put()` helper that calls exactly that shape. The CLI uses
// the same wire shape directly, so a `files edit awss3/<acc>/...`
// hits the cloud bucket by going around the GUI's broken
// `onSaveFile` plumbing.
//
// Tencent is the lone cloud-drive holdout: its UPLOAD path uses
// a separate `/drive/create_direct_upload_task` +
// `/drive/direct_upload_file` octet protocol that the standard
// resources handler doesn't share, and we don't yet have a
// wire-shape signoff that the small-PUT path against
// `/api/resources/tencent<...>` is honored end-to-end. We keep
// it on the deny-list until that's verified — same conservative
// stance `files upload` takes for tencent.
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

// supportedFileTypes is the allow-list of namespaces whose
// `/api/resources/<fileType><subPath>` PUT endpoint we trust as
// the canonical write surface for editing a single file:
//
//   - drive / sync / cache / external — the LarePass GUI itself
//     wires `onSaveFile` here (utils.ts `put` / `saveFile` /
//     `updateFile`), so the wire shape is well-trodden.
//   - awss3 / google / dropbox — the GUI's `onSaveFile` mistakenly
//     routes these through drive's saveFile (a per-driver class
//     wiring bug, see the package docstring), but the underlying
//     wire endpoint exists uniformly for every cloud driver
//     that's wired into the resources handler. The CLI hits it
//     directly so cloud-bucket text edits actually land in the
//     bucket.
//
// Tencent is intentionally NOT in this list — see the package
// docstring for the upload-protocol divergence that makes us
// conservative there.
var supportedFileTypes = map[string]struct{}{
	"drive":    {},
	"sync":     {},
	"cache":    {},
	"external": {},
	"awss3":    {},
	"google":   {},
	"dropbox":  {},
}

// SupportedFileTypesList is the alphabetically-sorted comma-joined
// rendering of supportedFileTypes for error messages. Computed
// once so the (cold) refusal path doesn't allocate on every Plan
// call.
const SupportedFileTypesList = "awss3, cache, drive, dropbox, external, google, sync"

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
		return Op{}, fmt.Errorf(
			"edit: fileType %q is not supported (supported: %s); "+
				"the LarePass web app's edit flow is wired only for these namespaces — "+
				"cloud-drive / share / internal targets have no working PUT endpoint",
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

// Fetch GETs the current contents of the remote file at `plainPath`
// (an un-encoded `<fileType>/<extend>/<sub>` triple, no leading
// '/') and returns them as a buffered byte slice. We buffer
// rather than streaming because:
//
//   - The cobra layer needs the bytes whole to write the temp
//     file AND to compute the post-edit "did the user actually
//     change anything" hash; streaming twice would double the
//     wire cost.
//   - Edit's typical input is a config / note file (KB-MB range);
//     the buffer cost is negligible. Users editing a multi-GB
//     file via $EDITOR should expect to pay for whatever
//     workflow that implies.
//
// Errors:
//   - non-2xx response → *HTTPError, same as Put.
//   - 404 in particular is preserved with its original Status so
//     the cobra layer's `--create` flag can branch on
//     IsNotFound() to decide whether to start with empty
//     contents.
func (c *Client) Fetch(ctx context.Context, plainPath string) ([]byte, error) {
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
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodGet,
		}
	}
	return body, nil
}

// IsNotFound reports whether `err` is a 404 from any of this
// package's wire calls. Used by the cobra layer's `--create` flag
// to decide whether to proceed with an empty starting buffer
// instead of failing.
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
