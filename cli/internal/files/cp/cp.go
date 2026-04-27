// Package cp implements the wire side of `olares-cli files cp` and
// `olares-cli files mv`. Both verbs hit the per-user files-backend's
// PATCH /api/paste/<node>/ endpoint with a JSON body of
// {action, source, destination}; action is "copy" or "move", and
// source/destination are plain (decoded) UTF-8 paths shaped like
// `/<fileType>/<extend><subPath>` — the LarePass web app's `from`/`to`
// after decodeURIComponent. See
// apps/packages/app/src/api/files/v2/common/utils.ts pasteAction L60-L109
// (and the per-driver `copy`/`paste` builders in
// apps/packages/app/src/api/files/v2/drive/data.ts) for the source of
// truth on shaping these strings.
//
// The wire endpoint returns one task_id per request, so a multi-source
// `cp src1 src2 dst/` is N PATCH calls (one per source). We keep them
// serial here rather than parallel: a paste task is essentially a
// metadata operation server-side (the actual byte movement runs on
// the files-backend's task queue, and we surface only the task_id),
// so the win from concurrency is small and the failure story
// (partial success across N parallel requests with no transactional
// rollback) is worse.
package cp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Action picks the wire `action` value. The endpoint accepts only
// these two strings — anything else returns a server-side error.
type Action string

const (
	ActionCopy Action = "copy"
	ActionMove Action = "move"
)

// pasteMultiNodeFileTypes mirrors the web app's pasteMutiNodesDriveType
// (see v2/common/utils.ts L55-L58): for these storage classes the
// path's <extend> segment IS the node name, and we use it as the
// {node} URL segment when building /api/paste/<node>/. For all other
// classes the {node} comes from a fallback default — see ResolveNode.
//
// Keep in lockstep with the web app's set; when the backend gains a
// new "nodey" storage class, both sides have to flip together.
var pasteMultiNodeFileTypes = map[string]struct{}{
	"external": {},
	"cache":    {},
}

// Source is one user-supplied source path, normalized so the planner
// has a single canonical shape to validate. Construct via the cobra
// layer's path parser; the planner intentionally doesn't know about
// FrontendPath so this package stays free of cobra/cmdutil deps.
type Source struct {
	// FileType + Extend + SubPath together form the wire-shape source
	// after a "/" prefix and "/" join — e.g. ("drive", "Home",
	// "/Documents/foo.pdf") → "/drive/Home/Documents/foo.pdf".
	FileType string
	Extend   string
	// SubPath always starts with '/'. A trailing '/' is preserved
	// straight through to the wire as the directory marker.
	SubPath string
	// IsDirIntent: did the user signal this is a directory (trailing
	// '/' on the path)? Plan errors out for IsDirIntent=true sources
	// without recursive=true, mirroring `cp`/`mv`'s -r/-R refusal.
	IsDirIntent bool
}

// Destination is the parsed `<dst>` arg, interpreted under the
// "trailing slash means directory" rule that the rest of files-cli
// already uses (upload.go's <remote>, download.go's resolveLocalFile,
// rm.go's IsDirIntent, ...). Keeping the rule consistent across verbs
// means users only have to internalise it once.
type Destination struct {
	FileType    string
	Extend      string
	SubPath     string
	IsDirIntent bool
}

// Op is one PATCH /api/paste/<node>/ call, fully resolved. Source and
// Destination are the final, decoded, plain UTF-8 strings that go in
// the JSON body — already prefixed with `/<fileType>/<extend>` and
// carrying the directory trailing-slash where relevant.
type Op struct {
	Action      Action
	Source      string
	Destination string
	// IsDir is the directory hint — copied through from the source's
	// IsDirIntent so log lines / future progress tracking can branch
	// without re-parsing Source.
	IsDir bool
	// Node is the {node} URL segment for /api/paste/<node>/, resolved
	// per Op (External/Cache contribute their Extend, others fall
	// back to the caller-supplied default). See ResolveNode for the
	// cascade.
	Node string
}

// Client is the per-FilesURL handle for paste calls. HTTPClient is
// expected to be the factory-provided client whose refreshingTransport
// injects `X-Authorization` (not `Authorization: Bearer`, see
// pkg/cmdutil/factory.go for why) and refreshes on 401/403.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// HTTPError carries the status + truncated body of a non-2xx response
// so the cobra layer can branch on the status code (e.g. to give a
// friendly "not found" or auth-issue CTA). Same shape as the
// per-package HTTP errors in upload / rm / download to keep the error
// contract uniform.
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

// Plan turns (sources, destination, action, recursive, default node)
// into the list of Ops to PATCH. Validation is Unix `cp`/`mv`-style:
//
//   - Source.SubPath == "/" → "refusing to copy/move the root of <fileType>/<extend>"
//     (paste-the-whole-volume is not a meaningful operation through this
//     endpoint and the consequences of doing it accidentally are large).
//   - Source.IsDirIntent without recursive=true → "is a directory:
//     pass -r/-R" (matches Unix `cp` / `mv` -r refusal).
//   - len(srcs) > 1 with !Destination.IsDirIntent → "target must be a
//     directory (end with '/') when copying more than one source".
//   - Source resolves identical to Destination → "source and
//     destination are the same" (no-op, almost certainly a typo).
//   - Destination is inside a Source (cycle) → reject ("cannot copy
//     a directory into itself").
//
// Resolution of the per-Op Destination wire string:
//   - dst.IsDirIntent (drop-into-dir mode): each src's basename is
//     appended to the dst subpath (with the trailing '/' for dir
//     sources). This is the "cp foo bar/" UX.
//   - !dst.IsDirIntent (exact-target / rename mode): dst.SubPath is
//     used verbatim, with a trailing '/' synthesised when the source
//     is a directory (so the wire shape stays consistent).
//
// Per-Op node resolution follows the web app's
// `dst_node || src_node || default` cascade — see ResolveNode for the
// detailed rules. flagNode (when non-empty) overrides every Op's
// resolved node, mirroring `--node` in `files upload`.
func Plan(
	srcs []Source,
	dst Destination,
	action Action,
	recursive bool,
	defaultNode string,
	flagNode string,
) ([]Op, error) {
	if len(srcs) == 0 {
		return nil, errors.New("cp: no sources supplied")
	}
	if action != ActionCopy && action != ActionMove {
		return nil, fmt.Errorf("cp: invalid action %q (want %q or %q)",
			action, ActionCopy, ActionMove)
	}

	// Single-source pre-checks. We surface a clear "is a dir without
	// -r" message before the multi-source branch so users get the
	// most actionable error first.
	for _, s := range srcs {
		if strings.Trim(s.SubPath, "/") == "" {
			return nil, fmt.Errorf(
				"refusing to %s the root of %s/%s",
				action, s.FileType, s.Extend)
		}
		if s.IsDirIntent && !recursive {
			return nil, fmt.Errorf(
				"%s/%s%s is a directory: pass -r/-R to %s it recursively",
				s.FileType, s.Extend, s.SubPath, action)
		}
	}
	if len(srcs) > 1 && !dst.IsDirIntent {
		return nil, fmt.Errorf(
			"%s: target %q must end with '/' when more than one source is given (got %d sources)",
			action, dst.FileType+"/"+dst.Extend+dst.SubPath, len(srcs))
	}

	ops := make([]Op, 0, len(srcs))
	for _, s := range srcs {
		op, err := planOne(s, dst, action, defaultNode, flagNode)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

// planOne builds the wire shape + node for a single (src, dst) pair
// after Plan-level validation has already trimmed the obvious errors.
// Kept private so callers go through Plan and thus pick up the
// multi-source / recursive checks; planOne by itself is not safe to
// call on raw user input.
func planOne(s Source, dst Destination, action Action, defaultNode, flagNode string) (Op, error) {
	srcWire := buildWire(s.FileType, s.Extend, s.SubPath, s.IsDirIntent)

	var dstWire string
	if dst.IsDirIntent {
		base := lastSegment(s.SubPath)
		if base == "" {
			// Defensive — Plan already rejected root sources, but
			// guard the index so any future regression there shows up
			// as a typed error rather than producing a malformed
			// wire path.
			return Op{}, fmt.Errorf("internal: cannot derive basename from source %q", s.SubPath)
		}
		parent := dst.SubPath
		if !strings.HasSuffix(parent, "/") {
			parent += "/"
		}
		dstWire = "/" + dst.FileType + "/" + dst.Extend + parent + base
		if s.IsDirIntent {
			dstWire += "/"
		}
	} else {
		// Exact-target / rename mode. dst.SubPath is used verbatim;
		// only synthesise a trailing '/' when the source is a dir so
		// the wire shape matches what the backend expects for
		// "copy/move a directory tree" (see PosixStorage.Paste in
		// files/pkg/drivers/posix/posix/posix.go for the dir-vs-file
		// branch).
		dstWire = "/" + dst.FileType + "/" + dst.Extend + dst.SubPath
		if s.IsDirIntent && !strings.HasSuffix(dstWire, "/") {
			dstWire += "/"
		}
	}

	if srcWire == dstWire {
		return Op{}, fmt.Errorf(
			"%s: source and destination are the same (%s); nothing to do",
			action, srcWire)
	}
	// Cycle check: dst inside src (e.g. cp -r /a /a/sub). The cheap
	// way is to anchor src with a trailing '/' so a substring like
	// "/a/sub" is required (not "/abc"); both sides already use the
	// directory marker for dir paths so the prefix test is reliable.
	if s.IsDirIntent {
		anchor := srcWire
		if !strings.HasSuffix(anchor, "/") {
			anchor += "/"
		}
		if strings.HasPrefix(dstWire, anchor) {
			return Op{}, fmt.Errorf(
				"%s: destination %s is inside source %s (would create a cycle)",
				action, dstWire, srcWire)
		}
	}

	node := ResolveNode(s.FileType, s.Extend, dst.FileType, dst.Extend, defaultNode, flagNode)
	if node == "" {
		return Op{}, fmt.Errorf(
			"%s: cannot resolve {node} URL segment (no --node, no External/Cache hint, and the default is empty); pass --node to override",
			action)
	}

	return Op{
		Action:      action,
		Source:      srcWire,
		Destination: dstWire,
		IsDir:       s.IsDirIntent,
		Node:        node,
	}, nil
}

// ResolveNode reproduces the web app's `dst_node || src_node ||
// fallback` cascade for the {node} URL segment of /api/paste/<node>/:
//
//   - flagNode (when non-empty) wins outright. This is the `--node`
//     CLI override and matches `files upload`'s behavior.
//   - For External/Cache fileTypes the path's Extend IS the node name
//     (see pasteMultiNodeFileTypes). Destination wins over source —
//     that's the order the web app uses (see pasteAction in
//     v2/common/utils.ts: `item.dst_node || item.src_node || ...`).
//   - Otherwise fall back to the caller-supplied defaultNode (typically
//     the first entry from /api/nodes/, the same default `files upload`
//     uses).
//
// Returning "" signals "no node could be resolved"; callers should
// surface that as a clear error, not silently send it.
func ResolveNode(srcFileType, srcExtend, dstFileType, dstExtend, defaultNode, flagNode string) string {
	if flagNode != "" {
		return flagNode
	}
	if _, ok := pasteMultiNodeFileTypes[dstFileType]; ok && dstExtend != "" {
		return dstExtend
	}
	if _, ok := pasteMultiNodeFileTypes[srcFileType]; ok && srcExtend != "" {
		return srcExtend
	}
	return defaultNode
}

// buildWire assembles the LarePass-shaped wire path
// "/<fileType>/<extend><subPath>" — note that subPath already starts
// with '/', so we do NOT add another one between extend and subPath.
// The trailing '/' for directories is preserved from subPath if
// present; we add one defensively when isDir says so but the caller's
// subPath is missing it (e.g. a user typed `drive/Home/dir` with
// --recursive flag set on cp -r).
func buildWire(fileType, extend, subPath string, isDir bool) string {
	if !strings.HasPrefix(subPath, "/") {
		subPath = "/" + subPath
	}
	w := "/" + fileType + "/" + extend + subPath
	if isDir && !strings.HasSuffix(w, "/") {
		w += "/"
	}
	return w
}

// lastSegment returns the basename of a slash-separated path, ignoring
// any leading or trailing '/'. Mirrors the web app's behavior when it
// does `appendPath(parentPath, encodeUrl(element.name), ...)` — the
// `name` field is always set to the source's basename.
func lastSegment(sub string) string {
	s := strings.Trim(sub, "/")
	if s == "" {
		return ""
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// pasteRequestBody is the JSON body the files-backend's PATCH /api/paste
// endpoint binds to. The `code: -1` response branch (see web app's
// pasteAction L92-L99: `if (res.data.code === -1)`) signals a
// server-side rejection — typically a malformed path like a literal
// backslash that confuses the backend's path parser. We surface that
// as a clean error rather than letting it look like a successful empty
// response.
type pasteRequestBody struct {
	Action      Action `json:"action"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// pasteResponseEnvelope is the success/failure shape returned by
// /api/paste/<node>/. The web app reads `data.task_id` on success and
// branches on `data.code === -1` for the malformed-path failure mode.
type pasteResponseEnvelope struct {
	Code    *int   `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
}

// PasteOne sends one PATCH /api/paste/<node>/ for the supplied Op and
// returns the resulting task_id. The actual byte movement happens
// asynchronously on the files-backend's task queue — by the time this
// returns we only know "the server has queued the task", not "the
// task has finished". Callers that need completion semantics have to
// poll a separate task-status endpoint (not implemented here, kept as
// a future iteration).
//
// Errors:
//   - non-2xx response → *HTTPError (status / body preserved for
//     friendly reformatting in the cobra layer).
//   - 2xx with `code: -1` body → fmt.Errorf("server rejected ...")
//     so the malformed-path failure mode doesn't masquerade as
//     success.
//   - 2xx without a task_id → fmt.Errorf("server returned no task_id"),
//     same rationale: a "queued" answer with no handle is useless.
func (c *Client) PasteOne(ctx context.Context, op Op) (string, error) {
	if op.Node == "" {
		return "", fmt.Errorf("PasteOne: empty Node (Plan should have rejected this)")
	}
	endpoint := c.BaseURL + "/api/paste/" + url.PathEscape(op.Node) + "/"

	body, err := json.Marshal(pasteRequestBody{
		Action:      op.Action,
		Source:      op.Source,
		Destination: op.Destination,
	})
	if err != nil {
		return "", fmt.Errorf("marshal paste body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", &HTTPError{
			Status: resp.StatusCode,
			Body:   string(respBody),
			URL:    endpoint,
			Method: http.MethodPatch,
		}
	}

	// The web app peels one layer off the axios envelope (`res.data`)
	// before reading task_id. Our http.Client returns the raw HTTP
	// body, which IS that one layer — i.e. the JSON we read here is
	// the same shape as the web app's `res.data`. The decoder is lax
	// (UnknownFields ignored) so future server fields don't break us.
	var env pasteResponseEnvelope
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &env); err != nil {
			return "", fmt.Errorf("decode paste response: %w (body=%s)", err, truncateBody(respBody))
		}
	}
	if env.Code != nil && *env.Code == -1 {
		// Same code path the web app surfaces as "files.backslash_upload":
		// the server has rejected the path shape, typically because of
		// a backslash or other unsupported character. Don't pretend
		// this succeeded.
		msg := env.Message
		if msg == "" {
			msg = "server rejected the request (code -1); check for unsupported characters (e.g. backslash) in the source/destination path"
		}
		return "", fmt.Errorf("paste %s → %s: %s", op.Source, op.Destination, msg)
	}
	if env.TaskID == "" {
		return "", fmt.Errorf("paste %s → %s: server returned no task_id (body=%s)",
			op.Source, op.Destination, truncateBody(respBody))
	}
	return env.TaskID, nil
}

// Node mirrors the upload package's projection of the files-backend
// `FileNode`. We re-declare it here rather than depending on
// internal/files/upload so the cp package keeps its own surface
// minimal — the only field we actually consume is Name.
type Node struct {
	Name   string `json:"name"`
	Master bool   `json:"master"`
}

// nodesEnvelope is the {data: {nodes: [...]}} response shape returned
// by GET /api/nodes/. Same envelope as upload.Client.FetchNodes uses;
// see fetchNodeList in v2/common/utils.ts L320-L329 for the source of
// truth on the shape.
type nodesEnvelope struct {
	Data struct {
		Nodes []Node `json:"nodes"`
	} `json:"data"`
}

// FetchNodes calls GET {filesURL}/api/nodes/ and returns the configured
// Drive nodes. Used to pick the default {node} URL segment for paste
// calls when the user didn't pass --node and neither side of the
// operation lives on External/Cache. Same behavior as
// upload.Client.FetchNodes; we duplicate rather than import the upload
// package to avoid a dependency that exists only for two helper
// methods.
func (c *Client) FetchNodes(ctx context.Context) ([]Node, error) {
	endpoint := c.BaseURL + "/api/nodes/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(body),
			URL:    endpoint,
			Method: http.MethodGet,
		}
	}
	var env nodesEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("decode /api/nodes/ response: %w (body=%s)", err, truncateBody(body))
	}
	if len(env.Data.Nodes) == 0 {
		return nil, errors.New("files-backend returned no Drive nodes; cannot resolve default {node} for paste")
	}
	return env.Data.Nodes, nil
}

// truncateBody is the same helper rm/upload use to keep error
// messages bounded — paste responses are tiny in practice but a
// misconfigured server can send back an HTML error page that we
// don't want to splat across the user's terminal.
func truncateBody(b []byte) string {
	if len(b) <= 200 {
		return string(b)
	}
	return string(b[:200]) + "...(truncated)"
}
