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

// ProtectedDriveHomeChildren is the set of first-level child names
// directly under `drive/Home/` that the LarePass web app treats as
// system-managed and refuses to expose as rename / delete / cut /
// copy / paste targets. The CLI mirrors that policy on the verbs that
// MUTATE the source — `rename`, `rm`, and the `mv` source — so a
// scripted pipeline can't silently undo what the GUI prevents the
// user from doing by hand.
//
// Source of truth on the wire: the web app's
// apps/packages/app/src/stores/operation.ts `disableMenuItem` array
// (gated by `path === '/Files/Home/'` in `isDisableMenuItem`). The
// names below mirror that array verbatim — case-sensitive, including
// LarePass-specific quirks like `Huggingface` (one-word) and `Home`
// (a defensive entry for nested `Home/Home/` shapes the GUI also
// guards).
//
// Scope: the predicate fires ONLY when the path is exactly
// `drive/Home/<one-of-these>` with no deeper subpath — children of
// these dirs (e.g. `drive/Home/Pictures/Trip2024/`) are user content
// and stay fully writable, the same way the GUI's per-row menu only
// disables on the protected entry itself.
//
// `cp` (copy) is intentionally NOT gated by this list: copy preserves
// the source unchanged, so duplicating `drive/Home/Pictures/` to
// `drive/Home/PicturesBackup/` is a perfectly reasonable workflow
// that the LarePass GUI happens to not surface but the CLI sees no
// reason to forbid. The constraint is about preserving the protected
// directory itself, not about firewalling its bytes.
var ProtectedDriveHomeChildren = map[string]struct{}{
	"Home":        {},
	"Documents":   {},
	"Pictures":    {},
	"Movies":      {},
	"Downloads":   {},
	"Data":        {},
	"Cache":       {},
	"Code":        {},
	"Music":       {},
	"Ollama":      {},
	"Huggingface": {},
}

// protectedDriveHomeChildrenList is a stable, sorted, comma-joined
// rendering of ProtectedDriveHomeChildren for error messages and
// docstrings. Computed once so the (cold) refusal path doesn't
// allocate on every Plan call. Keep in sync with the map above.
var protectedDriveHomeChildrenList = func() string {
	out := make([]string, 0, len(ProtectedDriveHomeChildren))
	for k := range ProtectedDriveHomeChildren {
		out = append(out, k)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}()

// ProtectedDriveHomeChildrenList returns the canonical
// alphabetically-sorted comma-joined rendering of
// ProtectedDriveHomeChildren — useful for error messages that want
// to enumerate the protected names without re-sorting on each call.
func ProtectedDriveHomeChildrenList() string { return protectedDriveHomeChildrenList }

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

// IsExternalNodeRoot reports whether the path points at the
// `external/<node>/` layer with no subpath beyond it.
//
// On the wire, `external/<node>/` is NOT a real directory — the
// per-user files-backend exposes attached volumes (hdd1 / usb1 /
// smb-... mount points) as virtual children of this layer; see the
// LarePass web app's ExternalDataAPI.fetchDrive in
// apps/packages/app/src/api/files/v2/external/data.ts (when url=='/'
// it short-circuits via formatAppDataNode and returns the node /
// volume listing without ever calling /api/resources/external/). At
// `external/<node>/` itself the backend has no underlying filesystem
// to write into, so:
//
//   - POST /api/resources/external/<node>/<name>/ (mkdir) and the
//     follow-on chunk POST in upload either fail server-side or hit
//     the auto-rename quirk against a non-existent target;
//   - PATCH /api/paste/<node>/ with destination /external/<node>/...
//     (cp / mv into the volume-listing layer) likewise has nowhere
//     to land.
//
// Callers that perform writes (`mkdir`, `cp` / `mv` destination,
// `upload`) use this predicate to fail fast client-side with a
// self-describing error pointing at the `external/<node>/<volume>/<sub>`
// shape required for any real I/O. Reads (`ls`, `cat`, `rm`,
// `rename`, `share`) DO work at this layer (ls in particular is the
// way users discover what volumes are attached), so the writes-only
// scope mirrors the LarePass web app's own behavior — its sidebar
// renders the volume listing as a navigable tree but disables
// "create folder" / "paste here" / "upload here" actions while the
// user is at that level.
func (p FrontendPath) IsExternalNodeRoot() bool {
	return p.FileType == "external" && strings.Trim(p.SubPath, "/") == ""
}

// IsCacheNodeRoot reports whether the path points exactly at the
// `cache/<node>/` layer with no subpath beyond it.
//
// Unlike `external/<node>/` (which is a fully virtual volume-listing
// layer, see IsExternalNodeRoot), `cache/<node>/` IS backed by a real
// per-node filesystem on the wire — `ls` / `cat` / `cp` / `upload` all
// work fine against it. The reason this predicate exists is narrower:
//
//   - In the LarePass web app, the Cache namespace's root view at
//     `/Cache/` is rendered by [`CacheDataAPI.fetchCache`](apps/packages/app/src/api/files/v2/cache/data.ts)
//     which short-circuits via `formatAppDataNode` and synthesizes
//     children from `filesStore.nodes` (the Olares cluster's node
//     list) instead of hitting `/api/resources/cache/...`. So when
//     the user is sitting at `/Files/Cache/` they're picking a node,
//     not a directory — and "share this node" is not a meaningful
//     operation: a share at that level points at no concrete dataset,
//     just an empty container the recipient has no way to populate.
//
//   - Once the user picks a node and navigates into `cache/<node>/`,
//     the wire goes back to the regular `/api/resources/cache/<node>/`
//     directory listing, and shares of `cache/<node>/<sub>/` are real
//     and useful.
//
// Used by `frontendPathToShareTarget` to fail fast on
// `share internal|public|smb cache/<node>/`, mirroring the LarePass
// web app's UX (the `/Files/Cache/` view shows a node picker, not a
// row the user can right-click → Share). Other verbs (`ls`, `cp`,
// `mkdir`, ...) DO work at this layer and are unaffected.
func (p FrontendPath) IsCacheNodeRoot() bool {
	return p.FileType == "cache" && strings.Trim(p.SubPath, "/") == ""
}

// IsProtectedDriveHomeChild reports whether this path points exactly
// at one of the system-managed first-level children directly under
// `drive/Home/` (see ProtectedDriveHomeChildren for the list and the
// rationale).
//
// The predicate is intentionally narrow:
//
//   - FileType MUST be "drive" and Extend MUST be "Home" — the policy
//     applies only to the LarePass-managed Home volume, not to Data
//     or any other namespace.
//   - SubPath MUST be exactly one segment (e.g. "/Pictures" or
//     "/Pictures/", with or without a directory marker). Deeper
//     paths like "/Pictures/Trip2024/" are user content and are NOT
//     protected — the same scope the GUI uses by gating on the
//     selected row's name only when the user is at /Files/Home/.
//   - The single segment must match a name in
//     ProtectedDriveHomeChildren case-sensitively. The GUI compares
//     enum string values, so we do too (so e.g. `pictures` lowercase
//     won't trip the predicate, but it would also not be a real
//     directory under Home — these names are the system-created
//     bootstrap dirs and their casing is fixed).
//
// Used by `rename`, `rm`, and the `mv` source-side check to refuse
// the operation client-side with a self-describing error. `cp` does
// NOT use this — see the docstring on ProtectedDriveHomeChildren for
// why duplicating the bytes is fine even when renaming/deleting the
// original is not.
func (p FrontendPath) IsProtectedDriveHomeChild() bool {
	if p.FileType != "drive" || p.Extend != "Home" {
		return false
	}
	seg := strings.Trim(p.SubPath, "/")
	if seg == "" || strings.Contains(seg, "/") {
		return false
	}
	_, ok := ProtectedDriveHomeChildren[seg]
	return ok
}
