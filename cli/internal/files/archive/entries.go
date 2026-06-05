// entries.go: wire side of GET /api/archive/<node>/entries.
//
// This endpoint walks an archive's table of contents and streams
// every entry to the client as application/x-ndjson. The shape
// the spec promises (and the cobra layer surfaces) is:
//
//   Each per-entry line:
//     { "path": "dir/file.txt", "size": 1024, "modified": 1716800000,
//       "is_dir": false, "encrypted": false }
//
//   Final success line:
//     { "_done": true, "total": 42 }
//
//   Mid-stream error line (HTTP is already 200 by this point):
//     { "_error": "...", "code": "<code>" }
//
//   code ∈ { password_invalid, password_required, archive_corrupt,
//            volume_missing, canceled, not_found, internal }
//
// Pre-stream failures (URI parse, reader.Open, bad params) ride
// the standard 400/500 JSON-body path; only AFTER the Walk has
// begun does the server switch to in-band sentinels (it can no
// longer roll back the HTTP status, so it encodes the error as a
// distinguishable JSON object).
//
// Client semantics: we expose StreamEntries(ctx, opts, cb) which
// reads NDJSON line by line, invokes `cb` per entry, and returns
// either the final `_done.total` count OR a typed
// *EntriesStreamError when a `_error` sentinel arrives. Caller's
// `cb` returning a non-nil error aborts the stream cleanly — by
// closing the response body, which propagates to the server and
// trips its "client disconnected → cancel underlying 7z" path
// the spec calls out.
package archive

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Entry is one row in the NDJSON stream — a single archive
// member's metadata. The wire shape is what the spec promises;
// extra fields the server may add in future versions are simply
// ignored by the decoder.
//
// Note on `Modified`: the spec ships unix seconds. We expose the
// raw integer so callers can format it however they want (the
// cobra layer renders it as RFC3339 in `--json=false` mode for
// human readability).
type Entry struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Modified  int64  `json:"modified"`
	IsDir     bool   `json:"is_dir"`
	Encrypted bool   `json:"encrypted"`
}

// EntriesOptions captures the query parameters of GET
// /api/archive/<node>/entries. The wire query string is built in
// buildEntriesURL — kept centralised so the entry.go neighbour
// can mirror the same parameter normalisation.
type EntriesOptions struct {
	// Source is the wire path of the archive to walk.
	// Sent on the wire as `source=`.
	Source string

	// Format is the locally-resolved archive container (zip,
	// 7z, tar, ...). NOT sent on the wire — the spec lists
	// only `source` as a query parameter, and the server
	// infers the format from the source extension. The cobra
	// layer still uses it to gate things like `--password-stdin`
	// (zip / 7z only), so we keep it on the options struct.
	Format string

	// Node is the {node} URL segment. Same default cascade as
	// compress / extract, with the relaxation that any
	// node that can read the source volume works.
	Node string
}

// EntriesStreamError is the typed error returned when the
// server emits a `{_error, code}` sentinel mid-stream. The
// `Code` field is one of the spec's documented values
// (password_invalid / password_required / archive_corrupt /
// volume_missing / canceled / not_found / internal); the cobra
// layer's reformatter maps each onto a friendly CTA (e.g.
// password_required → "pass --password-stdin").
//
// We model this distinctly from HTTPError because:
//
//   - the HTTP transport-level status was 200 (the stream
//     started successfully), so wrapping it in HTTPError
//     would be a lie;
//   - the cobra layer wants to branch on `Code` specifically,
//     not on a status code that doesn't exist for this case;
//   - tests want a stable typed error to assert on.
type EntriesStreamError struct {
	Message string // the server's free-form description
	Code    string // one of the documented enum values
}

func (e *EntriesStreamError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("entries stream error [%s]: %s", e.Code, e.Message)
	}
	return "entries stream error: " + e.Message
}

// Documented `code` values for *EntriesStreamError and
// *EntryError. Callers can import these for a status-code-like
// switch. New codes the server adds are surfaced verbatim
// through Code without a constant — the type stays open.
//
// Per spec:
//
//   - entries (NDJSON `_error` line) emits any of:
//     password_invalid / password_required / archive_corrupt /
//     volume_missing / canceled / not_found / internal
//   - entry (single-shot JSON body) emits the same set plus
//     `entry_too_large` (HTTP 413) which entries cannot, since
//     the per-entry stream is purely metadata.
const (
	CodePasswordInvalid  = "password_invalid"
	CodePasswordRequired = "password_required"
	CodeArchiveCorrupt   = "archive_corrupt"
	CodeVolumeMissing    = "volume_missing"
	CodeCanceled         = "canceled"
	CodeNotFound         = "not_found"
	CodeEntryTooLarge    = "entry_too_large"
	CodeInternal         = "internal"
)

// EntryFunc is the per-entry callback for StreamEntries. Returns
// (continue?, error). Returning a non-nil error aborts the
// stream — the response body is closed, which (per the spec)
// triggers server-side cancellation of the underlying 7z child
// process. Use that for early-exit in interactive previews
// (e.g. user typed Ctrl-C, or a UI counter has reached its
// display cap).
type EntryFunc func(Entry) error

// StreamEntries reads /api/archive/<node>/entries and invokes
// cb for every per-entry line. Returns the `total` count from
// the final `_done` line on success, or the typed error from
// the first `_error` sentinel encountered (whichever comes
// first).
//
// HTTP transport errors (network, 4xx/5xx BEFORE the stream
// started) surface as *HTTPError; auth refresh runs inside the
// transport before the body is touched, so a 401 on the first
// poll either retries transparently or surfaces with the auth
// CTA.
//
// Decoding model: we parse line-by-line with bufio.Scanner so
// memory stays O(longest-line). The default Scanner buffer
// (64 KiB) is too small for archives with absurdly long entry
// paths — we bump it to 1 MiB which comfortably covers every
// real-world path the file-server has ever seen.
func (c *Client) StreamEntries(
	ctx context.Context,
	opts EntriesOptions,
	password string,
	cb EntryFunc,
) (int, error) {
	if err := validateEntriesOptions(opts); err != nil {
		return 0, err
	}
	if cb == nil {
		return 0, errors.New("StreamEntries: cb is required")
	}

	endpoint := buildEntriesURL(c, opts)
	debugArchiveURL("entries", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/x-ndjson")
	passwordHeader(req, password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return 0, httpErrorFromResponse(resp, endpoint, http.MethodGet)
	}

	// Defensive: a server misconfiguration could 200 with the
	// wrong Content-Type. We still try to parse — the wire bytes
	// are the source of truth — but log via the error path if
	// the type header doesn't match what we expect.
	if ct := resp.Header.Get("Content-Type"); ct != "" &&
		!strings.HasPrefix(strings.ToLower(ct), "application/x-ndjson") {
		// Surface as a soft warning by wrapping into the error
		// returned IF parsing fails; otherwise continue. We
		// don't fail eagerly here because some proxies rewrite
		// Content-Type to a generic application/json on cache
		// hits, and the body shape is still NDJSON.
		_ = ct // touched only to make the linter happy if we ever decide to log it
	}

	scanner := bufio.NewScanner(resp.Body)
	// Bump the per-line buffer cap to 1 MiB. The default 64 KiB
	// is too small for archives where a single entry path
	// contains a long suffix tree of nested dirs (common with
	// chunked dataset dumps and node_modules archives).
	scanner.Buffer(make([]byte, 64*1024), 1<<20)

	for scanner.Scan() {
		line := scanner.Bytes()
		// Skip stray empty lines defensively — NDJSON is
		// strictly one-object-per-line, but a server with a
		// flaky writer might emit a stray blank.
		if len(line) == 0 {
			continue
		}
		// Decode into a generic header first so we can branch
		// on `_done` / `_error` sentinels without a second
		// unmarshal of the same bytes.
		var hdr ndjsonLineHeader
		if err := json.Unmarshal(line, &hdr); err != nil {
			return 0, fmt.Errorf("decode ndjson line: %w (line=%s)", err, truncateBody(line))
		}
		switch {
		case hdr.Done != nil && *hdr.Done:
			// Stream succeeded. The total is what the server
			// observed AFTER finishing the walk; callers can
			// use it as a sanity check against their cb's call
			// count.
			return parseTotal(hdr.Total), nil
		case hdr.Error != "":
			return 0, &EntriesStreamError{
				Message: hdr.Error,
				Code:    hdr.Code,
			}
		}
		// Regular per-entry line — re-decode into the typed
		// Entry shape. We could parse straight from `hdr`'s
		// extra fields via a tagged union, but a second
		// unmarshal is cleaner and the per-line overhead is
		// negligible compared to the bufio + HTTP read cost.
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return 0, fmt.Errorf("decode entry line: %w (line=%s)", err, truncateBody(line))
		}
		if err := cb(e); err != nil {
			// Caller aborted — close the body promptly so the
			// server's "client disconnected → cancel 7z" path
			// fires. The defer above takes care of the close;
			// returning early is sufficient signal.
			return 0, err
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("read ndjson stream: %w", err)
	}
	// EOF without a `_done` sentinel means the server hung up
	// mid-stream without summarising. The spec doesn't promise
	// `_done` is always emitted (think of an EOF during a
	// "canceled" walk), but the absence is informative — surface
	// it so callers can distinguish a clean walk from a
	// truncated one.
	return 0, errors.New("entries stream ended without a `_done` sentinel")
}

// ndjsonLineHeader is a permissive union that holds either a
// regular entry OR a sentinel line. We use it as a first-pass
// decoder so the StreamEntries loop has a single switch over
// `Done` / `Error` / regular — without committing to "which
// shape did we get?" prematurely.
//
// `Total` is interface{} (decoded into a float64 by the json
// package on numeric input) because the spec doesn't fully pin
// the type and a misconfigured server might send "42" as a
// string. parseTotal coerces both shapes.
type ndjsonLineHeader struct {
	Done  *bool       `json:"_done,omitempty"`
	Total interface{} `json:"total,omitempty"`
	Error string      `json:"_error,omitempty"`
	Code  string      `json:"code,omitempty"`
}

// parseTotal coerces ndjsonLineHeader.Total (the spec says
// integer, but json.Number is safer in case the server emits a
// string under future-proofing). Returns 0 on any non-numeric
// shape — the value is advisory (entry count for the user,
// nothing else branches on it).
func parseTotal(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case string:
		n, err := strconv.Atoi(t)
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

// buildEntriesURL stitches the query string for
// /api/archive/<node>/entries.
//
// Wire shape per spec §3:
//
//	GET /api/archive/<node>/entries?source=<archive-URI>
//
// The spec lists `source` as the only documented query
// parameter — the server infers the archive container format
// from the file extension server-side, so we deliberately do
// NOT forward our locally-derived Format hint. (We keep
// EntriesOptions.Format anyway: the cobra layer still uses it
// for local checks like "is this format passwordable?" before
// hitting the wire.)
//
// The parameter name and path-value encoding are driven by
// archiveQueryShape (see client.go) so the user can flip
// OLARES_ARCHIVE_PARAM_SOURCE / _PATH_ENCODE / _STRIP_LEADING_SLASH
// to test alternative wire formats without a recompile, while the
// spec doc's "Query 参数" table is being finalised.
func buildEntriesURL(c *Client, opts EntriesOptions) string {
	shape := currentArchiveQueryShape()
	q := shape.sourceParam + "=" + shape.encodePathValue(opts.Source)
	return c.archiveURL(opts.Node, "entries") + "?" + q
}

// validateEntriesOptions mirrors the spirit of
// validateCompressOptions. Format is the only "really required"
// field beyond Source / Node — but the cobra layer's default
// derivation from Source's extension means a missing format is
// uncommon in practice.
func validateEntriesOptions(opts EntriesOptions) error {
	if opts.Node == "" {
		return errors.New("entries: empty Node (cobra layer should resolve a default via /api/nodes/ or --node)")
	}
	if strings.TrimSpace(opts.Source) == "" {
		return errors.New("entries: source archive path is empty")
	}
	if opts.Format == "" {
		return errors.New("entries: --format is required (one of: " + JoinFormats() + ")")
	}
	if !IsSupportedFormat(opts.Format) {
		return fmt.Errorf("entries: unsupported --format %q; valid formats: %s",
			opts.Format, JoinFormats())
	}
	return nil
}

// IsEntriesStreamError reports whether `err` is the typed
// in-band stream error from StreamEntries. Convenience for the
// cobra layer's error reformatter.
func IsEntriesStreamError(err error) (*EntriesStreamError, bool) {
	var e *EntriesStreamError
	for cur := err; cur != nil; {
		if cast, ok := cur.(*EntriesStreamError); ok {
			e = cast
			return e, true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := cur.(unwrapper)
		if !ok {
			return nil, false
		}
		cur = u.Unwrap()
	}
	return nil, false
}
