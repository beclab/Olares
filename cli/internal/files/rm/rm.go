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
// AccessToken is sent as `X-Authorization` (not `Authorization: Bearer`),
// because Olares' edge stack only forwards the X-Authorization header to
// per-user services. See pkg/cmdutil/factory.go for the full rationale.
type Client struct {
	HTTPClient  *http.Client
	BaseURL     string // FilesURL, e.g. https://files.alice.olares.com
	AccessToken string
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
		if t.IsDirIntent && !recursive {
			return nil, fmt.Errorf(
				"%s/%s%s%s is a directory: pass -r/-R to remove it recursively",
				t.FileType, t.Extend, t.ParentSubPath, t.Name)
		}
		dirent := "/" + t.Name
		if t.IsDirIntent {
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
	if c.AccessToken != "" {
		req.Header.Set("X-Authorization", c.AccessToken)
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
