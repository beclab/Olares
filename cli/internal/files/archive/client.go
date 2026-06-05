// Package archive implements the wire side of the per-user
// files-backend's `/api/archive/<node>/` surface, which powers
// `olares-cli files compress` / `extract` / `archive entries` /
// `archive cat`.
//
// Four endpoints share a common shape:
//
//	POST /api/archive/<node>/compress  →  {task_id}  (async, server task queue)
//	POST /api/archive/<node>/extract   →  {task_id}  (async, server task queue)
//	GET  /api/archive/<node>/entries   →  application/x-ndjson stream
//	GET  /api/archive/<node>/entry     →  application/octet-stream (single entry bytes)
//
// Per the backend contract (the document the cobra layer mirrors
// into help text):
//
//   - All four require the `X-Bfl-User` header. We DO NOT set that
//     here: the per-user `files.<terminus>` host is already scoped
//     to the active Olares ID and the edge resolves the user from
//     `X-Authorization` (see pkg/cmdutil/factory.go's
//     refreshingTransport for the auth recipe). If a future
//     deployment requires us to send X-Bfl-User explicitly, plumb
//     it through `Options.UserHeader` rather than baking a value
//     into the package.
//   - `X-Archive-Password` is the password slot for encrypted
//     archives. Only meaningful for zip / 7z; the backend ignores
//     it for other formats. The CLI never accepts a password via
//     argv — pass it through stdin and let the cobra layer plumb
//     it into the option struct.
//   - All four URLs carry a single `{node}` segment. Compress /
//     extract write through the per-node task queue, so picking
//     the right node matters (the LarePass web app uses the
//     destination's master node — same cascade `files cp` uses).
//     Entries / entry are read-only stream endpoints and can use
//     any node that mounts the archive's volume; the cobra layer
//     defaults to the first /api/nodes/ entry, consistent with cp.
//
// This file holds the shared scaffolding — Client / HTTPError /
// URL builders / do() helper. The verb-specific operations live in
// compress.go / extract.go / entries.go / entry.go, and the task
// polling helper used by `--wait` lives in task.go.
package archive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// debugArchiveURL prints the supplied URL (and a per-verb label) to
// stderr when the env var OLARES_CLI_DEBUG_ARCHIVE_URL is set. Used
// by entries / entry / compress / extract so the user can confirm
// the exact wire we're hitting when iterating on the spec's query-
// parameter shape.
//
// Cheap, no-op when disabled. We don't promote this to a real
// --debug flag because the audience is "spec-mismatch debugging by
// the CLI author / power user", not end users — env-var gating keeps
// the surface unsurprising for everyone else.
func debugArchiveURL(label, u string) {
	if os.Getenv("OLARES_CLI_DEBUG_ARCHIVE_URL") == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "[olares-cli archive %s] %s\n", label, u)
}

// envOr returns the env value if non-empty, else the supplied default.
// Used by the archive query-shape overrides below — declared in
// client.go (not buried in entries.go) because entry.go shares them
// and we want the override matrix documented in one place.
func envOr(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return fallback
}

// archiveQueryShape holds the four knobs that let the user iterate
// on the wire format of the entries / entry endpoints without
// recompiling. Defaults are our best read of the spec:
//
//	OLARES_ARCHIVE_PARAM_SOURCE        name of the archive-path
//	                                   query parameter.
//	                                   Default: "source"
//	                                   (also plausible: "path",
//	                                   "src", "file", "file_path").
//
//	OLARES_ARCHIVE_PARAM_ENTRY         name of the inner-entry-path
//	                                   query parameter on /entry.
//	                                   Default: "path"
//	                                   (also plausible: "entry",
//	                                   "name", "file").
//
//	OLARES_ARCHIVE_STRIP_LEADING_SLASH strip the leading "/" from
//	                                   the archive path before
//	                                   encoding. Default off;
//	                                   set to "1" to mirror the
//	                                   /api/resources/<path>
//	                                   convention.
//
//	OLARES_ARCHIVE_PATH_ENCODE         encoder for query values:
//	                                     "component" (default,
//	                                       EncodeURIComponent —
//	                                       slashes become %2F),
//	                                     "segment"   (EncodeURL —
//	                                       slashes preserved
//	                                       per segment),
//	                                     "raw"       (no encoding;
//	                                       only safe when values
//	                                       have no reserved chars).
//
// Once the spec's "Query 参数" table is settled and CI runs against
// a real backend, this whole matrix should collapse into the wire
// shape that works — keep only the constants and delete the env-var
// plumbing.
type archiveQueryShape struct {
	sourceParam      string
	entryParam       string
	stripLeadingSlsh bool
	pathEncoder      string
}

func currentArchiveQueryShape() archiveQueryShape {
	return archiveQueryShape{
		sourceParam:      envOr("OLARES_ARCHIVE_PARAM_SOURCE", "source"),
		entryParam:       envOr("OLARES_ARCHIVE_PARAM_ENTRY", "path"),
		stripLeadingSlsh: os.Getenv("OLARES_ARCHIVE_STRIP_LEADING_SLASH") == "1",
		pathEncoder:      envOr("OLARES_ARCHIVE_PATH_ENCODE", "component"),
	}
}

// encodePathValue applies the requested encoder to a query value.
// Defaults to EncodeURIComponent (matches the LarePass web app and
// upload's `?file_path=` convention).
func (s archiveQueryShape) encodePathValue(v string) string {
	if s.stripLeadingSlsh {
		v = strings.TrimPrefix(v, "/")
	}
	switch s.pathEncoder {
	case "raw":
		return v
	case "segment":
		return encodepath.EncodeURL(v)
	default:
		// "component" — slashes become %2F.
		return encodepath.EncodeURIComponent(v)
	}
}

// Client is the per-FilesURL handle the archive verbs share. Cheap
// to construct; reuse one per cobra invocation.
//
// HTTPClient is expected to be the factory-provided client whose
// refreshingTransport injects `X-Authorization` (not `Authorization:
// Bearer`, see pkg/cmdutil/factory.go for why) and transparently
// refreshes on 401/403. For the streaming verbs (entries / entry)
// the caller should pass `f.HTTPClientWithoutTimeout(ctx)` because
// archive walks can take longer than the standard 30 s budget — a
// large 7z with many entries easily blows past that.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string // FilesURL, e.g. https://files.alice.olares.com
}

// HTTPError carries the status + truncated body of a non-2xx
// response so the cobra layer can branch on the status code (e.g.
// to give a friendly "not found" or auth-issue CTA). Same shape as
// the other per-package HTTP errors in this CLI to keep the error
// contract uniform.
//
// Streaming verbs may also surface a non-2xx error AFTER the HTTP
// status was 200 — entries' NDJSON encodes mid-stream errors as
// in-band sentinel lines (see entries.go). Those are a separate
// typed error (*EntriesStreamError) and NOT modelled here.
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

// IsHTTPStatus reports whether `err` is a *HTTPError with the given
// status. Convenience for callers (and reformatters) that branch on
// 401 / 403 / 404. Robust to wrapped errors via errors.As at the
// callsite — `archive.IsHTTPStatus(err, 404)` matches even when err
// has been wrapped by fmt.Errorf.
func IsHTTPStatus(err error, status int) bool {
	var h *HTTPError
	if asHTTPError(err, &h) {
		return h.Status == status
	}
	return false
}

// asHTTPError tries to unwrap `err` into *HTTPError. Tiny helper so
// callers don't have to import "errors" alongside this package for
// the common "is this a 404?" branch.
func asHTTPError(err error, dst **HTTPError) bool {
	// Walk the wrap chain manually rather than depending on
	// errors.As to keep this package's import set tight. Same
	// shape as the predicate used in download.IsNotFound.
	for cur := err; cur != nil; {
		if h, ok := cur.(*HTTPError); ok {
			*dst = h
			return true
		}
		// errors.Unwrap analogue without the stdlib import — Go's
		// fmt.Errorf wrapping is the only producer in this
		// codebase, and *fmt.wrapError exposes Unwrap().
		type unwrapper interface{ Unwrap() error }
		u, ok := cur.(unwrapper)
		if !ok {
			return false
		}
		cur = u.Unwrap()
	}
	return false
}

// archiveURL returns `<BaseURL>/api/archive/<node>/<verb>`. `verb`
// must be one of "compress" / "extract" / "entries" / "entry"; the
// package's exported entry points are the only legitimate callers,
// so we don't waste a `switch verb` here — a typo would surface as
// the server's 404.
//
// `node` is escaped via url.PathEscape; "/" is rare in node names
// but the LarePass web app's `urlEncoded(node)` also escapes it, so
// we stay consistent.
func (c *Client) archiveURL(node, verb string) string {
	return c.BaseURL + "/api/archive/" + url.PathEscape(node) + "/" + verb
}

// passwordHeader applies the `X-Archive-Password` header to req
// when `password` is non-empty. Centralised so the four verbs all
// inject the password through one code path — a hot spot for
// "did I forget to set the header on the new verb?" regressions.
//
// We do NOT enforce the "only zip/7z carry a password" rule here —
// the backend ignores the header for other formats, and the cobra
// layer's format-validator surfaces the rule with a clear error.
// Splitting the rule between two layers would leave the package
// without a single source of truth for the contract.
func passwordHeader(req *http.Request, password string) {
	if password == "" {
		return
	}
	req.Header.Set("X-Archive-Password", password)
}

// trimSlashes drops leading and trailing '/' from a path-shape
// string. Used by the wire-shape helpers in compress.go / extract.go
// when stitching `<fileType>/<extend>/<sub>` into the canonical
// `/<fileType>/<extend>/<sub>` form the backend expects.
func trimSlashes(s string) string {
	return strings.Trim(s, "/")
}

// readBoundedBody reads at most `limit` bytes from r so an error
// response from a misconfigured server (HTML 500 / multi-MB stack
// trace) doesn't bloat the error message. Tail bytes beyond the
// limit are discarded.
//
// Returns the bytes read and any io error from the first
// `limit`-byte slice. Callers wrap that into HTTPError.Body
// directly — the truncation marker is added by HTTPError.Error()
// itself, so we keep the body verbatim here.
func readBoundedBody(r io.Reader, limit int64) []byte {
	if limit <= 0 {
		return nil
	}
	buf, _ := io.ReadAll(io.LimitReader(r, limit))
	// Drain the rest so the underlying connection can be reused
	// (matters most for non-stream endpoints where the body is
	// small; for stream endpoints the caller closes early on
	// error and the connection is reset anyway).
	_, _ = io.Copy(io.Discard, r)
	return buf
}

// httpErrorFromResponse builds a *HTTPError from a non-2xx
// response. Centralised so every verb attaches the same fields
// (Status / Body / URL / Method) and the cobra layer's status
// switch sees a uniform shape regardless of which endpoint
// failed.
func httpErrorFromResponse(resp *http.Response, endpoint, method string) error {
	body := readBoundedBody(resp.Body, 1<<16) // 64 KiB cap
	return &HTTPError{
		Status: resp.StatusCode,
		Body:   string(body),
		URL:    endpoint,
		Method: method,
	}
}

// do executes a single HTTP request and reads the full response
// body on 2xx, just like upload.Client.do but with the
// archive-specific header injection (X-Archive-Password) wired in.
//
// Used by compress / extract (small JSON request + small JSON
// response). The streaming verbs (entries / entry) bypass this
// helper and stream resp.Body directly so the caller can consume
// NDJSON / octet-stream incrementally.
func (c *Client) do(
	ctx context.Context,
	method, endpoint string,
	body io.Reader,
	contentType, password string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	passwordHeader(req, password)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, httpErrorFromResponse(resp, endpoint, method)
	}
	return io.ReadAll(resp.Body)
}

// truncateBody is the same helper rm/upload/cp use to keep error
// messages bounded. Hoisted out of the per-verb files so all four
// archive verbs share one truncation strategy.
func truncateBody(b []byte) string {
	if len(b) <= 200 {
		return string(b)
	}
	return string(b[:200]) + "...(truncated)"
}
