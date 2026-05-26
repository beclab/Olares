// Package rm implements the wire side of `olares-cli files rm`. It
// drives the per-user files-backend's batch DELETE endpoint:
//
//	DELETE /api/resources/<encParentDir>/   body: {"dirents": [...]}
//
// The endpoint takes one parent directory in the URL and removes the
// listed entries inside it. Each dirent is a leading-slash name like
// `/foo` (file) or `/sub/` (directory); see
// files/pkg/drivers/posix/posix/posix.go's PosixStorage.Delete for
// the server's iteration logic.
//
// We mirror the LarePass web app's
// apps/packages/app/src/api/files/v2/common/utils.ts
// `batchDeleteFileItems` helper, which groups items by parent so the
// shape stays "fewest possible HTTP requests" — handy when removing,
// say, 200 files scattered across two directories.
package rm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Client is the per-FilesURL handle used by DeleteBatch.
//
// HTTPClient is expected to be the factory-provided client whose
// refreshingTransport injects `X-Authorization` (not `Authorization:
// Bearer`, see pkg/cmdutil/factory.go for why) and transparently refreshes
// on 401/403.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string // FilesURL, e.g. https://files.alice.olares.com
}

// HTTPError carries the status + truncated body of a non-2xx response
// so the caller can branch on the status code (e.g. to give a friendly
// "not found" message vs. an auth-issue CTA). Same shape as the other
// per-package HTTP errors in this CLI to keep the error contract
// uniform.
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

// Target is one user-supplied remote path to delete, normalized so the
// planner has a single canonical shape to group on. The cobra cmd
// constructs these from FrontendPath values; see ToTarget in cmd
// layer for the conversion (kept out of this package so the planner
// stays free of the FrontendPath type).
type Target struct {
	// FileType + Extend together identify the storage class+volume
	// (drive/Home, sync/<repo>, ...). Two targets with the same
	// (FileType, Extend, ParentSubPath) tuple share a parent and can
	// be batched into a single DELETE.
	FileType string
	Extend   string
	// ParentSubPath is the parent directory's path relative to
	// `<FileType>/<Extend>` — always starts with '/' and ends with
	// '/' (or is just "/" for items directly under Extend). This is
	// the value that the `/api/resources/<parent>/` URL is built
	// from. Keeping the trailing slash explicit avoids "/Home" vs
	// "/Home/" ambiguity.
	ParentSubPath string
	// Name is the basename of the entry to remove (no slashes).
	Name string
	// IsDirIntent: did the user signal this is a directory (e.g. by
	// passing a trailing slash on the path)? Required to be true for
	// directory removals when --recursive is set; the planner errors
	// out for IsDirIntent=true without --recursive (Unix-style).
	IsDirIntent bool
}

// Group is one batch DELETE: a parent path and the list of dirents to
// remove from it. The wire shape comes straight from the LarePass web
// app's batchDelete helper.
type Group struct {
	// FileType / Extend / ParentSubPath: same meaning as on Target,
	// shared by every dirent in the group.
	FileType      string
	Extend        string
	ParentSubPath string
	// Dirents is the list of `/<name>` (file) or `/<name>/` (dir)
	// strings to send in the request body. Sorted alphabetically so
	// the wire request is deterministic for tests / replay.
	Dirents []string
}

// Plan validates `--recursive` against each Target's IsDirIntent flag
// and groups the targets by parent directory. The returned []*Group
// is sorted by (FileType, Extend, ParentSubPath) so callers iterate
// in a stable order; within each group dirents are also sorted.
//
// Dirent shape on the wire:
//
//   - File deletion (no -r): dirent is `/<name>` (no trailing slash).
//     The server's POSIX driver routes this through the file-removal
//     path; sending it against a directory is a noisy server-side
//     error that's hard to recover from in scripts.
//   - Directory deletion (-r in scope): dirent is `/<name>/` (WITH a
//     trailing slash). `recursive=true` is taken as the user's
//     declaration of intent — every target in the call is treated
//     as a directory regardless of whether the original path string
//     had a trailing slash. This matches Unix `rm -r foo` semantics:
//     once `-r` is in play, the user has accepted that they want
//     directory-recursive removal, and forcing them to also remember
//     to type the trailing `/` is needless ceremony (and was the
//     source of "I added -r, why didn't it delete the folder?"
//     reports against an earlier revision of this planner).
//
// Errors:
//   - any IsDirIntent=true target with recursive=false → `is a
//     directory: pass -r/-R to remove it` (matches Unix `rm`'s
//     refusal).
//   - any Target with empty Name → "refusing to delete the root of
//     <fileType>/<extend>". This guards against `rm drive/Home/`
//     accidentally meaning "wipe my Drive" — that operation would
//     have to be expressed differently.
func Plan(targets []Target, recursive bool) ([]*Group, error) {
	if len(targets) == 0 {
		return nil, errors.New("rm: no targets supplied")
	}

	type key struct {
		fileType, extend, parent string
	}
	groupIdx := map[key]int{}
	var groups []*Group

	// Track per-(group, dirent) seen-set so duplicates in the user's
	// input collapse to one wire entry. We DO want duplicate names
	// across different parents to land in their own groups.
	seen := map[string]map[string]struct{}{}

	for _, t := range targets {
		if t.Name == "" {
			return nil, fmt.Errorf("refusing to delete the root of %s/%s",
				t.FileType, t.Extend)
		}
		// LarePass-aligned policy for external mountpoints: the
		// AI-mountpoint surface Olares creates at the external
		// node root holds a backend-owned layout (depth-1 `ai/`
		// plus the four depth-2 feature dirs output / model /
		// comfyui / ollama) that the LarePass GUI greys out via
		// `externalFolderWhiteList` + `externalAiFolderWhiteList`
		// in apps/packages/app/src/stores/operation.ts (gated by
		// `ieExternalRootPath` and `isExternalAiPath`). These
		// directories are the contract Ollama / ComfyUI /
		// Huggingface readers look up by name; a scripted
		// `files rm -r external/<node>/ai/` would silently break
		// those apps without any GUI affordance to recover.
		//
		// IMPORTANT — order: this guard runs BEFORE
		// isExternalVolumeRoot below because the depth-1 entry
		// `external/<node>/ai/` is ALSO depth-1 under the node
		// (so the generic volume-root rejection would otherwise
		// fire first with a wrong, "this is a mounted disk"
		// message). The AI-mountpoint message is more accurate:
		// `ai` is a system-managed entry surfaced alongside the
		// real disk mounts, not itself a removable disk.
		//
		// Scope is EXACT: user content nested deeper —
		// e.g. external/<node>/ai/output/run-2026/ — stays
		// freely deletable so per-run cleanups keep working.
		// Authoritative names live in cli/cmd/ctl/files/path.go's
		// ProtectedExternalChildren + ProtectedExternalAiChildren.
		if isProtectedExternalChild(t.FileType, t.Extend, t.ParentSubPath, t.Name) {
			return nil, fmt.Errorf(
				"refusing to delete external/%s%s%s/: this is a system-managed AI mountpoint folder reserved by Files; "+
					"the LarePass GUI also disables delete on this entry. "+
					"Names refused at external/<node>/: {%s}; at external/<node>/ai/: {%s}. "+
					"Delete its contents instead (e.g. files rm -r external/%s%s%s/<entry>) and the folder itself will stay.",
				t.Extend, t.ParentSubPath, t.Name,
				protectedExternalChildrenList, protectedExternalAiChildrenList,
				t.Extend, t.ParentSubPath, t.Name)
		}
		// external/<node>/<volume>/ refusal — symmetric with
		// mkdir.Plan's depth-1 guard (cli/internal/files/mkdir/
		// mkdir.go). The volume itself is a LarePass-managed
		// mount point (USB-0, SMB-..., per-disk entries surfaced
		// at the external/<node>/ volume-listing layer), not a
		// regular directory; the files-backend's DELETE handler
		// on /api/resources/external/<node>/<volume>/ doesn't
		// know how to unmount the volume — it just iterates the
		// volume's contents and removes them, which presents as
		// a successful 2xx but actually wipes everything the
		// user had stored on that disk. That's a "blast radius
		// of an entire mount point" hazard, same flavor as
		// `rm -rf drive/Home/`, so we refuse client-side and
		// point at the only meaningful shape (depth ≥ 2 under
		// the node). The match condition mirrors
		// frontendPathToRmTarget's split: parent reduces to
		// "" / "/" exactly when the user typed
		// `external/<node>/<single-segment>/`.
		//
		// The depth-1 `ai` whitelist (isProtectedExternalChild
		// above) preempts this branch with a more specific
		// message; everything else at depth-1 lands here.
		if isExternalVolumeRoot(t.FileType, t.ParentSubPath, t.Name) {
			return nil, fmt.Errorf(
				"refusing to delete external/%s/%s/: this is a mounted volume root (managed via LarePass, not via files-backend rm); "+
					"the DELETE handler would iterate the volume and wipe everything on the underlying disk. "+
					"Point at a sub-path inside the volume instead, e.g. external/%s/%s/<sub>/.",
				t.Extend, t.Name, t.Extend, t.Name)
		}
		// LarePass-aligned policy: the system-managed first-level
		// children directly under drive/Home/ (Pictures / Music /
		// Movies / Downloads / Documents / Code / Ollama /
		// Huggingface / Cache / Data / Home) refuse deletion. The
		// LarePass GUI greys out the delete action for these
		// entries when the user is at /Files/Home/ via the
		// `disableMenuItem` array in
		// apps/packages/app/src/stores/operation.ts; without this
		// guard a scripted `files rm -r drive/Home/Pictures` would
		// produce a state the GUI couldn't reach (and would
		// destroy bootstrap directories that user apps assume
		// exist).
		//
		// The match scope is the EXACT first-level entry —
		// children of these dirs (e.g. drive/Home/Pictures/Trip/)
		// are user content and remain freely deletable. The
		// authoritative names live in cli/cmd/ctl/files/path.go's
		// ProtectedDriveHomeChildren; the duplicate map below
		// keeps this package free of a cmd/ctl/files dependency
		// (same pattern internal/files/* uses for other policy).
		if isProtectedDriveHomeChild(t.FileType, t.Extend, t.ParentSubPath, t.Name) {
			return nil, fmt.Errorf(
			"refusing to delete drive/Home/%s: this is a system-managed Home folder reserved by Files; "+
				"the Files GUI also disables delete for {%s} under drive/Home/. "+
				"Delete its contents instead (e.g. files rm -r drive/Home/%s/<entry>) and the folder itself will stay.",
				t.Name, protectedDriveHomeChildrenList, t.Name)
		}
		if t.IsDirIntent && !recursive {
			return nil, fmt.Errorf(
				"%s/%s%s%s is a directory: pass -r/-R to remove it recursively",
				t.FileType, t.Extend, t.ParentSubPath, t.Name)
		}
		dirent := "/" + t.Name
		// `-r` on the command line is the user's declaration that
		// every supplied target is a directory; force the trailing
		// slash so the wire request matches the Unix-style `rm -r
		// foo` shape regardless of whether the user typed `foo` or
		// `foo/`. Without `-r`, IsDirIntent has already been
		// rejected by the guard above, so this branch is a pure
		// "treat as file" shape.
		if t.IsDirIntent || recursive {
			dirent += "/"
		}

		k := key{t.FileType, t.Extend, t.ParentSubPath}
		idx, ok := groupIdx[k]
		if !ok {
			idx = len(groups)
			groupIdx[k] = idx
			groups = append(groups, &Group{
				FileType:      t.FileType,
				Extend:        t.Extend,
				ParentSubPath: t.ParentSubPath,
			})
			seen[k.parent+"|"+k.fileType+"|"+k.extend] = map[string]struct{}{}
		}
		seenKey := k.parent + "|" + k.fileType + "|" + k.extend
		if _, dup := seen[seenKey][dirent]; dup {
			continue
		}
		seen[seenKey][dirent] = struct{}{}
		groups[idx].Dirents = append(groups[idx].Dirents, dirent)
	}

	// Stable orderings: groups by (fileType, extend, parent), dirents
	// by name within each group.
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].FileType != groups[j].FileType {
			return groups[i].FileType < groups[j].FileType
		}
		if groups[i].Extend != groups[j].Extend {
			return groups[i].Extend < groups[j].Extend
		}
		return groups[i].ParentSubPath < groups[j].ParentSubPath
	})
	for _, g := range groups {
		sort.Strings(g.Dirents)
	}

	return groups, nil
}

// protectedDriveHomeChildren mirrors the LarePass web app's
// disableMenuItem array (apps/packages/app/src/stores/operation.ts)
// that the GUI uses — gated by `path === '/Files/Home/'` — to grey
// out cut / copy / delete / rename for system-managed Home children.
// Names are case-sensitive and align 1:1 with the GUI, including
// LarePass-quirk casings like the one-word `Huggingface`.
//
// Authoritative source on the CLI is
// cli/cmd/ctl/files/path.go's ProtectedDriveHomeChildren — keep the
// two synchronized if the web app's array ever changes.
var protectedDriveHomeChildren = map[string]struct{}{
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

const protectedDriveHomeChildrenList = "Cache, Code, Data, Documents, Downloads, Home, Huggingface, Movies, Music, Ollama, Pictures"

// isProtectedDriveHomeChild reports whether (fileType, extend,
// parent, name) addresses one of the system-managed first-level
// children directly under drive/Home/. Mirrors
// FrontendPath.IsProtectedDriveHomeChild semantics from
// cli/cmd/ctl/files/path.go: only EXACT first-level entries match
// (parent must be the drive/Home/ root), and the comparison is
// case-sensitive against protectedDriveHomeChildren.
// isExternalVolumeRoot reports whether (fileType, parent, name)
// addresses the bare volume root under external/<node>/. The
// frontendPathToRmTarget split lands `external/<node>/<volume>/`
// as ParentSubPath="/" + Name="<volume>" + IsDirIntent=true, so
// the canonical "this is a depth-1 entry under external/<node>"
// shape is exactly the parent-is-root + non-empty-name pair on
// the external fileType. Symmetric with mkdir.Plan's depth-1
// guard which uses `!strings.Contains(clean, "/")` on the
// pre-split clean path — both end up rejecting the same wire
// surface. See SKILL.md's "volume listing layer" note for the
// user-visible contract.
func isExternalVolumeRoot(fileType, parent, name string) bool {
	if fileType != "external" {
		return false
	}
	if name == "" {
		return false
	}
	if strings.Trim(parent, "/") != "" {
		return false
	}
	return true
}

func isProtectedDriveHomeChild(fileType, extend, parent, name string) bool {
	if fileType != "drive" || extend != "Home" {
		return false
	}
	if strings.Trim(parent, "/") != "" {
		return false
	}
	if name == "" {
		return false
	}
	_, ok := protectedDriveHomeChildren[name]
	return ok
}

// protectedExternalChildren mirrors the LarePass web app's
// externalFolderWhiteList (apps/packages/app/src/stores/operation.ts):
// the entry names directly under `external/<node>/` whose rename /
// delete affordance is greyed out by `isDisableMenuItem` when gated
// by `ieExternalRootPath` (regex `^/Files/External/[^/]+/?$`).
// Currently just `ai` — the AI mountpoint surface Olares surfaces
// at the node root alongside the real disk mounts.
//
// Path-shape note: the GUI URL `/Files/External/<X>/` and the CLI
// path `external/<X>/` share `<X>` as the LarePass `masterNode`
// (see apps/.../external/data.ts:77), which is FrontendPath.Extend
// on the CLI. The GUI's `ai/` row at `/Files/External/<node>/`
// therefore maps to CLI `external/<node>/ai/` — ParentSubPath="/"
// + Name="ai" in this package's (parent, name) split.
//
// Authoritative source on the CLI is
// cli/cmd/ctl/files/path.go's ProtectedExternalChildren — keep the
// two in sync if the web app's array ever changes (the
// duplicate-in-internal pattern matches protectedDriveHomeChildren
// above; same reason — internal/files/rm should not depend on
// cmd/ctl/files for policy data).
var protectedExternalChildren = map[string]struct{}{
	"ai": {},
}

// protectedExternalAiChildren mirrors externalAiFolderWhiteList from
// the same operation.ts store: entries directly under
// `external/<node>/ai/` (depth-2 in SubPath) that the GUI refuses
// to rename / delete when gated by `isExternalAiPath` (regex
// `^/Files/External/[^/]+/ai/?$`). These are the per-feature
// directories shared by the apps that use the AI mountpoint
// (output / model / comfyui / ollama). Authoritative source:
// cli/cmd/ctl/files/path.go's ProtectedExternalAiChildren.
var protectedExternalAiChildren = map[string]struct{}{
	"output":  {},
	"model":   {},
	"comfyui": {},
	"ollama":  {},
}

const (
	protectedExternalChildrenList   = "ai"
	protectedExternalAiChildrenList = "comfyui, model, ollama, output"
)

// isProtectedExternalChild reports whether (fileType, extend,
// parent, name) addresses one of the LarePass-managed AI mountpoint
// folders the GUI refuses rename / delete on. Mirrors
// FrontendPath.IsProtectedExternalChild from
// cli/cmd/ctl/files/path.go but uses the rm package's (parent, name)
// split shape (frontendPathToRmTarget peels off the trailing
// segment).
//
// Two layers, both anchored at the external NODE root (Extend),
// NOT inside any nested `<volume>` segment:
//
//   - depth-1: parent reduces to "" (empty/root), name is in
//     protectedExternalChildren (currently just "ai"). The wire
//     shape `external/<node>/ai/` lands as ParentSubPath="/" +
//     Name="ai".
//   - depth-2: parent reduces to "ai" (one trimmed segment), name
//     is in protectedExternalAiChildren (output / model / comfyui
//     / ollama). The wire shape `external/<node>/ai/<entry>/` lands
//     as ParentSubPath="/ai/" + Name="<entry>".
//
// Extend (the `<node>` segment) is opaque to the policy — the GUI
// regex `^/Files/External/[^/]+/...` ignores the node entirely.
//
// Scope is EXACT: any deeper parent ("/ai/output/..." or deeper)
// is user content and stays freely deletable so per-run /
// per-experiment cleanups keep working.
func isProtectedExternalChild(fileType, extend, parent, name string) bool {
	if fileType != "external" {
		return false
	}
	_ = extend
	if name == "" {
		return false
	}
	trimmed := strings.Trim(parent, "/")
	if trimmed == "" {
		// parent is root → external/<node>/<name>/ shape (depth-1).
		_, ok := protectedExternalChildren[name]
		return ok
	}
	if trimmed == "ai" {
		// parent is /ai/, name = <entry> (depth-2). Anchor on
		// the exact "ai" parent — any other depth-1 dir is user
		// or volume content (the GUI's `isExternalAiPath` regex
		// only matches `/ai/`).
		_, ok := protectedExternalAiChildren[name]
		return ok
	}
	return false
}

// deleteRequestBody is the JSON body shape the files-backend's DELETE
// resource handler binds to (resources_service.go line 209). We keep
// the wire representation in a typed struct rather than a map so any
// future field additions show up in code review.
type deleteRequestBody struct {
	Dirents []string `json:"dirents"`
}

// DeleteBatch sends one DELETE request for the given group. URL is
// `<BaseURL>/api/resources/<encFileType>/<encExtend><encParent>` and
// body is `{"dirents": [...]}`.
//
// `parent` always ends with '/' on the wire; if the caller's
// ParentSubPath is missing the trailing slash we add it here so the
// FileParam.convert split-on-/ check on the server side passes.
func (c *Client) DeleteBatch(ctx context.Context, g *Group) error {
	if len(g.Dirents) == 0 {
		return nil
	}
	parent := g.ParentSubPath
	if !strings.HasSuffix(parent, "/") {
		parent += "/"
	}
	plain := g.FileType + "/" + g.Extend + parent
	endpoint := c.BaseURL + "/api/resources/" + encodepath.EncodeURL(plain)

	bodyBytes, err := json.Marshal(deleteRequestBody{Dirents: g.Dirents})
	if err != nil {
		return fmt.Errorf("marshal delete body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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
			Method: http.MethodDelete,
		}
	}
	return nil
}
