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
	"math"
	"regexp"
	"sort"
	"strconv"
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

// singleFileCompressionFormats is the subset of SupportedFormats
// that are raw single-stream compressors: they wrap exactly ONE
// file's bytes with no container/index, so they cannot carry a
// directory tree or multiple members.
//
// This mirrors the LarePass web app's `singleFileCompressionFormats`
// (TermiPass utils/interface/archive.ts) and its
// `canCompressItemsWithFormat` gate: gzip / bzip2 / xz may only
// compress a SINGLE, non-directory source. Feeding them a directory
// or multiple sources is a user error the CLI rejects up front
// instead of letting the backend produce a surprising single-file
// archive (or fail opaquely).
//
// tar.gz / tar.bz2 / tar.xz / tgz are NOT here: those tar FIRST
// (a real multi-member container) and THEN compress the tarball,
// so they happily carry directories and multiple sources.
var singleFileCompressionFormats = map[string]struct{}{
	"gzip":  {},
	"bzip2": {},
	"xz":    {},
}

// previewUnsupportedFormats is the subset that the archive
// preview endpoints (`/entries`, `/entry`) cannot walk. They are
// the bare single-stream compressors whose payload is an opaque
// byte run with no listable entry table — there is literally
// nothing to enumerate, so `files archive entries` / `cat` would
// only ever surface an empty or misleading listing.
//
// This mirrors the LarePass web app's
// `unsupportedArchivePreviewExtensions` (= bz2 / bzip2 / xz):
//
//   - bzip2 (.bz2 / .bzip2) and xz (.xz) → no entry table → no preview.
//   - gzip (.gz / .gzip) is INTENTIONALLY absent: the web app lists
//     gzip as previewable, so we mirror that and allow it.
//   - The tar.* compound formats (tar.gz / tar.bz2 / tar.xz / tgz)
//     are real tar containers and remain fully previewable; only
//     the BARE bzip2 / xz formats are gated.
//
// Extract (`files extract`) is NOT gated by this set — unpacking a
// bare bzip2 / xz stream into its single decompressed file is a
// legitimate, supported operation; only the entry-listing preview
// is meaningless for them.
var previewUnsupportedFormats = map[string]struct{}{
	"bzip2": {},
	"xz":    {},
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

// IsSingleFileCompressionFormat reports whether `format` is a raw
// single-stream compressor (gzip / bzip2 / xz) that can only wrap
// ONE non-directory file. See singleFileCompressionFormats for the
// LarePass parity rationale.
func IsSingleFileCompressionFormat(format string) bool {
	_, ok := singleFileCompressionFormats[strings.ToLower(format)]
	return ok
}

// SupportsPreview reports whether `format` can be walked by the
// archive preview endpoints (`files archive entries` / `cat`).
// Returns false for the bare bzip2 / xz single-stream compressors,
// which have no listable entry table. Mirrors the LarePass web
// app's unsupportedArchivePreviewExtensions gate.
func SupportsPreview(format string) bool {
	_, ok := previewUnsupportedFormats[strings.ToLower(format)]
	return !ok
}

// ValidateSingleFileCompression enforces the gzip / bzip2 / xz
// "single non-directory source" rule for `compress`. It mirrors
// the LarePass web app's canCompressItemsWithFormat gate so the
// CLI and the web UI reject the exact same inputs.
//
// `sourceCount` is the number of compress sources; `anyDir`
// reports whether ANY of them is (or is declared as) a directory.
// For non-single-file formats this is always nil — they happily
// carry directories and multiple members.
//
// Returns a typed, actionable error naming the format and pointing
// the user at a container format (zip / 7z / tar.gz) when they tried
// to pack a directory or multiple files into a single-stream codec.
func ValidateSingleFileCompression(format string, sourceCount int, anyDir bool) error {
	if !IsSingleFileCompressionFormat(format) {
		return nil
	}
	if sourceCount > 1 {
		return fmt.Errorf(
			"format %q compresses a single file only; got %d sources. "+
				"Use a container format (zip, 7z, tar, tar.gz, tar.bz2, tar.xz, tgz) to pack multiple entries",
			format, sourceCount)
	}
	if anyDir {
		return fmt.Errorf(
			"format %q compresses a single file only and cannot pack a directory. "+
				"Use a container format (zip, 7z, tar, tar.gz, tar.bz2, tar.xz, tgz) to pack a directory",
			format)
	}
	return nil
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

// ParseVolumeSize parses a split-volume size string with an
// optional unit suffix into whole MiB — the unit the compress
// endpoint's `volumeSizeMB` field expects. The server accepts
// integer MiB only, so this parser normalizes the user input to
// an integer MiB count.
//
//	"100"     → 100   (bare number = MiB)
//	"100MB"   → 100
//	"1.5GB"   → 1536  (1.5 * 1024, rounded up)
//	"2G"      → 2048
//
// Accepted suffixes (case-insensitive): M / MB, G / GB.
// A bare number is treated as MiB. The result is rounded UP to the
// nearest MiB and floored at 1, matching the web app's
// Math.max(1, ceil(mbValue)) — the backend rejects a 0-size
// volume, and rounding up keeps fractional GiB inputs from silently
// shrinking the split size.
//
// Returns an error for empty input, a non-numeric value, an
// unknown suffix, or a non-positive size, so a typo surfaces as a
// clean client-side message instead of an opaque backend 400.
func ParseVolumeSize(raw string) (int, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("empty volume size")
	}
	// Split the numeric head from the (optional) unit tail. We
	// scan from the end so "1.5gb" / "100 mb" / "100" all work.
	lower := strings.ToLower(s)
	lower = strings.TrimSpace(lower)
	var numPart, unit string
	switch {
	case strings.HasSuffix(lower, "gb"):
		numPart, unit = lower[:len(lower)-2], "g"
	case strings.HasSuffix(lower, "mb"):
		numPart, unit = lower[:len(lower)-2], "m"
	case strings.HasSuffix(lower, "g"):
		numPart, unit = lower[:len(lower)-1], "g"
	case strings.HasSuffix(lower, "m"):
		numPart, unit = lower[:len(lower)-1], "m"
	default:
		// Reject other alpha suffixes early (e.g. KB/TB) so the
		// error points at unsupported units explicitly.
		if strings.HasSuffix(lower, "b") || strings.HasSuffix(lower, "k") || strings.HasSuffix(lower, "t") {
			return 0, fmt.Errorf("invalid volume size %q: unsupported unit (only MB and GB are supported)", raw)
		}
		numPart, unit = lower, "m" // bare number = MiB
	}
	numPart = strings.TrimSpace(numPart)
	value, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid volume size %q: %q is not a number (use forms like 100MB, 1.5GB, or a bare MiB count)", raw, numPart)
	}
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("invalid volume size %q: must be a positive number", raw)
	}
	var mb float64
	switch unit {
	case "g":
		mb = value * 1024
	default: // "m"
		mb = value
	}
	// Round up, floor at 1 — mirrors normalizeSplitVolume.
	mib := int(math.Ceil(mb))
	if mib < 1 {
		mib = 1
	}
	return mib, nil
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
	// Multi-volume main-part naming used by backend/archive tools:
	//   *.zip.001 / *.zip.002 / ...
	//   *.7z.001  / *.7z.002  / ...
	// The user may pass the first (main) part directly; treat it as
	// zip / 7z for extract and preview format inference.
	if m := splitVolumeMainRe.FindStringSubmatch(lower); len(m) == 2 {
		return m[1]
	}
	// Longest-suffix-first. The order here is the contract the
	// cobra-layer help text quotes; do NOT alphabetise.
	for _, c := range []struct {
		ext    string
		format string
	}{
		{".tar.gz", "tar.gz"},
		{".tar.bz2", "tar.bz2"},
		{".tar.xz", "tar.xz"},
		{".bzip2", "bzip2"},
		{".gzip", "gzip"},
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

var splitVolumeMainRe = regexp.MustCompile(`\.(zip|7z)\.\d+$`)
