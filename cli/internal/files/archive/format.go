// format.go: canonical format / conflict / level validation for
// the archive surface. Lifted into its own file so the cobra
// layer's pre-flight (and the JSON marshalling in compress.go /
// extract.go) can call into one canonical contract instead of
// duplicating the matrix.
//
// Per the backend's archive endpoint contract (the spec the user
// supplied — mirrored at the top of compress.go):
//
//	compress writable + extract / preview readable formats:
//	    zip, 7z, tar, tar.gz, tgz, tar.bz2, tar.xz, gzip, bzip2, xz
//	passwordable formats:                  zip, 7z
//	multi-volume / split-archive formats:  zip, 7z
//	conflict policies:                     rename | overwrite | skip
//	compress level domain:                 0..9  (0 = no compression / store)
//
// The CLI side validates these client-side so a typo (e.g.
// --format=tgz vs --format=tar.gz) surfaces as a clean error
// instead of an opaque 400. Same fail-fast spirit as the cp
// preflight: catch the problem before the wire call goes out.
package archive

import (
	"fmt"
	"sort"
	"strings"
)

// SupportedFormats lists every archive container the backend
// understands for BOTH compress and extract/entries/entry. We
// model them as plain strings rather than a typed enum because the
// wire shape IS the string ("tar.gz" not TarGz) — a typed enum
// would force a string-conversion round-trip on every marshal.
//
// Keep alphabetised so the error message that lists "valid
// formats" reads naturally. Add an entry here when (and only
// when) the backend gains support — the CLI's job is to mirror,
// not anticipate.
var SupportedFormats = []string{
	"7z",
	"bzip2",
	"gzip",
	"tar",
	"tar.bz2",
	"tar.gz",
	"tar.xz",
	"tgz",
	"xz",
	"zip",
}

// passwordableFormats is the subset of SupportedFormats that
// accept an `X-Archive-Password` header. The backend silently
// ignores the header on other formats, but the CLI rejects the
// mismatch so a user who typed `--password-stdin` with
// `--format=tar` gets a clear "tar does not support passwords"
// up front (otherwise the request "succeeds" without ever
// encrypting anything — a quiet footgun).
var passwordableFormats = map[string]struct{}{
	"zip": {},
	"7z":  {},
}

// multiVolumeFormats is the subset that supports the
// `volumeSizeMB` knob on compress. Same rationale as the
// password gate: the backend ignores volumeSizeMB on other
// formats, and "I asked for 100 MB volumes but got a single
// 4 GB tar.gz" is the kind of silent surprise a CLI should
// refuse up front.
var multiVolumeFormats = map[string]struct{}{
	"zip": {},
	"7z":  {},
}

// MinLevel / MaxLevel bound the `level` parameter on compress.
// 0 = store (no compression), 9 = maximum (slowest). The exact
// per-codec semantics live in the backend's 7zz / gzip / zstd /
// xz wrappers; we only enforce the outer bounds so a typo like
// `--level=99` doesn't survive into the JSON body.
const (
	MinLevel = 0
	MaxLevel = 9
)

// Conflict is the on-collision policy the backend's
// archive-extract / archive-compress writer applies when a target
// path already exists at the destination.
//
// Wire shape is the lowercase string itself; we keep the enum
// typed so the cobra layer can declare the flag default with no
// risk of typo. ConflictDefault matches the backend's documented
// default (rename) so passing an empty string here is harmless —
// the validator coerces "" to ConflictRename before sending.
type Conflict string

const (
	ConflictRename    Conflict = "rename"
	ConflictOverwrite Conflict = "overwrite"
	ConflictSkip      Conflict = "skip"
	ConflictDefault            = ConflictRename
)

// validConflicts lists the on-wire values the backend accepts.
// Same pattern as SupportedFormats — kept as a slice so error
// messages can enumerate the valid set without re-sorting on
// every parse failure.
var validConflicts = []Conflict{
	ConflictRename,
	ConflictOverwrite,
	ConflictSkip,
}

// supportedFormatsSet derives from SupportedFormats once at init
// time so the validator's hot path is a single map lookup. The
// list shape is preserved for error-message enumeration; the map
// is the predicate.
var supportedFormatsSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(SupportedFormats))
	for _, f := range SupportedFormats {
		m[f] = struct{}{}
	}
	return m
}()

// IsSupportedFormat reports whether `format` (lowercase) is in
// the canonical set the backend understands. The cobra layer
// rejects unknown formats with a "valid formats: ..." enumeration
// built from SupportedFormats.
func IsSupportedFormat(format string) bool {
	_, ok := supportedFormatsSet[strings.ToLower(format)]
	return ok
}

// SupportsPassword reports whether `format` accepts an
// X-Archive-Password header. zip / 7z only — see the
// passwordableFormats docstring for the why.
func SupportsPassword(format string) bool {
	_, ok := passwordableFormats[strings.ToLower(format)]
	return ok
}

// SupportsMultiVolume reports whether `format` accepts the
// `volumeSizeMB` knob on compress. zip / 7z only.
func SupportsMultiVolume(format string) bool {
	_, ok := multiVolumeFormats[strings.ToLower(format)]
	return ok
}

// ParseConflict coerces a user-supplied --conflict flag into a
// Conflict value. Empty → ConflictDefault (rename) — that's the
// backend's documented default and matches the LarePass web
// app's UX (every drop-in defaults to a non-destructive rename).
//
// Returns a typed error naming the offending value AND the valid
// set so a typo like `--conflict=overwite` is immediately
// actionable.
func ParseConflict(raw string) (Conflict, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return ConflictDefault, nil
	}
	for _, c := range validConflicts {
		if string(c) == s {
			return c, nil
		}
	}
	return "", fmt.Errorf("invalid --conflict %q; valid values: %s", raw, joinConflicts(validConflicts))
}

// ValidateLevel checks the compress-level range. 0 = store
// (no compression), 9 = max (slowest). The MaxLevel cap is
// inclusive because that's the upper bound every backend codec
// understands; level=10 surfaces as the server's opaque 400.
//
// Sentinel value -1 means "use the backend's default" — the
// cobra layer maps the unset flag (DefaultLevel) to -1 before
// calling, so the JSON body omits the field entirely. We return
// nil for the sentinel so callers don't have to special-case it.
func ValidateLevel(level int) error {
	if level < 0 {
		// Sentinel for "unset" — handled by omitempty in the
		// request struct. No further validation needed.
		return nil
	}
	if level > MaxLevel {
		return fmt.Errorf("invalid --level %d; valid range: %d..%d", level, MinLevel, MaxLevel)
	}
	return nil
}

// ValidateFormat is the canonical entry point for the cobra
// layer. Single string in, typed error out — same shape as
// share.ParsePermission and cp.Plan's per-source validators.
//
// `usage` is a short verb name (e.g. "compress" / "extract")
// included in the error so a multi-verb cobra cmd doesn't have
// to wrap the error itself.
func ValidateFormat(format, usage string) error {
	if format == "" {
		return fmt.Errorf("%s: --format is required (one of: %s)", usage, JoinFormats())
	}
	if !IsSupportedFormat(format) {
		return fmt.Errorf("%s: unsupported --format %q; valid formats: %s",
			usage, format, JoinFormats())
	}
	return nil
}

// JoinFormats renders the canonical alphabetised comma-joined
// list of SupportedFormats. Used by error messages so the
// "valid formats: ..." suffix stays stable across callers.
func JoinFormats() string {
	out := make([]string, len(SupportedFormats))
	copy(out, SupportedFormats)
	sort.Strings(out)
	return strings.Join(out, ", ")
}

// joinConflicts renders the canonical comma-joined list of
// validConflicts in their natural order (rename first because
// it's the default — the others are stable but secondary).
func joinConflicts(cs []Conflict) string {
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = string(c)
	}
	return strings.Join(parts, ", ")
}

// FormatFromExtension is a best-effort "guess the archive format
// from a filename" helper for the cobra layer's `--format` default.
// The user's destination basename (compress) or source basename
// (extract / entries / entry) is fed in; the helper returns the
// matching SupportedFormats entry or "" when nothing fits.
//
// Order matters: longer suffixes (`.tar.gz`, `.tar.bz2`, `.tar.xz`)
// must be checked BEFORE the shorter ones (`.gz`, `.bz2`, `.xz`)
// because every `.tar.gz` also matches `.gz`. Map iteration is
// unordered in Go, so we keep an explicit slice with the longest-
// first ordering baked in.
//
// We return "" (not an error) for unknown extensions because the
// caller's pattern is "fall through to require --format" rather
// than "this is fatal" — the cobra layer surfaces the missing
// format with its own contextualised message.
func FormatFromExtension(name string) string {
	lower := strings.ToLower(name)
	// Longest-suffix-first. The order here is the contract the
	// cobra-layer help text quotes; do NOT alphabetise.
	for _, c := range []struct {
		ext    string
		format string
	}{
		{".tar.gz", "tar.gz"},
		{".tar.bz2", "tar.bz2"},
		{".tar.xz", "tar.xz"},
		{".tgz", "tgz"},
		{".7z", "7z"},
		{".tar", "tar"},
		{".zip", "zip"},
		{".gz", "gzip"},
		{".bz2", "bzip2"},
		{".xz", "xz"},
	} {
		if strings.HasSuffix(lower, c.ext) {
			return c.format
		}
	}
	return ""
}
