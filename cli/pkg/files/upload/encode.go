// Package upload implements the chunked / resumable upload client that the
// `olares-cli files upload` command drives. It speaks the same wire protocol
// as the LarePass web app (Resumable.js + the Drive v2 endpoints under
// /upload/upload-link, /upload/file-uploaded-bytes, /api/nodes,
// /api/resources/drive/Home/...). See docs/notes/auth-2fa-semantics.md for
// the auth header convention shared with the rest of the CLI, and the plan
// at .cursor/plans/cli_files_upload_resumable_*.plan.md for the design
// rationale.
//
// encode.go: percent-encoding helpers that mirror the web app's
// apps/packages/app/src/utils/encode.ts (encodeUrl). The standard library
// alone is NOT a 1:1 substitute:
//
//   - url.QueryEscape encodes a space as '+' (form encoding) — JS
//     encodeURIComponent uses '%20'.
//   - url.QueryEscape escapes '!' '*' '(' ')' '\'' — JS encodeURIComponent
//     does not.
//
// Both differences would round-trip to the server differently and break
// resume / probe path-matching for filenames containing those bytes, so we
// implement encodeURIComponent ourselves and use it everywhere we touch a
// path / filename / parent_dir value.
package upload

import (
	"strings"
)

// encodeURIComponent mirrors JavaScript's encodeURIComponent: it leaves
// the unreserved set (RFC 3986) plus !*'() alone and percent-encodes the
// rest as UTF-8 bytes. This is the building block for both EncodeURL
// (path-segment encoding, joined with '/') and the query-value encoding
// the upload protocol uses for parent_dir / file_name.
//
// Why we don't reuse net/url:
//   - url.QueryEscape encodes ' ' as '+' (form encoding). The Drive
//     backend was written against a JS client that emits '%20', so the
//     two representations are not interchangeable for filename parity
//     (probe and chunk POST must see byte-identical names for resume to
//     line up).
//   - url.PathEscape leaves '?', '#', '&', '=' alone (they're valid
//     within a path component) but those characters DO need escaping when
//     the value is destined for a query parameter, which is the bulk of
//     our use case.
func encodeURIComponent(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldNotEncode(c) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(upperHex[c>>4])
		b.WriteByte(upperHex[c&0x0f])
	}
	return b.String()
}

// EncodeURL is the Go counterpart of apps/packages/app/src/utils/encode.ts
// `encodeUrl`: split on '/', encodeURIComponent each segment, rejoin with
// '/'. Used wherever a path is interpolated into a URL path component
// (e.g. the file_path query value the server uses to derive the upload
// link, or the /api/resources/drive/Home/... mkdir endpoint).
//
// The leading and trailing '/' are preserved verbatim so callers can keep
// the "directory hint" semantics the backend relies on (a trailing '/'
// signals "this is a directory" in several files-backend code paths; see
// files/pkg/models/file_param.go).
func EncodeURL(p string) string {
	if p == "" {
		return ""
	}
	// split keeps empty leading/trailing parts so the leading/trailing
	// slashes survive the rejoin (e.g. "/a/b/" → ["", "a", "b", ""] →
	// "/a/b/" again after encoding the non-empty pieces).
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if seg == "" {
			continue
		}
		parts[i] = encodeURIComponent(seg)
	}
	return strings.Join(parts, "/")
}

const upperHex = "0123456789ABCDEF"

// shouldNotEncode is the membership test for JS encodeURIComponent's
// "leave alone" set: A-Z a-z 0-9 plus the marks `- _ . ~ ! * ' ( )`.
func shouldNotEncode(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	}
	switch c {
	case '-', '_', '.', '~', '!', '*', '\'', '(', ')':
		return true
	}
	return false
}
