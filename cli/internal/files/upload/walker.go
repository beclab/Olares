// walker.go: turn a local-path + remote-path pair into a flat list of
// per-file upload tasks (UploadOpts) plus the empty-directory mkdirs
// the chunk-only protocol can't express on its own.
//
// Path semantics (deliberately rsync-LIKE-but-not-rsync):
//
//   - <local> is a regular file:
//   - <remote> ends with '/' → upload to <remote>/<basename(local)>
//   - else                   → upload to <remote> (treat as full target
//     path, i.e. user is renaming on the way in)
//   - <local> is a directory:
//   - <remote> MUST end with '/' (the destination is a directory).
//   - The walker recursively emits every regular file under <local>;
//     each file's RelativePath includes <basename(local)> as the
//     top-level component so the source folder's name appears under
//     <remote> on the server (i.e. `upload mydir drive/Home/X/` →
//     drive/Home/X/mydir/...). This matches the LarePass folder-upload
//     UI, which always preserves the picked folder's name.
//   - Empty subdirectories are recorded as EmptyDirs so the cobra
//     command can pre-mkdir them before the chunk uploads start —
//     Resumable.js's chunk pipeline can't represent a 0-byte
//     directory entry on its own.
//
// All wire-level paths use POSIX-style '/' separators regardless of the
// host OS, because the server expects forward-slash paths.
package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileTask is one regular file to upload. The cobra command turns each
// FileTask into an UploadOpts and pushes it through Client.UploadFile;
// see Plan.ToUploadOpts for the conversion (kept on Plan so a future
// `--dry-run` can render tasks without actually uploading).
type FileTask struct {
	// LocalPath is the on-disk path (absolute or working-directory-
	// relative — same form the user passed in).
	LocalPath string
	// RelativePath is the file's path relative to the destination
	// parent_dir, in POSIX form. For a single-file upload this is just
	// the basename / target name; for a folder upload this includes
	// the source-folder prefix (e.g. "mydir/sub/foo.txt").
	RelativePath string
	// RemoteName is the bare basename for the upload (typically the
	// last segment of RelativePath; for the "rename on upload" single-
	// file case it differs from filepath.Base(LocalPath)).
	RemoteName string
	// Size is the file size in bytes at plan time. Useful for sorting
	// (largest-first scheduling) and for the progress display.
	Size int64
}

// Plan is the structured result of resolving a (<local>, <remote>) pair
// against the local filesystem. The cobra command consumes it directly:
// run EmptyDirs through Client.Mkdir, then schedule Files through an
// errgroup of Client.UploadFile.
type Plan struct {
	// ParentDir is the constant `/drive/Home/...` parent directory
	// (with trailing '/') that the upload session is anchored to. Same
	// value goes into UploadOpts.ParentDir for every file.
	ParentDir string
	// RelativeRoot is the path RELATIVE to /Home that ParentDir maps
	// to (no leading or trailing slash, e.g. "Documents/Backups"). The
	// cobra command passes this to Client.Mkdir to ensure the
	// destination dir itself exists before any file upload runs. May
	// be empty when uploading directly to /Home.
	RelativeRoot string
	// EmptyDirs lists the additional sub-directories (POSIX-relative
	// to RelativeRoot, no leading slash, no trailing slash) that need
	// to be pre-created because the source contains them but no files
	// were emitted underneath. Sorted shallow-to-deep so naive
	// sequential mkdir works.
	EmptyDirs []string
	// Files is the flat list of per-file upload tasks. Order is
	// deterministic (sorted by RelativePath) so retries / dry-runs are
	// stable.
	Files []FileTask
}

// BuildPlan validates inputs against the local filesystem and returns
// a Plan ready to be executed.
//
// `remoteSubPath` is the path RELATIVE to drive/Home as parsed from a
// FrontendPath (so "Documents/Backups" or "Documents/Backups/" — with
// or without leading slash; both are accepted). The trailing slash IS
// significant: it tells BuildPlan to interpret the remote as a
// directory rather than a file rename target.
//
// Errors:
//   - <local> doesn't exist
//   - <local> is a directory but <remote> doesn't end with '/'
func BuildPlan(localPath, remoteSubPath string) (*Plan, error) {
	st, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", localPath, err)
	}

	remoteIsDir := strings.HasSuffix(remoteSubPath, "/")
	cleanRemote := strings.Trim(remoteSubPath, "/")

	if st.Mode().IsRegular() {
		return planForFile(localPath, st.Size(), cleanRemote, remoteIsDir), nil
	}
	if st.IsDir() {
		if !remoteIsDir {
			return nil, fmt.Errorf("local %q is a directory; remote %q must end with '/'",
				localPath, remoteSubPath)
		}
		return planForDir(localPath, cleanRemote)
	}
	return nil, fmt.Errorf("%s is not a regular file or directory", localPath)
}

// planForFile handles the single-regular-file branch.
//
//   - remoteIsDir (caller-supplied trailing '/'): upload as
//     <cleanRemote>/<basename>. cleanRemote is the parent_dir.
//   - !remoteIsDir: cleanRemote is the full target path; we split it
//     into parent + basename so parent_dir is well-formed (the chunk
//     POST always wants a directory parent_dir, never a full path).
func planForFile(localPath string, size int64, cleanRemote string, remoteIsDir bool) *Plan {
	base := filepath.Base(localPath)

	var (
		relativeRoot string // dir under /Home (no slashes)
		remoteName   string // file basename on server
	)
	if remoteIsDir {
		relativeRoot = cleanRemote
		remoteName = base
	} else if cleanRemote == "" {
		// "/" with no trailing slash and no body — upload to /Home/<base>.
		relativeRoot = ""
		remoteName = base
	} else {
		// Treat cleanRemote as the full destination path.
		idx := strings.LastIndex(cleanRemote, "/")
		if idx < 0 {
			relativeRoot = ""
			remoteName = cleanRemote
		} else {
			relativeRoot = cleanRemote[:idx]
			remoteName = cleanRemote[idx+1:]
		}
	}

	return &Plan{
		ParentDir:    parentDirFor(relativeRoot),
		RelativeRoot: relativeRoot,
		Files: []FileTask{{
			LocalPath:    localPath,
			RelativePath: remoteName,
			RemoteName:   remoteName,
			Size:         size,
		}},
	}
}

// planForDir handles the recursive directory branch. The source
// directory's basename becomes the top-level under the destination
// (so `upload mydir drive/Home/X/` produces drive/Home/X/mydir/...).
// To upload contents-only without the wrapper folder, the user can do
// `upload mydir/sub drive/Home/X/mydir/sub/` — i.e. specify the
// destination explicitly. We deliberately do NOT support the rsync
// `local/` "contents only" trailing-slash convention in this first
// cut because it conflicts with the file-vs-directory disambiguation
// trailing slashes carry on the remote side; the few users who need
// it can chain mv on the server.
func planForDir(localDir, cleanRemote string) (*Plan, error) {
	srcBase := filepath.Base(localDir)
	relativeRoot := strings.Trim(cleanRemote, "/")

	plan := &Plan{
		ParentDir:    parentDirFor(relativeRoot),
		RelativeRoot: relativeRoot,
	}

	// Pre-collect: walk first, then sort + decide which dirs are
	// "empty" (no files emitted under them). This is O(N) extra memory
	// over a streaming walk but lets us emit a deterministic plan + a
	// clean EmptyDirs list, which makes the resulting upload trivially
	// resumable across runs (no order-of-operations subtleties).
	var allDirs []string
	dirHasFile := map[string]bool{}

	err := filepath.WalkDir(localDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// Compute the path relative to localDir, then prefix with
		// srcBase so the source folder name appears in the upload tree.
		rel, relErr := filepath.Rel(localDir, path)
		if relErr != nil {
			return relErr
		}
		// On Windows, filepath.Rel would return backslash-separated.
		// Normalize to POSIX for the wire.
		rel = filepath.ToSlash(rel)
		if rel == "." {
			rel = ""
		}

		// posixRel is the relative path under ParentDir on the server
		// — always includes the source basename as the first segment.
		posixRel := srcBase
		if rel != "" {
			posixRel = srcBase + "/" + rel
		}

		if d.IsDir() {
			// Skip the source root itself; "<srcBase>" gets implicitly
			// created when its first file lands. Track sub-directories
			// so we can mkdir the empty ones explicitly later.
			if rel == "" {
				return nil
			}
			allDirs = append(allDirs, posixRel)
			return nil
		}
		if !d.Type().IsRegular() {
			// Skip symlinks / devices / sockets — we don't want to
			// silently follow links and explode the upload, and we
			// can't meaningfully upload a device.
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		plan.Files = append(plan.Files, FileTask{
			LocalPath:    path,
			RelativePath: posixRel,
			RemoteName:   filepath.Base(path),
			Size:         info.Size(),
		})
		// Mark every ancestor directory of this file as "has-file" so
		// it doesn't show up in EmptyDirs.
		for d := dirOfPosix(posixRel); d != ""; d = dirOfPosix(d) {
			if dirHasFile[d] {
				break
			}
			dirHasFile[d] = true
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", localDir, err)
	}

	// Determine the empty dirs: subtree dirs the walk saw that don't
	// contain any file. Sorted shallow-first so iterative mkdir works.
	for _, d := range allDirs {
		if !dirHasFile[d] {
			plan.EmptyDirs = append(plan.EmptyDirs, d)
		}
	}
	sort.SliceStable(plan.EmptyDirs, func(i, j int) bool {
		return depth(plan.EmptyDirs[i]) < depth(plan.EmptyDirs[j])
	})
	sort.SliceStable(plan.Files, func(i, j int) bool {
		return plan.Files[i].RelativePath < plan.Files[j].RelativePath
	})

	return plan, nil
}

// ToUploadOpts converts one FileTask into an UploadOpts ready for
// Client.UploadFile, threading per-call settings (node, chunk size,
// retries) from the cobra command. Side-effect free.
func (p *Plan) ToUploadOpts(t FileTask, node string, chunkSize int64, maxRetries int) UploadOpts {
	return UploadOpts{
		LocalPath:    t.LocalPath,
		Node:         node,
		ParentDir:    p.ParentDir,
		RemoteName:   t.RemoteName,
		RelativePath: t.RelativePath,
		ChunkSize:    chunkSize,
		MaxRetries:   maxRetries,
	}
}

// parentDirFor returns the `/drive/Home/...` parent_dir form for a
// /Home-relative directory. Always ends with '/'.
func parentDirFor(relativeRoot string) string {
	rr := strings.Trim(relativeRoot, "/")
	if rr == "" {
		return "/drive/Home/"
	}
	return "/drive/Home/" + rr + "/"
}

// dirOfPosix returns the POSIX-style parent directory of `p`, or ""
// when p has no '/'. Used to walk up the ancestor chain when marking
// dirs that contain files.
func dirOfPosix(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		return ""
	}
	return p[:idx]
}

// depth counts '/' separators; used to sort EmptyDirs shallow-first
// (ancestor dirs before descendants) so mkdir order is correct.
func depth(p string) int {
	return strings.Count(p, "/")
}
