// Package files implements the `olares-cli files ...` command tree, which
// talks to the per-user files-backend (the upstream `files` repo) over its
// /api/resources REST surface.
//
// The files-backend models every resource as a 3-segment "front-end path":
//
//	<fileType>/<extend>/<subPath>
//
// where `fileType` selects the storage class, `extend` selects the concrete
// volume / repo / account inside that class, and `subPath` is the relative
// path inside that volume. See files/pkg/models/file_param.go (FileParam.convert)
// and files/pkg/common/constant.go for the canonical definitions.
//
// We expose the full path verbatim on the CLI surface — the user always
// types all three segments — so the tooling stays close to the protocol.
// path.go centralizes parsing & validation; commands like `ls`, `cat`, `cp`
// (the latter two land in Phase 2) all consume the resulting FrontendPath.
package files

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Known fileType values understood by the files-backend.
// Mirrors files/pkg/common/constant.go (Drive, Cache, Sync, External, AwsS3,
// GoogleDrive ("google"), DropBox, TencentCos ("tencent"), Share, Internal).
// The list is intentionally case-sensitive lowercase: the backend lowercases
// the first segment before matching, so we accept only the canonical form on
// input to avoid surprising behavior.
var knownFileTypes = map[string]struct{}{
	"drive":    {},
	"cache":    {},
	"sync":     {},
	"external": {},
	"awss3":    {},
	"dropbox":  {},
	"google":   {},
	"tencent":  {},
	"share":    {},
	"internal": {},
}

// driveExtends enumerates the only valid `extend` values when fileType=="drive".
// Backend enforcement: files/pkg/models/file_param.go convert() L59-61.
var driveExtends = map[string]struct{}{
	"Home": {},
	"Data": {},
}

// knownFileTypesList is a stable, sorted, comma-joined rendering of
// knownFileTypes, computed once so the (cold) error path doesn't allocate
// on every parse failure. Keep in sync with knownFileTypes above.
var knownFileTypesList = func() string {
	out := make([]string, 0, len(knownFileTypes))
	for k := range knownFileTypes {
		out = append(out, k)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}()

// FrontendPath is the parsed view of a 3-segment files-backend front-end path.
// Construct via ParseFrontendPath; the zero value has no meaning.
type FrontendPath struct {
	// FileType is the (always-lowercase) storage class: drive/cache/sync/...
	FileType string
	// Extend is the volume / repo / account selector. Its semantics depend on
	// FileType (Home|Data for drive, node name for cache/external, repo_id for
	// sync, account key for cloud drives, ...). CLI-side we only hard-validate
	// the drive case; everything else is left to the backend.
	Extend string
	// SubPath is the path inside Extend, always starting with '/'. Root is "/".
	// A trailing slash present in the input is preserved (the backend uses it
	// as a "this is a directory" hint in some places). It is also synthesized
	// for `<fileType>/<extend>` (no subpath), where the only valid backend
	// interpretation is "the extend root", because the backend's
	// FileParam.convert() splits on '/' and rejects len < 3 for non-root
	// resources.
	SubPath string
}

// ParseFrontendPath parses a user-supplied path string into a FrontendPath.
//
// Examples:
//
//	"drive/Home/"                    → {drive, Home, "/"}
//	"drive/Home"                     → {drive, Home, "/"}    (root synthesized)
//	"drive/Home/Documents"           → {drive, Home, "/Documents"}
//	"drive/Home/Documents/"          → {drive, Home, "/Documents/"}
//	"sync/<repo_id>/sub/dir"         → {sync, <repo_id>, "/sub/dir"}
//	"awss3/<account>/<bucket>/k.txt" → {awss3, <account>, "/<bucket>/k.txt"}
//
// Validation:
//   - Path must have at least 2 non-empty segments (fileType + extend).
//     `drive/Home` (no trailing slash, no subpath) is accepted and treated as
//     the extend root, because the backend's FileParam.convert() splits on
//     '/' and rejects len < 3 — there is no valid reading of a bare extend
//     other than "the root directory".
//   - FileType must be a known value (case-sensitive lowercase). Unknown
//     values fail fast on the client to avoid an opaque 500 from the server.
//   - When FileType=="drive", Extend must be "Home" or "Data" (case-sensitive).
//   - Other FileTypes' Extend values (node names, repo ids, account keys) are
//     not pre-validated locally; the backend is the source of truth for those.
//   - Path traversal segments like ".." are NOT stripped here — the backend
//     applies its own sandboxing. We do collapse runs of "//" to a single "/"
//     via path.Clean while preserving any user-supplied trailing slash.
func ParseFrontendPath(raw string) (FrontendPath, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return FrontendPath{}, fmt.Errorf("front-end path is empty; expected <fileType>/<extend>[/<subPath>] (e.g. drive/Home/, sync/<repo_id>/)")
	}

	hadTrailingSlash := strings.HasSuffix(raw, "/")
	trimmed := strings.Trim(raw, "/")
	parts := strings.Split(trimmed, "/")
	// strings.Split never returns empty slice; guard against the all-empty
	// case for defense in depth.
	if len(parts) == 0 || parts[0] == "" {
		return FrontendPath{}, fmt.Errorf("front-end path %q is empty after trimming slashes", raw)
	}
	if len(parts) < 2 {
		return FrontendPath{}, fmt.Errorf("front-end path %q must have <fileType>/<extend>[/<subPath>] (got only %d segment(s); try e.g. %q)",
			raw, len(parts), parts[0]+"/<extend>/")
	}

	fileType := parts[0]
	if _, ok := knownFileTypes[fileType]; !ok {
		return FrontendPath{}, fmt.Errorf("unknown fileType %q, expected one of: %s",
			fileType, knownFileTypesList)
	}

	extend := parts[1]
	if extend == "" {
		return FrontendPath{}, fmt.Errorf("front-end path %q has empty <extend> segment", raw)
	}
	if fileType == "drive" {
		if _, ok := driveExtends[extend]; !ok {
			return FrontendPath{}, fmt.Errorf("drive extend must be Home or Data (got %q)", extend)
		}
	}

	sub := "/"
	if len(parts) > 2 {
		// path.Clean collapses "//", strips trailing "/" — restore the latter
		// from the caller's intent below.
		sub = path.Clean("/" + strings.Join(parts[2:], "/"))
	}
	if hadTrailingSlash && !strings.HasSuffix(sub, "/") {
		sub += "/"
	}

	return FrontendPath{
		FileType: fileType,
		Extend:   extend,
		SubPath:  sub,
	}, nil
}

// String renders the canonical front-end path as `<fileType>/<extend><subPath>`.
// SubPath always starts with '/', so the output naturally looks like
// "drive/Home/Documents" or "drive/Home/" for the root with a trailing slash.
//
// This is the human-readable form, suitable for error messages and logs.
// For URL construction use URLPath() — String() does NOT percent-encode.
func (p FrontendPath) String() string {
	return p.FileType + "/" + p.Extend + p.SubPath
}

// URLPath returns the same logical path as String() but percent-encoded
// with internal/files/encodepath.EncodeURL — the Go counterpart of the web app's
// apps/packages/app/src/utils/encode.ts `encodeUrl` (encodeURIComponent per
// '/' segment). This MUST stay aligned with download/cat/rm/upload, which
// all use internal/files/encodepath; url.PathEscape is not equivalent (e.g. '+' and
// '!*'()' differ) and would make `ls` hit different wire paths than the
// other verbs for the same user-typed path.
func (p FrontendPath) URLPath() string {
	return encodepath.EncodeURL(p.String())
}

// HasTrailingSlash reports whether String() ends with '/' — i.e. whether the
// path represents a directory. Useful for callers that want to disambiguate
// "list this directory" from "fetch this resource by exact name" without
// re-parsing. Derived from SubPath so it always agrees with String(); in
// particular, `<fileType>/<extend>` (no subpath, no trailing slash on input)
// reports true because SubPath is "/" — the only valid interpretation of a
// bare extend reference is its root directory.
func (p FrontendPath) HasTrailingSlash() bool {
	return strings.HasSuffix(p.SubPath, "/")
}
