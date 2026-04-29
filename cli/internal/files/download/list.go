// list.go: list a remote directory via GET /api/resources/<encPath>/.
// The walker calls this once per directory to drive the recursive
// download. This is not the `olares-cli files ls` implementation (that
// lives in cli/cmd/ctl/files/ls.go); we only share the same JSON envelope
// shape (see ls.go's listingResponse) but project down to name + isDir +
// size — everything the download walker needs.
package download

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Entry is one item in a directory listing.
type Entry struct {
	Name  string
	IsDir bool
	Size  int64
}

// itemEnvelope: the per-item shape inside the parent envelope's
// `items` array. We keep the json-vs-Go field mapping in one place
// rather than tagging Entry directly, so the public type stays free
// of wire-format concerns. The full backend envelope also carries
// parent-level Name / Modified / NumDirs / NumFiles fields (see
// cli/cmd/ctl/files/ls.go's listingResponse for the shape `files ls`
// renders) — the walker doesn't need them, so we don't decode them.
//
// `FileSize` is the cloud-drive variant of `size` (see the comment on
// List); we decode both and prefer the populated one in the caller.
type itemEnvelope struct {
	Name     string `json:"name"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	FileSize int64  `json:"fileSize"`
}

// List does GET /api/resources/<encPlainPath>/ and returns the entries
// inside that directory. The trailing slash is enforced internally —
// the backend's FileParam.convert rejects requests with fewer than 3
// '/'-split segments, and the trailing slash is what guarantees that
// invariant for shallow paths like `drive/Home/`.
//
// The envelope includes a `parent` block, but we only consume `items`
// here; callers that need parent metadata should Stat the path
// separately.
//
// Two envelope shapes are accepted:
//
//   - Drive / Sync / Cache / External / Share: children live in
//     `items`, each with `size` (number) and the per-FileInfo metadata.
//   - Cloud drives (awss3 / google / dropbox / tencent): children
//     live in `data` instead, and each child reports its byte count
//     under `fileSize` (the server-side `size` field is also populated
//     on most versions but not all). Listings on these namespaces are
//     what `cat` / the recursive walker hit when the user navigates
//     into a connected cloud bucket.
//
// The decode tolerates the `mode`/`modified` empty-string values that
// cloud listings emit by simply not decoding them — the walker only
// needs (Name, IsDir, Size).
func (c *Client) List(ctx context.Context, plainPath string) ([]Entry, error) {
	if !strings.HasSuffix(plainPath, "/") {
		plainPath += "/"
	}
	endpoint := c.resourcesURL(plainPath)
	body, err := c.do(ctx, http.MethodGet, endpoint, nil, http.Header{
		"Accept": []string{"application/json"},
	})
	if err != nil {
		return nil, err
	}
	var env struct {
		Items []itemEnvelope `json:"items"`
		Data  []itemEnvelope `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("decode listing response: %w", err)
	}
	source := env.Items
	if len(source) == 0 && len(env.Data) > 0 {
		source = env.Data
	}
	out := make([]Entry, 0, len(source))
	for _, it := range source {
		size := it.Size
		if size == 0 && it.FileSize != 0 {
			size = it.FileSize
		}
		out = append(out, Entry{Name: it.Name, IsDir: it.IsDir, Size: size})
	}
	return out, nil
}
