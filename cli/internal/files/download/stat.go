// stat.go: figure out whether a remote path is a file or a directory
// and how big it is. Used by:
//
//   - the download cobra cmd, to decide single-file vs. recursive
//     directory mode and to print a remote-size line up-front;
//   - the cat cobra cmd, to refuse early when the user points it at a
//     directory (the `/api/raw` endpoint would 400, but the error
//     message is much friendlier to surface here);
//   - the recursive walker, to seed the traversal at the user-supplied
//     root.
//
// Implementation strategy: list the parent directory and look up the
// basename in its items array, exactly like the LarePass web app does
// (every navigation in the UI uses the parent's listing for per-entry
// metadata, never a single-resource probe). Why not GET
// /api/resources/<encFilePath> directly? — see the comment on
// statByParentListing for the gory details, but the short version is
// that the backend's "List" handler is hard-coded to set
// `Content: true` (files/pkg/drivers/posix/posix/posix.go's getFiles)
// and tries to slurp the file's contents, returning HTTP 500 for
// json / binary / large files. We only need (Name, IsDir, Size), so
// the parent-listing path is both more reliable and strictly cheaper
// than fetching content we don't want.
package download

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// StatInfo is the projection of files-backend's per-resource envelope
// that the download flow needs. The backend's full FileInfo carries
// many more fields (mode / modified / type / numFiles / numDirs / ...);
// we only decode what download / cat / walker actually branch on.
type StatInfo struct {
	// Name is the basename the backend reports for the resource.
	Name string
	// IsDir is true when the resource is a directory.
	IsDir bool
	// Size is the file size in bytes; meaningless for directories
	// (the backend may report 0 or the sum of immediate children
	// depending on driver — we don't rely on this for dirs).
	Size int64
}

// Stat resolves `plainPath` (an un-encoded `<fileType>/<extend>/<sub>`
// triple) to a (Name, IsDir, Size) record, by listing the parent
// directory and looking up the basename. The parent-listing strategy
// is what the LarePass web app uses; a direct
// GET /api/resources/<encFilePath> is unreliable on the current
// backend (returns HTTP 500 for many real files because the underlying
// List handler tries to read file content into the response).
//
// Special cases:
//   - When `plainPath` resolves to the root of `<fileType>/<extend>`
//     (e.g. "drive/Home", "sync/<repo>") there is no parent to list
//     — Stat returns a synthetic IsDir=true record. This matches what
//     `files ls drive/Home/` does logically (the volume root is
//     always a directory).
//   - When the parent directory itself doesn't exist OR auth fails,
//     the underlying List error is returned verbatim; callers use
//     IsNotFound to distinguish "this path doesn't exist on the
//     server" from "your token is bad / network is down".
//   - When the parent exists but the basename isn't in its items
//     array, Stat returns an *HTTPError with Status=404 so callers
//     can branch on IsNotFound uniformly.
func (c *Client) Stat(ctx context.Context, plainPath string) (*StatInfo, error) {
	clean := strings.Trim(plainPath, "/")
	if clean == "" {
		return nil, errors.New("Stat: empty path")
	}
	segs := strings.Split(clean, "/")
	// Need at least 3 segments (fileType / extend / leaf) to have a
	// parent under <fileType>/<extend>. With 2 or fewer segments
	// we're already at the volume root — synthesise a dir record.
	if len(segs) <= 2 {
		return &StatInfo{Name: segs[len(segs)-1], IsDir: true}, nil
	}
	return c.statByParentListing(ctx, segs)
}

// statByParentListing implements the parent-list-and-lookup strategy.
//
// Why we don't probe GET /api/resources/<encFilePath>:
//
//	files/pkg/hertz/biz/handler/api/resources/resources_service.go's
//	GetResourcesMethod always invokes Storage.List, and
//	files/pkg/drivers/posix/posix/posix.go's List in turn calls
//	getFiles(..., Expand, Content) — Content=true means the backend
//	tries to read the entire file into the response on a single-file
//	GET. That blows up (HTTP 500) for json / binary / large files,
//	even though the metadata it returns first would have been
//	perfectly fine. Listing the parent + finding the entry there
//	side-steps this entirely and matches what the LarePass web app
//	already does.
func (c *Client) statByParentListing(ctx context.Context, segs []string) (*StatInfo, error) {
	leaf := segs[len(segs)-1]
	parentSegs := segs[:len(segs)-1]
	parentPath := strings.Join(parentSegs, "/") + "/"

	entries, err := c.List(ctx, parentPath)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Name == leaf {
			return &StatInfo{Name: e.Name, IsDir: e.IsDir, Size: e.Size}, nil
		}
	}
	// Parent listing succeeded but our leaf isn't there — synthesise
	// a 404 so the caller's IsNotFound predicate fires.
	return nil, &HTTPError{
		Status: http.StatusNotFound,
		Body:   fmt.Sprintf("entry %q not found in parent listing", leaf),
		URL:    c.resourcesURL(parentPath),
		Method: http.MethodGet,
	}
}

// IsNotFound reports whether `err` represents "this remote path
// doesn't exist". Two cases collapse here:
//
//   - the parent directory listing returned 404 (HTTPError);
//   - the parent listing succeeded but the leaf basename wasn't in
//     its items (synthetic 404 from statByParentListing).
//
// Callers (the download cobra cmd) use this to decide whether to
// emit a "did you mean ...?" hint vs. a generic auth/network error.
// errors.As keeps the predicate robust if the error gets wrapped by
// a higher layer.
func IsNotFound(err error) bool {
	var hErr *HTTPError
	if errors.As(err, &hErr) {
		return hErr.Status == http.StatusNotFound
	}
	return false
}
