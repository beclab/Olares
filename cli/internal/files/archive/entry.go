// entry.go: wire side of GET /api/archive/<node>/entry.
//
// Returns the raw bytes of a single entry inside an archive,
// without ever materialising the full archive contents on
// disk. The response Content-Type is application/octet-stream,
// Content-Disposition: attachment; filename="<basename>" — so
// the caller's writer ends up with exactly the entry's bytes,
// nothing else.
//
// Errors before the body starts streaming surface as a standard
// JSON error envelope { "error": "...", "code": "..." }; we
// model them through *HTTPError so the cobra layer's reformatter
// branches uniformly on status. Mid-stream errors AFTER the
// status line has been sent show up as a truncated body and a
// non-EOF read error — there's no in-band sentinel here (the
// spec promises octet-stream, not a framed format).
package archive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// EntryOptions captures the query parameters of GET
// /api/archive/<node>/entry.
type EntryOptions struct {
	// Source is the wire path of the archive containing the
	// target entry. Sent on the wire as `source=`.
	Source string

	// Path is the in-archive path of the entry to fetch. Use
	// the slash-separated form the server emitted on the
	// `entries` stream (no leading slash, no trailing slash
	// on file entries). Sent on the wire as `path=`.
	Path string

	// Format is the locally-resolved archive container — used
	// for local checks at the cobra layer (e.g. gating
	// `--password-stdin` to zip / 7z). NOT sent on the wire:
	// per spec §4 the only query parameters are `source` and
	// `path`, and the server infers the format from the
	// source's extension.
	Format string

	// Node is the {node} URL segment.
	Node string
}

// EntryDownload describes the response metadata the cobra layer
// renders alongside the actual byte stream (e.g. for the
// `archive cat -o <local>` shape that resolves the local
// filename from the server's suggestion). The caller can ignore
// the Filename hint if they have their own target path.
type EntryDownload struct {
	// Filename is the basename the server suggested via
	// Content-Disposition. Empty when the header is missing /
	// malformed (some proxies strip it).
	Filename string

	// ContentLength is the byte count the server promised via
	// Content-Length when set; -1 when unknown (chunked
	// transfer or omitted header). The cobra layer uses this
	// to print "X bytes" progress lines without a Stat.
	ContentLength int64

	// BytesWritten is the count Stream copied into the caller's
	// writer. Always set after a successful Stream.
	BytesWritten int64
}

// StreamEntry streams a single archive entry's raw bytes into
// `w` and returns metadata describing the transfer.
//
// Errors:
//
//   - non-2xx HTTP status surfaces as *HTTPError. Special case:
//     when the body is the documented `{error, code}` JSON shape,
//     we map it onto *EntryError so the cobra layer's
//     reformatter can give a friendlier CTA than the truncated
//     JSON dump.
//   - io errors while copying the body surface as the raw
//     io/net error (no wrapping) so the caller can
//     errors.Is-check for context.Canceled / io.UnexpectedEOF.
func (c *Client) StreamEntry(
	ctx context.Context,
	opts EntryOptions,
	password string,
	w io.Writer,
) (EntryDownload, error) {
	if err := validateEntryOptions(opts); err != nil {
		return EntryDownload{}, err
	}
	if w == nil {
		return EntryDownload{}, errors.New("StreamEntry: w is required")
	}

	endpoint := buildEntryURL(c, opts)
	debugArchiveURL("entry", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return EntryDownload{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	passwordHeader(req, password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return EntryDownload{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return EntryDownload{}, classifyEntryError(resp, endpoint)
	}

	dl := EntryDownload{
		Filename:      parseContentDispositionFilename(resp.Header.Get("Content-Disposition")),
		ContentLength: resp.ContentLength,
	}
	n, err := io.Copy(w, resp.Body)
	dl.BytesWritten = n
	if err != nil {
		// Stream broke mid-copy. Return what we got — the
		// caller may want to know how many bytes did make
		// it before the failure.
		return dl, err
	}
	return dl, nil
}

// EntryError is the typed shape of an in-band error returned by
// GET /api/archive/<node>/entry when the server failed to open
// the entry (path not in archive, password bad, etc.). The HTTP
// status reflects the error class (404 for not_found, 4xx for
// password problems, 500 for internal); we expose Code so the
// cobra layer can map the same enum used by the entries stream
// onto a uniform CTA.
//
// Note: this type wraps an HTTPError so callers that branch on
// status still see the right code via errors.As(*HTTPError),
// while callers that want the human-friendly Code branch on
// errors.As(*EntryError). Both predicates fire on the same
// underlying error.
type EntryError struct {
	*HTTPError
	Code    string
	Message string
}

func (e *EntryError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s [%s]: %s", e.HTTPError.Error(), e.Code, e.Message)
	}
	return e.HTTPError.Error()
}

// Unwrap exposes the wrapped *HTTPError so errors.Is /
// errors.As walk through transparently — the cobra layer's
// 401/403 → "profile login" reformatter still fires when an
// EntryError carries an underlying 401.
func (e *EntryError) Unwrap() error { return e.HTTPError }

// classifyEntryError peels the body of a non-2xx response and
// builds the right typed error. The documented shape is a
// small JSON object with `error` and `code` keys; if the body
// doesn't decode (server returned an HTML 502 / proxy error
// page / raw text), we fall back to the bare HTTPError so the
// status code is at least usable.
func classifyEntryError(resp *http.Response, endpoint string) error {
	body := readBoundedBody(resp.Body, 1<<16)
	httpErr := &HTTPError{
		Status: resp.StatusCode,
		Body:   string(body),
		URL:    endpoint,
		Method: http.MethodGet,
	}
	// Try the documented shape — if it doesn't parse, the bare
	// HTTPError carries everything we have. We don't reject the
	// fall-through case: a misconfigured server (or an upstream
	// gateway between the CLI and files.<terminus>) can ship
	// non-JSON 4xx bodies that we still want to surface.
	var env struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if len(body) > 0 && json.Unmarshal(body, &env) == nil && (env.Error != "" || env.Code != "") {
		return &EntryError{
			HTTPError: httpErr,
			Code:      env.Code,
			Message:   env.Error,
		}
	}
	return httpErr
}

// buildEntryURL stitches the query string for
// /api/archive/<node>/entry.
//
// Wire shape per spec §4:
//
//	GET /api/archive/<node>/entry?source=<archive-URI>&path=<inner>
//
// The spec lists exactly two required query parameters:
//
//   - `source` — wire path of the containing archive.
//   - `path`   — in-archive entry path (round-trips the same
//     JSON key /entries emits for each per-entry line).
//
// Parameter names and path-value encoding are driven by
// archiveQueryShape (see client.go) so the user can flip
// OLARES_ARCHIVE_PARAM_SOURCE / _PARAM_ENTRY / _PATH_ENCODE /
// _STRIP_LEADING_SLASH while the spec's "Query 参数" table is
// settled. Format never travels on the wire — it stays a
// cobra-layer local for password-compat checks.
func buildEntryURL(c *Client, opts EntryOptions) string {
	shape := currentArchiveQueryShape()
	q := shape.sourceParam + "=" + shape.encodePathValue(opts.Source) +
		"&" + shape.entryParam + "=" + shape.encodePathValue(opts.Path)
	return c.archiveURL(opts.Node, "entry") + "?" + q
}

// validateEntryOptions runs the client-side preflight. Single
// path validation here keeps the wire-side surface tight.
func validateEntryOptions(opts EntryOptions) error {
	if opts.Node == "" {
		return errors.New("entry: empty Node (cobra layer should resolve a default via /api/nodes/ or --node)")
	}
	if strings.TrimSpace(opts.Source) == "" {
		return errors.New("entry: source archive path is empty")
	}
	if strings.TrimSpace(opts.Path) == "" {
		return errors.New("entry: in-archive entry path is empty")
	}
	if opts.Format == "" {
		return errors.New("entry: --format is required (one of: " + JoinFormats() + ")")
	}
	if !IsSupportedFormat(opts.Format) {
		return fmt.Errorf("entry: unsupported --format %q; valid formats: %s",
			opts.Format, JoinFormats())
	}
	return nil
}

// parseContentDispositionFilename pulls the basename out of a
// `Content-Disposition: attachment; filename="<basename>"`
// header.
//
// We don't use mime.ParseMediaType because that fails on a
// fairly common quirk: backends emit filenames with characters
// that need RFC 5987 encoding but only escape some of them.
// Our concrete need is "get the basename for display" — a
// best-effort substring extractor is more robust than a strict
// RFC parser for that goal.
func parseContentDispositionFilename(cd string) string {
	if cd == "" {
		return ""
	}
	const key = "filename="
	idx := strings.Index(strings.ToLower(cd), key)
	if idx < 0 {
		return ""
	}
	val := cd[idx+len(key):]
	// Strip optional surrounding quotes. We don't try to handle
	// embedded escaped quotes (`filename="foo\"bar"`) because
	// the file-server's writer doesn't produce them in practice
	// — if a real-world archive ever ships an entry name with
	// a literal `"`, the user can still --output their own
	// filename and ignore the server suggestion.
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "\"") {
		end := strings.Index(val[1:], "\"")
		if end >= 0 {
			return val[1 : 1+end]
		}
		return val[1:]
	}
	// Unquoted: stop at the first ';' (next parameter).
	if i := strings.Index(val, ";"); i >= 0 {
		val = val[:i]
	}
	return strings.TrimSpace(val)
}
