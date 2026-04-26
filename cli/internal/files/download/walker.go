// walker.go: turn a remote directory path into a flat list of files to
// download plus the empty subdirectories that need to be created
// locally. Symmetric to upload's walker.go (which goes the other way),
// but drives a remote `List` instead of a local `filepath.WalkDir`.
//
// Path semantics:
//
//   - The remote root's own basename becomes the top-level under the
//     local destination, matching the LarePass folder-download UX
//     (download `drive/Home/Documents/` into `./out/` produces
//     `./out/Documents/...`). To pull contents only, the user can
//     pass an explicit local target path that already includes the
//     leaf name.
//   - All wire paths use POSIX '/'; they're converted to host paths
//     via filepath.Join only at the local-write boundary.
package download

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// FileTask is one file scheduled for download. The download cobra
// command turns each task into a DownloadFile call, optionally in
// parallel via errgroup.
type FileTask struct {
	// RemotePlainPath is the un-encoded `<fileType>/<extend>/<sub>`
	// triple Client.DownloadFile expects. Always points at a regular
	// file (the walker filters out directories before adding tasks).
	RemotePlainPath string
	// LocalPath is the full local destination path (including the
	// recreated directory tree). Parent directories are created on
	// the fly by DownloadFile via os.MkdirAll.
	LocalPath string
	// RelativePath is the path RELATIVE to the user-supplied local
	// destination root, in POSIX form. Useful for progress display
	// (`./out/Documents/foo.txt` reads more naturally as
	// `Documents/foo.txt`).
	RelativePath string
	// Size is the file size in bytes from the remote listing. May be
	// 0 for genuinely empty files; relied on by the downloader's
	// resume probe and by the cobra command's progress totals.
	Size int64
}

// Plan is the structured result of resolving a remote root + local
// destination pair. The cobra command consumes Files in parallel and
// (optionally) creates EmptyDirs locally before the file downloads
// start so the on-disk tree matches the remote one even if a
// directory happens to be empty.
type Plan struct {
	// Files is the flat list of file download tasks. Ordering is
	// deterministic (sorted by RelativePath) so retries / dry-runs
	// are stable across runs.
	Files []FileTask
	// EmptyDirs lists subdirectories under LocalRoot that contain no
	// files. Sorted shallow-to-deep so iterative MkdirAll works
	// without surprises. May be empty for a fully-populated tree.
	EmptyDirs []string
	// LocalRoot is the local directory the plan is anchored to —
	// either the user's --dst (in directory mode) or that path with
	// the remote basename appended (the LarePass-folder-picker
	// behavior). Useful for the cobra command's summary line.
	LocalRoot string
}

// BuildPlan walks the remote tree rooted at `remoteRoot` and lays
// out the corresponding local file paths under `localBase`.
//
// `remoteRoot` is the un-encoded plain path like `drive/Home/Documents`
// (the trailing slash is added internally — Stat/List handle either
// form). `localBase` is the user-supplied local destination directory
// (must exist OR be createable; the cobra cmd validates this before
// calling).
//
// On success Plan.Files is non-empty for any non-empty remote tree;
// callers should still defensively handle the all-empty case (a
// directory with only empty subdirectories) by checking len(Files).
func BuildPlan(
	ctx context.Context,
	c *Client,
	remoteRoot, localBase string,
) (*Plan, error) {
	// Pull the leaf name off the remote root so we can recreate it as
	// the top-level directory under localBase, matching the upload
	// walker's "preserve the source folder name" behavior.
	cleanRoot := strings.Trim(remoteRoot, "/")
	if cleanRoot == "" {
		return nil, fmt.Errorf("remote root is empty")
	}
	leaf := path.Base(cleanRoot)
	plan := &Plan{
		LocalRoot: filepath.Join(localBase, leaf),
	}

	// The walker is depth-first iterative (BFS would be fine too;
	// pick depth-first so progress output shows files within a
	// directory together, which reads more naturally to humans).
	type frame struct {
		// remotePath is the parent directory's plain path on the
		// server, e.g. "drive/Home/Documents/photos".
		remotePath string
		// relPath is the same path under the local LocalRoot, in
		// POSIX form (e.g. "photos" or "photos/2026"). Empty for the
		// root itself.
		relPath string
	}
	stack := []frame{{remotePath: cleanRoot, relPath: ""}}

	for len(stack) > 0 {
		// Pop from the end (depth-first).
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		entries, err := c.List(ctx, top.remotePath+"/")
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", top.remotePath, err)
		}

		if len(entries) == 0 && top.relPath != "" {
			// Genuine empty subdirectory: the local tree should mirror
			// it. The root itself is excluded because we always create
			// it at the cobra layer.
			plan.EmptyDirs = append(plan.EmptyDirs, top.relPath)
			continue
		}

		// Push children — directories onto the stack for further
		// traversal, files into Plan.Files. Sort entries inside
		// the loop so the depth-first traversal yields a stable
		// per-directory ordering even if the server hands us an
		// unsorted list (which it generally doesn't, but we don't
		// want to bake that assumption in).
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				// Files first inside this directory so their progress
				// line lands before we descend into the next subdir's
				// listing.
				return !entries[i].IsDir
			}
			return entries[i].Name < entries[j].Name
		})

		for _, e := range entries {
			childRel := joinPosix(top.relPath, e.Name)
			childRemote := top.remotePath + "/" + e.Name

			if e.IsDir {
				// Push subdirectory; it'll be popped after this loop
				// and processed (which may produce further pushes).
				stack = append(stack, frame{
					remotePath: childRemote,
					relPath:    childRel,
				})
				continue
			}
			plan.Files = append(plan.Files, FileTask{
				RemotePlainPath: childRemote,
				LocalPath:       filepath.Join(plan.LocalRoot, filepath.FromSlash(childRel)),
				RelativePath:    childRel,
				Size:            e.Size,
			})
		}
	}

	// Stable, predictable orderings for retries / dry-runs. Files are
	// sorted by RelativePath; empty dirs shallow-first so iterative
	// MkdirAll matches expectation.
	sort.SliceStable(plan.Files, func(i, j int) bool {
		return plan.Files[i].RelativePath < plan.Files[j].RelativePath
	})
	sort.SliceStable(plan.EmptyDirs, func(i, j int) bool {
		di := strings.Count(plan.EmptyDirs[i], "/")
		dj := strings.Count(plan.EmptyDirs[j], "/")
		if di != dj {
			return di < dj
		}
		return plan.EmptyDirs[i] < plan.EmptyDirs[j]
	})

	return plan, nil
}

// joinPosix joins two POSIX path fragments with a single '/'. Empty
// `parent` means `child` is the top-level entry under LocalRoot.
func joinPosix(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "/" + child
}
