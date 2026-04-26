// api.go: thin HTTP client for the per-user files-backend's Drive v2
// upload-related endpoints. The wire surface mirrors what
// apps/packages/app/src/api/files/v2/drive/data.ts (getFileServerUploadLink,
// getFileUploadedBytes) and apps/packages/app/src/api/files/v2/drive/utils.ts
// (createDir / postCreateFile) call from the web app, so probe / resume /
// chunk POST behavior stay byte-compatible across both clients.
//
// The HTTP client is supplied by the caller (so tests can use httptest)
// and the access token is passed in via X-Authorization on every request
// (same convention as the rest of olares-cli — see
// pkg/cmdutil/factory.go's authTransport for the rationale).
package upload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/beclab/Olares/cli/pkg/files/encodepath"
)

// Client is the per-FilesURL handle used by uploader.go and the cobra
// command. It is cheap to construct; reuse one per `files upload`
// invocation.
//
// AccessToken is sent as `X-Authorization` (not `Authorization: Bearer`),
// because Olares' edge stack only forwards the X-Authorization header to
// per-user services. See pkg/cmdutil/factory.go for the full rationale.
type Client struct {
	HTTPClient  *http.Client
	BaseURL     string // FilesURL, e.g. https://files.alice.olares.com
	AccessToken string
}

// Node is the projection of files-backend's `FileNode` that we actually
// use. The full struct also has a `master` boolean (see
// apps/packages/app/src/stores/files.ts), which we don't need for the
// upload flow — we just take the first node's name as the path segment
// for /upload/upload-link/{node}/ and /upload/file-uploaded-bytes/{node}/.
type Node struct {
	Name   string `json:"name"`
	Master bool   `json:"master"`
}

// nodesEnvelope mirrors the {data: {nodes: [...]}} response shape that
// fetchNodeList in apps/packages/app/src/api/files/v2/common/utils.ts
// L320-L329 unpacks. We keep the envelope local to this package since
// it's not useful elsewhere.
type nodesEnvelope struct {
	Data struct {
		Nodes []Node `json:"nodes"`
	} `json:"data"`
}

// FetchNodes calls GET {filesURL}/api/nodes/ and returns the configured
// Drive nodes. The CLI uses nodes[0].Name (or a user-supplied --node
// override) as the per-request `{node}` path segment for the upload
// endpoints — same convention as the web app's getUploadNode().
//
// Errors:
//   - any non-2xx status surfaces as fmt.Errorf with status + body
//   - empty `data.nodes` is reported with a clear message; the upload
//     flow can't proceed without a node identifier.
func (c *Client) FetchNodes(ctx context.Context) ([]Node, error) {
	endpoint := c.BaseURL + "/api/nodes/"
	body, err := c.do(ctx, http.MethodGet, endpoint, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", endpoint, err)
	}
	var env nodesEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("decode /api/nodes/ response: %w (body=%s)", err, truncateBody(body))
	}
	if len(env.Data.Nodes) == 0 {
		return nil, errors.New("files-backend returned no Drive nodes; cannot upload")
	}
	return env.Data.Nodes, nil
}

// GetUploadLink calls GET {filesURL}/upload/upload-link/{node}/?file_path=<enc(parentDir)>&from=web.
// The server replies with a plaintext path (e.g.
// `/seafhttp/upload-aj/<repo>/?...`) that the browser then POSTs chunks
// to. The web app appends `?ret-json=1` to that path so the per-chunk
// response is JSON instead of a redirect — we do the same to match.
//
// `parentDir` is the parent directory path WITH the `/drive/Home/...`
// prefix and a TRAILING slash (e.g. `/drive/Home/Documents/`). That's
// what the web app passes through `files.formatPathtoUrl` →
// `path.pathname` before plumbing it into this call (see
// apps/packages/app/src/utils/resumejs.ts L412-L416).
//
// Returned string is a relative path (no scheme/host); the chunk POST
// uses `c.BaseURL + uploadLink` as the target.
func (c *Client) GetUploadLink(ctx context.Context, node, parentDir string) (string, error) {
	endpoint := c.BaseURL +
		"/upload/upload-link/" + url.PathEscape(node) +
		"/?file_path=" + encodepath.EncodeURIComponent(parentDir) +
		"&from=web"
	body, err := c.do(ctx, http.MethodGet, endpoint, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", endpoint, err)
	}
	link := strings.TrimSpace(string(body))
	if link == "" {
		return "", fmt.Errorf("upload-link response is empty (parent_dir=%q)", parentDir)
	}
	// Append ?ret-json=1 the same way the web app does (resumejs.ts /
	// data.ts), so per-chunk responses come back as JSON instead of
	// the default redirect that the browser-style upload would follow.
	if strings.Contains(link, "?") {
		link += "&ret-json=1"
	} else {
		link += "?ret-json=1"
	}
	return link, nil
}

// uploadedBytesEnvelope is the JSON shape returned by
// /upload/file-uploaded-bytes/. The web app reads `uploadedBytes` and
// floors `uploadedBytes / chunkSize` to find the next chunk to send.
type uploadedBytesEnvelope struct {
	UploadedBytes int64 `json:"uploadedBytes"`
}

// GetUploadedBytes asks the server how many bytes of `<parentDir>/<filename>`
// have already been received. New / never-seen files return 0 (or an error
// the web app silently swallows — see resumejs.ts: any non-2xx is treated
// as "start from scratch"). We adopt the same lenient policy so a fresh
// upload doesn't fail just because the server doesn't know about the file
// yet.
//
// `parentDir` follows the same convention as GetUploadLink: full
// `/drive/Home/...` path with a trailing slash. `filename` is the bare
// basename (no directory components).
func (c *Client) GetUploadedBytes(ctx context.Context, node, parentDir, filename string) (int64, error) {
	q := url.Values{}
	q.Set("parent_dir", parentDir)
	q.Set("file_name", filename)
	endpoint := c.BaseURL +
		"/upload/file-uploaded-bytes/" + url.PathEscape(node) +
		"/?" + q.Encode()
	body, err := c.do(ctx, http.MethodGet, endpoint, nil, nil, "")
	if err != nil {
		// Match the web app's silent fallback: probe failures (file
		// doesn't exist, transient 404, ...) all collapse to "we
		// haven't uploaded anything yet".
		return 0, nil //nolint:nilerr // intentional: see docstring
	}
	var env uploadedBytesEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return 0, nil //nolint:nilerr // best-effort probe; restart from 0
	}
	if env.UploadedBytes < 0 {
		return 0, nil
	}
	return env.UploadedBytes, nil
}

// Mkdir POSTs an empty body to /api/resources/drive/Home/<encoded relSubPath>/
// to create a directory under Drive/Home. The trailing slash is what the
// backend uses to discriminate "create directory" from "create empty file"
// (postCreateFile in v2/common/utils.ts does the same thing — `isDir
// ? '/' : ”`).
//
// `relSubPath` is the directory path RELATIVE to /Home (e.g. "Documents"
// or "Documents/photos"); it should NOT include leading or trailing
// slashes — Mkdir handles slash placement and percent-encoding so the
// caller can hand in plain UTF-8 segments.
//
// IMPORTANT: This call is NOT idempotent on the server side. The
// files-backend auto-renames colliding directories ("Documents" exists
// → POST creates "Documents (1)") instead of returning 409. We treat
// the 409 path as "already exists" for completeness but most servers
// won't take that branch — callers should reserve Mkdir for paths
// they're confident don't exist yet (e.g. brand-new subdirectories
// they computed from a local walk). The 409 fast-path stays in case
// some deployments do return 409 for collisions.
func (c *Client) Mkdir(ctx context.Context, relSubPath string) error {
	clean := strings.Trim(relSubPath, "/")
	if clean == "" {
		// Drive/Home root always exists; nothing to do.
		return nil
	}
	encoded := encodepath.EncodeURL(clean)
	endpoint := c.BaseURL + "/api/resources/drive/Home/" + encoded + "/"
	_, err := c.do(ctx, http.MethodPost, endpoint, nil, nil, "")
	if err != nil {
		var hErr *HTTPError
		if errors.As(err, &hErr) && hErr.Status == http.StatusConflict {
			// Directory already exists — exactly what we wanted.
			return nil
		}
		return fmt.Errorf("mkdir %q: %w", relSubPath, err)
	}
	return nil
}

// CreateEmptyFile POSTs an empty body to
// /api/resources/drive/Home/<encoded relPath> (no trailing slash) to
// materialize a zero-length file. The web app routes empty files through
// uploadEmptyFile() instead of the chunk pipeline (resumable.js cannot
// represent a 0-byte chunk), and we mirror that here.
//
// Unlike Mkdir, a 409 here is reported back to the caller — we don't
// silently overwrite or pretend success when the user explicitly asked
// to upload a file and a name collision happened.
func (c *Client) CreateEmptyFile(ctx context.Context, relPath string) error {
	clean := strings.Trim(relPath, "/")
	if clean == "" {
		return fmt.Errorf("CreateEmptyFile: empty path")
	}
	encoded := encodepath.EncodeURL(clean)
	endpoint := c.BaseURL + "/api/resources/drive/Home/" + encoded
	_, err := c.do(ctx, http.MethodPost, endpoint, nil, nil, "")
	if err != nil {
		return fmt.Errorf("create empty file %q: %w", relPath, err)
	}
	return nil
}

// HTTPError carries the status + truncated body of a non-2xx response so
// callers that care (Mkdir's 409 fast-path, the chunk uploader's permanent
// vs. retryable classification) can branch on the status code without
// stringly-typed error parsing.
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

// do performs a single HTTP request with the configured access token
// injected as `X-Authorization`, and returns the response body on 2xx.
// Non-2xx responses surface as *HTTPError so callers can branch on
// status (409 idempotent for Mkdir, the permanent/retryable split for
// the chunk uploader).
//
// `body` may be nil. `contentType` is honored only when non-empty. The
// extra `headers` map lets callers add chunk-POST headers
// (Content-Range / Content-Disposition) without bypassing the
// X-Authorization injection.
func (c *Client) do(
	ctx context.Context,
	method, endpoint string,
	body io.Reader,
	headers http.Header,
	contentType string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if c.AccessToken != "" {
		req.Header.Set("X-Authorization", c.AccessToken)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(respBody),
			URL:    endpoint,
			Method: method,
		}
	}
	return respBody, nil
}

func truncateBody(b []byte) string {
	if len(b) <= 200 {
		return string(b)
	}
	return string(b[:200]) + "...(truncated)"
}
