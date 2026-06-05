// Package smbmount implements the wire side of `olares-cli files smb`
// AND `olares-cli files nfs` — the CLI counterpart of the LarePass web
// app's "Connect to Server" flow for mounting an external SMB share or
// NFS export into the per-user files-backend's `external/<node>/...`
// namespace. SMB and NFS share the mount / unmount / favorites wire
// surface (`/api/mount`, `/api/unmount`, `/api/smb_history`) and differ
// only in the `?external_type=` switch and the request body shape
// (SMB: {smbPath,user,password}; NFS: {url[,operate:"list"]}, no
// credentials) — see Mount vs. MountNFS below.
//
// Source of truth on the wire (and the rationale behind every
// quirk this package handles):
//
//   - Mount  — POST /api/mount/<node>/?external_type=smb
//     body:  {smbPath, user, password}
//     reply: {code, message, data}
//       code == 200 → mounted successfully; the share is now visible
//                     under external/<node>/<entry-name>/
//       code == 300 → smbPath was a host-only address (e.g. //host)
//                     and the server is asking the user to pick one
//                     of the shares it discovered. `data` is an array
//                     of {path:"//host/share[/sub]", ...}; the GUI
//                     pops a chooser dialog (ConnectServerPath.vue)
//                     and re-mounts with the picked path.
//       any other   → server-side error; surface message verbatim.
//
//     LarePass call site: stores/files.ts L1263-L1302 mountSmbInExternal,
//     used by ConnectServerStep1.vue (when both creds are saved on a
//     favorite entry) and ConnectServerStep3.vue (after the user types
//     them in).
//
//   - Unmount — POST /api/unmount/<fileType>/<fileExtend>/<name>/?external_type=<type>
//     body:  {} (the path/query string carries every parameter)
//     reply: {code, message}
//       code == 200 → unmounted; the entry disappears from the
//                     external/<node>/ listing.
//       any other   → server-side error.
//
//     LarePass call site: stores/operation.ts L844-L887, triggered by
//     the right-click "Unmount" menu item on an external mount entry.
//     Note: the GUI uses POST (not DELETE) — same here, byte-for-byte.
//
//   - History — /api/smb_history/<node>/
//     GET    → array of {url, username?, password?, timestamp?} records
//              (no envelope; body is the array directly).
//     PUT    body: array of records to upsert (each {url, [username,
//              password]}).
//     DELETE body: array of {url} entries to drop.
//
//     LarePass call sites: ConnectServerStep1.vue L124-L140 (favorite
//     add / remove), ConnectServerStep3.vue L111-L113 (auto-save
//     credentials after a successful mount). The "favorite" UI is
//     a per-node history of previously-typed SMB addresses, with
//     optional saved credentials.
//
// FetchNodes mirrors the cp / upload counterparts byte-for-byte
// (envelope `{data:{nodes:[...]}}`, take nodes[0].Name as the
// default {node}). Duplicating it across packages — same as cp and
// upload do today — keeps the package independent and avoids a
// circular import.
package smbmount

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

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Node is one row of GET /api/nodes/. Same shape as cp.Node /
// upload.Node — duplicated locally so this package doesn't pull in
// cp's wire surface.
type Node struct {
	Name   string `json:"name"`
	Master bool   `json:"master,omitempty"`
}

// HistoryEntry is one row of GET /api/smb_history/<node>/.
//
// Username / Password are optional — the LarePass UI lets a user
// save just a URL ("favorite") and prompt for credentials each
// time, OR save the full triple ("seamless reconnect"). The
// timestamp is server-supplied and we round-trip it verbatim so a
// PUT that updates an existing row keeps the original creation
// time.
type HistoryEntry struct {
	URL       string `json:"url"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// MountOptions captures the body of POST /api/mount/<node>/?external_type=smb.
//
// Field names on the wire are `smbPath`, `user`, `password` (lower-
// camel for `smbPath`, snake-cased "user" rather than "username") —
// exactly what the GUI's mountSmbInExternal sends. Going through a
// typed struct rather than a freeform map keeps the JSON tags pinned
// at the package boundary.
type MountOptions struct {
	SMBPath  string `json:"smbPath"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// MountResult is the parsed body of a mount POST. The server returns
// {code, message, data} where:
//
//   - code 200 ⇒ the SMB share mounted; `data` is metadata the GUI
//     ignores (we ignore it too — the user just needs to know it
//     worked, and `files ls external/<node>/` is the canonical way
//     to confirm).
//   - code 300 ⇒ `smbPath` was a host-only address; `data` is
//     `[{path:"//host/share[/sub]", ...}, ...]`, the list of
//     discovered shares. The caller surfaces these so the user can
//     re-run with one of them.
//
// We model both branches in one struct so the cobra layer can switch
// on Code without re-decoding.
type MountResult struct {
	Code    int      // 200 = success; 300 = pick-a-share required
	Message string   // server-supplied diagnostic, surfaced verbatim
	Paths   []string // populated only when Code == 300
}

// HTTPError carries the status + truncated body of a non-2xx
// response. Same shape as share / cp / rename HTTPErrors so the
// cobra layer can reuse the standard 401 / 403 / 404 reformatter
// pattern via errors.As.
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

// Client is the per-FilesURL handle for SMB mount calls. HTTPClient
// is the factory-provided client whose refreshingTransport injects
// X-Authorization and refreshes on 401/403, same as the share / cp
// clients.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// nodesEnvelope mirrors the {data:{nodes:[...]}} shape served by
// /api/nodes/. Same shape cp / upload decode locally.
type nodesEnvelope struct {
	Data struct {
		Nodes []Node `json:"nodes"`
	} `json:"data"`
}

// FetchNodes returns the configured Drive nodes. Used by the cobra
// layer to resolve the default `{node}` URL segment when the user
// did NOT pass --node — the convention the LarePass app follows is
// `currentNode` ⇒ first /api/nodes/ entry on a fresh page load, so
// the CLI mirrors that.
func (c *Client) FetchNodes(ctx context.Context) ([]Node, error) {
	endpoint := c.BaseURL + "/api/nodes/"
	body, err := c.do(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var env nodesEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("decode /api/nodes/ response: %w (body=%s)", err, truncateBody(body))
	}
	if len(env.Data.Nodes) == 0 {
		return nil, errors.New("files-backend returned no Drive nodes; cannot resolve default {node}")
	}
	return env.Data.Nodes, nil
}

// mountEnvelope is the body of POST /api/mount/<node>/?external_type=smb.
// Two branches:
//
//   - code 200: `data` is some metadata object — we don't need it.
//   - code 300: `data` is `[{path:"...", ...}, ...]` — the list of
//     shares the user has to pick from.
//
// We decode `data` into a json.RawMessage first and only re-decode
// it as a path list when code == 300 — keeps the happy path cheap
// and avoids a "expected array, got object" parse error on success.
type mountEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// mountPath is one row of the code-300 `data` array. We only read
// `path` — the GUI also unpacks `sambaPath` / `dir` from it, but
// those are client-side conveniences computed off the same string.
type mountPath struct {
	Path string `json:"path"`
}

// Mount sends one POST /api/mount/<node>/?external_type=smb and
// returns the parsed result. `node` may be empty — the LarePass app
// drops the `<node>` segment entirely when no nodes are configured
// (see stores/files.ts L1266-L1272: `node.length ? '/' + node + '/' : ''`).
//
// Returned errors:
//   - *HTTPError: non-2xx status from the server (transport-level
//     failure, treated by the cobra layer as a 401/403/404
//     reformatting candidate).
//   - fmt.Errorf("server rejected (code N): ..."): the server
//     returned 2xx but a non-200/300 envelope code. The wire
//     message is preserved so the cobra layer can show it verbatim.
//
// On code == 300 the result is a non-nil *MountResult with
// Code == 300 and Paths populated; this is NOT an error — the cobra
// layer renders it as "pick a share and re-run".
func (c *Client) Mount(ctx context.Context, node string, opts MountOptions) (*MountResult, error) {
	if opts.SMBPath == "" {
		return nil, errors.New("Mount: empty SMB path")
	}
	endpoint := c.BaseURL + buildMountURL(node, "smb")
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("Mount: marshal body: %w", err)
	}
	raw, err := c.do(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	var env mountEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("Mount: decode response: %w (body=%s)", err, truncateBody(raw))
	}
	switch env.Code {
	case 200:
		return &MountResult{Code: 200, Message: env.Message}, nil
	case 300:
		var rows []mountPath
		if len(env.Data) > 0 {
			if err := json.Unmarshal(env.Data, &rows); err != nil {
				return nil, fmt.Errorf("Mount: decode share-list payload: %w (body=%s)", err, truncateBody(raw))
			}
		}
		paths := make([]string, 0, len(rows))
		for _, r := range rows {
			if r.Path != "" {
				paths = append(paths, r.Path)
			}
		}
		return &MountResult{Code: 300, Message: env.Message, Paths: paths}, nil
	default:
		// Surface the wire message + code so the user has both the
		// human-readable "what" and the server-discriminated "why".
		// The LarePass app routes this through `notifyFailed(msg)` —
		// we just throw the same string in an error wrapper.
		msg := env.Message
		if msg == "" {
			msg = "no message"
		}
		return nil, fmt.Errorf("server rejected (code %d): %s", env.Code, msg)
	}
}

// NFSMountOptions captures the body of POST
// /api/mount/<node>/?external_type=nfs.
//
// NFS mounts need no credentials (unlike SMB), so the wire body is
// just the target URL — either a full `host:/export` path (mount
// it directly) or a bare host/IP plus `operate:"list"` to ask the
// server to enumerate the host's exports. This mirrors LarePass's
// buildNfsMountPayload (stores/files.ts):
//
//	list request : {"url":"<host>", "operate":"list"}
//	mount request: {"url":"<host>:/<export>"}
type NFSMountOptions struct {
	// URL is the NFS target: `host` / `host:/export`.
	URL string
	// List, when true, sends `operate:"list"` so the server
	// enumerates the host's exports instead of mounting. The cobra
	// layer sets this when the user passed a bare host (no export
	// path) — the discovery half of the "Connect to Server" flow.
	List bool
}

// NFSExport is one row of an NFS list response. The server reports
// each export's path and whether it is already mounted (so the GUI
// can grey out mounted entries in its chooser; the CLI annotates
// them the same way). The daemon's underlying `showmount -e` also
// carries an ACL string, but the files-backend list surface the GUI
// consumes only needs path + mounted, so that's all we decode.
type NFSExport struct {
	Path    string `json:"path"`
	Mounted bool   `json:"mounted,omitempty"`
}

// NFSMountResult is the parsed body of an NFS mount/list POST.
//
// Unlike SMB (which signals "host-only, here are the shares" with a
// distinct code 300), the NFS surface returns the export list under
// the SAME code 200 it uses for a successful mount — the caller
// disambiguates via the request it sent (a list request vs. a real
// mount). We therefore carry an explicit `Listed` flag rather than
// overloading the code, matching LarePass's `isListRequest` switch.
type NFSMountResult struct {
	Code    int         // server envelope code (200 on the happy path)
	Message string      // server-supplied diagnostic, surfaced verbatim
	Listed  bool        // true ⇒ this was a list response; Exports is populated
	Exports []NFSExport // populated only when Listed
}

// MountNFS sends one POST /api/mount/<node>/?external_type=nfs.
//
// Two shapes, selected by opts.List:
//
//   - list request (opts.List == true): body {url, operate:"list"};
//     the server replies with the host's export list. The result
//     has Listed == true and Exports populated.
//   - mount request (opts.List == false): body {url}; the server
//     mounts `host:/export` into external/<node>/<entry>/. The
//     result has Listed == false and Code 200 on success.
//
// `node` may be empty — the same conditional `<node>` segment SMB
// uses applies here.
func (c *Client) MountNFS(ctx context.Context, node string, opts NFSMountOptions) (*NFSMountResult, error) {
	if opts.URL == "" {
		return nil, errors.New("MountNFS: empty url")
	}
	payload := map[string]string{"url": opts.URL}
	if opts.List {
		payload["operate"] = "list"
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("MountNFS: marshal body: %w", err)
	}
	endpoint := c.BaseURL + buildMountURL(node, "nfs")
	raw, err := c.do(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	var env mountEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("MountNFS: decode response: %w (body=%s)", err, truncateBody(raw))
	}

	// A list request always yields an export list (code 200 in the
	// common case; we also tolerate 300 in case a future backend
	// aligns NFS discovery with the SMB code). Decode the array and
	// hand it back annotated as Listed.
	if opts.List {
		if env.Code != 0 && env.Code != 200 && env.Code != 300 {
			return nil, fmt.Errorf("server rejected (code %d): %s", env.Code, msgOrDefault(env.Message))
		}
		exports, derr := decodeNFSExports(env.Data, raw)
		if derr != nil {
			return nil, derr
		}
		return &NFSMountResult{Code: env.Code, Message: env.Message, Listed: true, Exports: exports}, nil
	}

	switch env.Code {
	case 0, 200:
		return &NFSMountResult{Code: 200, Message: env.Message}, nil
	case 300:
		// The user passed what we classified as a full path but the
		// server still wants a pick — surface the list rather than
		// failing opaquely. Mirrors the SMB code-300 fallback.
		exports, derr := decodeNFSExports(env.Data, raw)
		if derr != nil {
			return nil, derr
		}
		return &NFSMountResult{Code: 300, Message: env.Message, Listed: true, Exports: exports}, nil
	default:
		return nil, fmt.Errorf("server rejected (code %d): %s", env.Code, msgOrDefault(env.Message))
	}
}

// decodeNFSExports parses the `data` array of an NFS list response
// into []NFSExport, tolerating an absent / empty payload (returns an
// empty slice). `raw` is only used to enrich the decode error.
func decodeNFSExports(data json.RawMessage, raw []byte) ([]NFSExport, error) {
	if len(data) == 0 {
		return []NFSExport{}, nil
	}
	var rows []NFSExport
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("MountNFS: decode export-list payload: %w (body=%s)", err, truncateBody(raw))
	}
	return rows, nil
}

// msgOrDefault folds an empty server message into a stable
// placeholder so error strings never read "... : ".
func msgOrDefault(msg string) string {
	if msg == "" {
		return "no message"
	}
	return msg
}

// unmountEnvelope is the body of POST /api/unmount/<...>/. Just a
// {code, message} pair — we don't need `data`.
type unmountEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// Unmount removes a previously-mounted external entry from the
// `external/<node>/` namespace. `fileType` is always "external" in
// the SMB workflow but the parameter is exposed so a future
// "unmount usb" / "unmount hdd" command can reuse this client.
//
// `externalType` MUST match the value used at mount time — for SMB
// shares that's "smb" (LarePass's ExternalType.SMB). The server
// tracks the entry's original external type so passing the wrong
// one here gets a server-side rejection.
//
// Returned errors mirror Mount: *HTTPError on non-2xx, descriptive
// fmt.Errorf on a non-200 envelope code.
func (c *Client) Unmount(ctx context.Context, fileType, fileExtend, name, externalType string) error {
	if fileType == "" {
		return errors.New("Unmount: empty fileType")
	}
	if fileExtend == "" {
		return errors.New("Unmount: empty fileExtend (node name)")
	}
	if name == "" {
		return errors.New("Unmount: empty name")
	}
	if externalType == "" {
		return errors.New("Unmount: empty externalType")
	}
	endpoint := c.BaseURL + buildUnmountURL(fileType, fileExtend, name, externalType)
	// Empty body — every parameter rides in the URL/query string.
	// LarePass sends `{}` (`CommonFetch.post(path, {})`); we do the
	// same so any server-side body length checks behave identically.
	raw, err := c.do(ctx, http.MethodPost, endpoint, []byte("{}"))
	if err != nil {
		return err
	}
	return decodeOptionalEnvelope(raw, "Unmount")
}

// HistoryList fetches the per-node SMB favorites list. The wire
// shape is the array directly (no envelope) — we tolerate an empty
// or absent body by returning an empty slice.
//
// `node` is required: every history entry is scoped to a specific
// node, mirroring how the GUI tabs the favorites list per
// `currentNode`.
func (c *Client) HistoryList(ctx context.Context, node string) ([]HistoryEntry, error) {
	if node == "" {
		return nil, errors.New("HistoryList: empty node")
	}
	endpoint := c.BaseURL + buildHistoryURL(node)
	raw, err := c.do(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return []HistoryEntry{}, nil
	}
	// The body is `[]` or `[{...}, ...]`. Be defensive against the
	// alternative envelope shape `{code, message, data:[...]}` —
	// the GUI uses the bare-array reading, but a future server-side
	// refactor could switch shapes without warning. Try both.
	if raw[0] == '[' {
		var out []HistoryEntry
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("HistoryList: decode array body: %w (body=%s)", err, truncateBody(raw))
		}
		return out, nil
	}
	var env struct {
		Code    int            `json:"code"`
		Message string         `json:"message,omitempty"`
		Data    []HistoryEntry `json:"data,omitempty"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("HistoryList: decode envelope body: %w (body=%s)", err, truncateBody(raw))
	}
	if env.Code != 0 && env.Code != 200 {
		msg := env.Message
		if msg == "" {
			msg = "no message"
		}
		return nil, fmt.Errorf("server rejected (code %d): %s", env.Code, msg)
	}
	if env.Data == nil {
		return []HistoryEntry{}, nil
	}
	return env.Data, nil
}

// HistoryUpsert PUTs an array of HistoryEntry rows. The server
// merges by `url`: an existing row is overwritten in place; a new
// `url` is appended. This is the same call ConnectServerStep1.vue's
// saveFavorite and ConnectServerStep3.vue's saveSmbHistory both
// fire — credentials may be omitted (favorite-only) or included
// (auto-reconnect).
//
// An empty `entries` is rejected client-side: the GUI would never
// send an empty list, and an empty PUT is at best a no-op and at
// worst a server-side wipe; surface the typo immediately rather
// than gambling on the server's interpretation.
func (c *Client) HistoryUpsert(ctx context.Context, node string, entries []HistoryEntry) error {
	if node == "" {
		return errors.New("HistoryUpsert: empty node")
	}
	if len(entries) == 0 {
		return errors.New("HistoryUpsert: empty entries")
	}
	for i, e := range entries {
		if e.URL == "" {
			return fmt.Errorf("HistoryUpsert: entries[%d] has empty url", i)
		}
	}
	endpoint := c.BaseURL + buildHistoryURL(node)
	body, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("HistoryUpsert: marshal body: %w", err)
	}
	return c.simpleEnvelopePUTorDELETE(ctx, http.MethodPut, endpoint, body, "HistoryUpsert")
}

// HistoryRemove DELETEs the rows whose `url` matches any entry in
// `urls`. Mirrors removeFavorite in ConnectServerStep1.vue.
//
// Cobra layer collects the URLs into the body shape `[{url}, ...]`
// — we keep the body construction here so callers don't have to
// know that the wire shape uses an array of objects rather than a
// query string.
func (c *Client) HistoryRemove(ctx context.Context, node string, urls []string) error {
	if node == "" {
		return errors.New("HistoryRemove: empty node")
	}
	if len(urls) == 0 {
		return errors.New("HistoryRemove: empty urls")
	}
	rows := make([]struct {
		URL string `json:"url"`
	}, 0, len(urls))
	for i, u := range urls {
		if u == "" {
			return fmt.Errorf("HistoryRemove: urls[%d] is empty", i)
		}
		rows = append(rows, struct {
			URL string `json:"url"`
		}{u})
	}
	endpoint := c.BaseURL + buildHistoryURL(node)
	body, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("HistoryRemove: marshal body: %w", err)
	}
	return c.simpleEnvelopePUTorDELETE(ctx, http.MethodDelete, endpoint, body, "HistoryRemove")
}

// simpleEnvelopePUTorDELETE consolidates the shared "fire the
// request, surface non-2xx via *HTTPError, otherwise treat as
// success" tail for HistoryUpsert and HistoryRemove. Both endpoints
// reply with one of three body shapes (verified against the live
// per-user files-backend):
//
//   - empty body
//   - JSON envelope `{code, message}` — only on a code-bearing
//     server-side rejection
//   - plaintext "Successfully ...." — the happy-path shape on the
//     current backend (e.g. DELETE /api/smb_history/<node>/ replies
//     with `Successfully deleted SMB history`)
//
// LarePass's `removeFavorite` / `saveFavorite` await the response
// and immediately re-fetch the history list — they never inspect the
// body at all (see ConnectServerStep1.vue L131-L140). The CLI must
// be at least as forgiving: any 2xx-with-non-JSON-object body is
// success.
func (c *Client) simpleEnvelopePUTorDELETE(ctx context.Context, method, endpoint string, body []byte, op string) error {
	raw, err := c.do(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	return decodeOptionalEnvelope(raw, op)
}

// decodeOptionalEnvelope is the shared "tolerate non-JSON success
// bodies" routine used by Unmount, HistoryUpsert, and HistoryRemove.
//
// Decoding rules — pinned to LarePass's behavior, NOT to a strict
// per-endpoint body contract:
//
//   - empty / whitespace-only           → success
//   - body whose first non-space byte
//     is NOT `{`                        → success (plaintext like
//     "Successfully deleted SMB history",
//     or a future-format change that
//     LarePass would silently ignore)
//   - JSON object with code 0 / 200     → success
//   - JSON object with any other code   → "server rejected (code
//     N): <message>" error
//
// We deliberately do NOT try to JSON-decode arrays / strings here:
// the only verbs that use this helper are PUT/DELETE/POST whose
// success-shape is "empty or status text". HistoryList does its own
// decoding (it expects a real array body) and isn't routed through
// here.
func decodeOptionalEnvelope(raw []byte, op string) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil
	}
	if trimmed[0] != '{' {
		return nil
	}
	var env unmountEnvelope
	if err := json.Unmarshal(trimmed, &env); err != nil {
		return fmt.Errorf("%s: decode response: %w (body=%s)", op, err, truncateBody(raw))
	}
	if env.Code != 0 && env.Code != 200 {
		msg := env.Message
		if msg == "" {
			msg = "no message"
		}
		return fmt.Errorf("%s: server rejected (code %d): %s", op, env.Code, msg)
	}
	return nil
}

// do is the shared HTTP machinery. Mirrors share.Client.do /
// permission.Client.do — marshal a body if present, set
// Content-Type / Accept, fire the request, surface non-2xx as
// *HTTPError so the cobra reformatter can branch on status.
func (c *Client) do(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(raw),
			URL:    endpoint,
			Method: method,
		}
	}
	return raw, nil
}

// buildMountURL renders /api/mount[/<node>]/?external_type=<type>.
//
// `externalType` is "smb" for SMB mounts and "nfs" for NFS mounts —
// the shared per-user files-backend mount surface dispatches on the
// query parameter (LarePass's mountServerInExternal sends the same
// `?external_type=` switch; see stores/files.ts).
//
// The `<node>` segment is conditionally present — when the
// per-user files-backend has no clustered nodes configured, the
// LarePass app drops it entirely (`/api/mount/?external_type=smb`).
// We replicate the same conditional shape rather than always
// inserting `<node>` so a single-node-less deployment doesn't get
// a 404 from a stricter route matcher.
func buildMountURL(node, externalType string) string {
	q := url.Values{}
	q.Set("external_type", externalType)
	if node == "" {
		return "/api/mount/?" + q.Encode()
	}
	return "/api/mount/" + encodepath.EncodeURL(node) + "/?" + q.Encode()
}

// buildUnmountURL renders
// /api/unmount/<fileType>/<fileExtend>/<name>/?external_type=<type>.
//
// Per-segment percent-encoding via encodepath.EncodeURL keeps the
// CLI byte-identical to LarePass — the same `encodeUrl` helper
// share / cp / rm / upload all defer to. The trailing slash on the
// `<name>/` segment matches the GUI exactly (the URL is built by
// string concatenation in stores/operation.ts L861-L870).
func buildUnmountURL(fileType, fileExtend, name, externalType string) string {
	q := url.Values{}
	q.Set("external_type", externalType)
	plain := fileType + "/" + fileExtend + "/" + name
	enc := encodepath.EncodeURL(plain)
	if !strings.HasSuffix(enc, "/") {
		enc += "/"
	}
	return "/api/unmount/" + enc + "?" + q.Encode()
}

// buildHistoryURL renders /api/smb_history/<node>/.
//
// Used by HistoryList (GET), HistoryUpsert (PUT), and
// HistoryRemove (DELETE). Per-segment percent-encoding so a node
// name with a non-ASCII char doesn't break the path.
func buildHistoryURL(node string) string {
	return "/api/smb_history/" + encodepath.EncodeURL(node) + "/"
}

// truncateBody renders a (possibly large) response body for
// inclusion in error messages without dumping multi-MB blobs into
// the user's terminal. The 500-byte cap matches HTTPError.Error.
func truncateBody(b []byte) string {
	const max = 500
	if len(b) > max {
		return string(b[:max]) + "...(truncated)"
	}
	return string(b)
}

// Compile-time assertion: keep the public error surface stable so
// the cobra layer's `errors.As(err, &*HTTPError{})` branches don't
// break under refactors.
var _ error = (*HTTPError)(nil)
