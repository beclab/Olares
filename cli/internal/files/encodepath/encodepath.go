// Package encodepath implements percent-encoding for Olares files-backend
// wire paths so the CLI matches the LarePass web client
// (apps/packages/app/src/utils/encode.ts: encodeUrl / encodeURIComponent).
//
// This is shared by upload, download, rm, and the files Cobra commands — it
// does not belong in package upload alone: net/url helpers are not
// byte-identical to JS (spaces as '+', !*'() escaping, etc.), and resume /
// probes require the same encoding everywhere.
package encodepath

import "strings"

// EncodeURIComponent mirrors JavaScript's encodeURIComponent: it leaves the
// unreserved set (RFC 3986) plus !*'() alone and percent-encodes the rest as
// UTF-8 bytes. Use for query values (e.g. file_path=) and header fragments
// where the server expects JS-shaped bytes.
//
// Why we don't reuse net/url:
//   - url.QueryEscape encodes a space as '+' (form encoding) — JS uses '%20'.
//   - url.QueryEscape escapes '!' '*' '(' ')' '\” — JS does not.
func EncodeURIComponent(s string) string {
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

// EncodeURL is the Go counterpart of encode.ts `encodeUrl`: split on '/',
// EncodeURIComponent each segment, rejoin with '/'. Use for path segments in
// /api/resources/..., /api/raw/..., and upload link paths.
//
// Leading and trailing '/' are preserved so directory-hint semantics survive.
func EncodeURL(p string) string {
	if p == "" {
		return ""
	}
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if seg == "" {
			continue
		}
		parts[i] = EncodeURIComponent(seg)
	}
	return strings.Join(parts, "/")
}

const upperHex = "0123456789ABCDEF"

// shouldNotEncode matches JS encodeURIComponent's leave-alone set.
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
