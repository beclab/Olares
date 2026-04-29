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
	"time"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Client is the per-FilesURL handle used by uploader.go and the cobra
// command. It is cheap to construct; reuse one per `files upload`
// invocation.
//
// HTTPClient is expected to be a factory-provided client whose
// refreshingTransport injects `X-Authorization` (not `Authorization:
// Bearer`, see pkg/cmdutil/factory.go for why) and transparently refreshes
// the token on 401/403 — except for chunk-streaming requests whose body
// is a *os.File (req.GetBody == nil), where retry is impossible and the
// 401 falls through to the caller.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string // FilesURL, e.g. https://files.alice.olares.com
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

// Mkdir POSTs an empty body to /api/resources/<encoded fullPath>/
// to create a directory under the selected namespace root. The trailing slash is what the
// backend uses to discriminate "create directory" from "create empty file"
// (postCreateFile in v2/common/utils.ts does the same thing — `isDir
// ? '/' : ”`).
//
// `fullPath` is the absolute frontend path (e.g. `/drive/Home/Documents`
// or `/sync/<repo_id>/docs`) without the `/api/resources` prefix.
//
// IMPORTANT: This call is NOT idempotent on the server side. The
// files-backend auto-renames colliding directories ("Documents" exists
// → POST creates "Documents (1)") instead of returning 409. We treat
// the 409 path as "already exists" for completeness but most servers
// won't take that branch — callers should reserve Mkdir for paths
// they're confident don't exist yet (e.g. brand-new subdirectories
// they computed from a local walk). The 409 fast-path stays in case
// some deployments do return 409 for collisions.
func (c *Client) Mkdir(ctx context.Context, fullPath string) error {
	clean := strings.Trim(fullPath, "/")
	if clean == "" {
		// Root always exists; nothing to do.
		return nil
	}
	encoded := encodepath.EncodeURL(clean)
	endpoint := c.BaseURL + "/api/resources/" + encoded + "/"
	_, err := c.do(ctx, http.MethodPost, endpoint, nil, nil, "")
	if err != nil {
		var hErr *HTTPError
		if errors.As(err, &hErr) && hErr.Status == http.StatusConflict {
			// Directory already exists — exactly what we wanted.
			return nil
		}
		return fmt.Errorf("mkdir %q: %w", fullPath, err)
	}
	return nil
}

// CreateEmptyFile POSTs an empty body to
// /api/resources/<encoded fullPath> (no trailing slash) to
// materialize a zero-length file. The web app routes empty files through
// uploadEmptyFile() instead of the chunk pipeline (resumable.js cannot
// represent a 0-byte chunk), and we mirror that here.
//
// Unlike Mkdir, a 409 here is reported back to the caller — we don't
// silently overwrite or pretend success when the user explicitly asked
// to upload a file and a name collision happened.
func (c *Client) CreateEmptyFile(ctx context.Context, fullPath string) error {
	clean := strings.Trim(fullPath, "/")
	if clean == "" {
		return fmt.Errorf("CreateEmptyFile: empty path")
	}
	encoded := encodepath.EncodeURL(clean)
	endpoint := c.BaseURL + "/api/resources/" + encoded
	_, err := c.do(ctx, http.MethodPost, endpoint, nil, nil, "")
	if err != nil {
		return fmt.Errorf("create empty file %q: %w", fullPath, err)
	}
	return nil
}

// --- Cloud-transfer task polling (stage 2 of cloud-drive uploads) ---
//
// Cloud drive uploads (awss3 / google / dropbox) are a two-stage
// operation: stage 1 is the regular chunked POST to the Olares files-
// backend (covered by UploadFile), and stage 2 is a server-side
// transfer task that copies the staged file from Olares-internal
// storage to the user's actual cloud bucket. The taskID for stage 2
// is returned in the FINAL stage-1 chunk's response body (see
// parseFinalChunkTaskID in uploader.go), and the client drives
// stage 2 by polling /api/task/<node>/?task_id=<id> until the status
// reaches a terminal value.
//
// This mirrors apps/packages/app/src/services/olaresTask/index.ts —
// the web app's Taskmanager uses the same endpoint + the same status
// vocabulary (pending/running/completed/failed/canceled/cancelled/paused).

// CloudTaskStatus values mirror OlaresTaskStatus from
// services/abstractions/olaresTask/interface.ts. The server returns
// the literal lowercase strings, so we keep them as untyped string
// constants (vs. a typed enum) — there's no validation; an unknown
// status is treated as "still in flight" and the loop polls again.
const (
	cloudTaskStatusCompleted = "completed"
	cloudTaskStatusFailed    = "failed"
	cloudTaskStatusCanceled  = "canceled"
	cloudTaskStatusCancelled = "cancelled" // server uses both spellings
)

// DefaultCloudTaskPollInterval is how often WaitCloudTask polls the
// task-status endpoint when the caller doesn't override it. 2s is a
// conservative compromise between responsiveness (cloud uploads of
// small files can finish in <5s end-to-end) and not flooding the
// server with status checks for big uploads.
const DefaultCloudTaskPollInterval = 2 * time.Second

// CloudTaskUpdate is the per-poll snapshot WaitCloudTask passes to
// its onUpdate callback. The cobra layer renders progress lines
// from these without having to know about the JSON envelope shape.
type CloudTaskUpdate struct {
	Status        string  // raw server status: pending / running / paused / ...
	Progress      float64 // 0..100 (server-reported, may stay at 0 for short tasks)
	CurrentPhase  int     // 1..TotalPhase, useful when the server splits the transfer in stages
	TotalPhase    int
	TotalFileSize int64
	FailedReason  string
}

// CloudTaskUpdateFunc is invoked once per poll while the task is
// still in flight (pending / running / paused / unknown). It is NOT
// invoked for the terminal status — that arrives via WaitCloudTask's
// return.
type CloudTaskUpdateFunc func(CloudTaskUpdate)

// taskQueryEnvelope is the JSON shape returned by GET
// /api/task/<node>/?task_id=<id>. The web app reads `task.status` /
// `task.progress` etc. straight off this envelope (see
// olaresTask/index.ts getTask + cloudUpload/setQueryResult). We mirror
// the field set conservatively — only fields we surface in
// CloudTaskUpdate or the failure-reason error are decoded.
type taskQueryEnvelope struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Task struct {
		ID            string  `json:"id"`
		Status        string  `json:"status"`
		Progress      float64 `json:"progress"`
		CurrentPhase  int     `json:"current_phase"`
		TotalPhase    int     `json:"total_phase"`
		TotalFileSize int64   `json:"total_file_size"`
		FailedReason  string  `json:"failed_reason,omitempty"`
	} `json:"task"`
}

// WaitCloudTask polls /api/task/<node>/?task_id=<taskID> at
// `interval` (or DefaultCloudTaskPollInterval if interval == 0) and
// returns when the task reaches a terminal status:
//
//   - completed                  → nil
//   - failed                     → fmt.Errorf with `failed_reason`
//     when the server provided
//     one, otherwise a generic
//     "task failed" message
//   - canceled / cancelled       → fmt.Errorf("task ... was cancelled")
//
// `onUpdate` is called once per poll whenever the task is NOT yet
// terminal — pass nil if the caller doesn't want progress updates.
//
// ctx cancellation is honored promptly between polls (and at every
// HTTP request via Client.do). Transient HTTP errors during polling
// (the task endpoint flapping, an in-cluster service redeploy)
// surface immediately as errors — we don't paper over them, because
// a long-running cloud transfer that can't be queried is
// indistinguishable from a stuck transfer; the caller should bubble
// up the failure.
func (c *Client) WaitCloudTask(
	ctx context.Context,
	node, taskID string,
	interval time.Duration,
	onUpdate CloudTaskUpdateFunc,
) error {
	if taskID == "" {
		return errors.New("WaitCloudTask: empty taskID")
	}
	if node == "" {
		return errors.New("WaitCloudTask: empty node")
	}
	if interval <= 0 {
		interval = DefaultCloudTaskPollInterval
	}

	q := url.Values{}
	q.Set("task_id", taskID)
	endpoint := c.BaseURL + "/api/task/" + url.PathEscape(node) + "/?" + q.Encode()

	for {
		body, err := c.do(ctx, http.MethodGet, endpoint, nil, nil, "")
		if err != nil {
			return fmt.Errorf("query cloud task %s on node %s: %w", taskID, node, err)
		}
		var env taskQueryEnvelope
		if len(body) > 0 {
			if err := json.Unmarshal(body, &env); err != nil {
				return fmt.Errorf("decode task query response for %s: %w (body=%s)",
					taskID, err, truncateBody(body))
			}
		}

		switch env.Task.Status {
		case cloudTaskStatusCompleted:
			return nil
		case cloudTaskStatusFailed:
			reason := env.Task.FailedReason
			if reason == "" {
				reason = "server reported failure with no failed_reason"
			}
			return fmt.Errorf("cloud transfer task %s failed: %s", taskID, reason)
		case cloudTaskStatusCanceled, cloudTaskStatusCancelled:
			return fmt.Errorf("cloud transfer task %s was cancelled server-side", taskID)
		}

		if onUpdate != nil {
			onUpdate(CloudTaskUpdate{
				Status:        env.Task.Status,
				Progress:      env.Task.Progress,
				CurrentPhase:  env.Task.CurrentPhase,
				TotalPhase:    env.Task.TotalPhase,
				TotalFileSize: env.Task.TotalFileSize,
				FailedReason:  env.Task.FailedReason,
			})
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
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
