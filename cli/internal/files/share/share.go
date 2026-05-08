// Package share implements the wire side of `olares-cli files share`,
// covering the three folder-sharing surfaces the LarePass web app
// exposes: Internal (in-Olares cross-user shares), Public (external
// links with optional password / expiration / upload-only), and SMB
// (Samba network shares).
//
// All three share types use the SAME create endpoint —
//
//	POST /api/share/share_path/<fileType>/<extend><subPath>/
//
// — and disambiguate via the `share_type` field in the JSON body
// (values: "internal" / "external" / "smb"). Source of truth is the
// LarePass web app's `share.create` helper at
// apps/packages/app/src/api/files/v2/common/share.ts L27-54, plus the
// per-type wrappers in components/files/share/{Internal,Public,SMB}/.
//
// Differences vs. cp / rename:
//
//   - The /api/share/ surface returns a {code, message, data} JSON
//     envelope on every verb (CommonFetch's interceptor surfaces
//     code != 0 as an error). We decode it explicitly and surface
//     code != 0 as an error with the server's message — same pattern
//     cp uses for `code: -1`.
//   - There's no <node> URL segment; the share endpoints are
//     uniformly served per-user without per-node routing. That makes
//     the wire client simpler than cp's PasteOne.
//   - List / Remove / Query share a single base URL
//     (/api/share/share_path/) and disambiguate by HTTP method +
//     query params, mirroring the web app exactly.
package share

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// Type is the share-flavor enum the backend's `share_type` field accepts.
// Wire values are the LarePass web app's ShareType constants — note
// that "external" (NOT "public") is the value for public-link shares;
// the CLI command surface calls this "public" because that's what the
// UI labels and the user types, but the wire stays "external".
type Type string

const (
	TypeInternal Type = "internal"
	TypePublic   Type = "external"
	TypeSMB      Type = "smb"
)

// Permission mirrors SharePermission in
// apps/packages/app/src/utils/interface/share.ts:
//
//	0 = none / no access (used in filter contexts; not a sensible
//	    default when CREATING a share)
//	1 = view (read-only)
//	2 = upload-only (public-link upload portal — user can drop
//	    files in but can't browse)
//	3 = edit (read + write; used for SMB read/write members and as
//	    the default for public-link recipients)
//	4 = admin (full control; the share's owner permission)
//
// The backend treats this as a small int, NOT an enum string — the
// JSON tag stays as the integer.
type Permission int

const (
	PermNone   Permission = 0
	PermView   Permission = 1
	PermUpload Permission = 2
	PermEdit   Permission = 3
	PermAdmin  Permission = 4
)

// String returns the canonical CLI label for a permission. The
// reverse direction is ParsePermission. Unknown values render as the
// underlying integer so a future server-side addition we don't
// recognise still produces a readable diagnostic.
func (p Permission) String() string {
	switch p {
	case PermNone:
		return "none"
	case PermView:
		return "view"
	case PermUpload:
		return "upload"
	case PermEdit:
		return "edit"
	case PermAdmin:
		return "admin"
	default:
		return strconv.Itoa(int(p))
	}
}

// ParsePermission accepts either a canonical label (case-insensitive)
// or a numeric string and returns the matching Permission. Used by
// the cobra layer to parse `--users user:edit` style flags. Numeric
// inputs are validated against the known range so a typo like ":7"
// fails fast instead of being shipped to the server.
func ParsePermission(s string) (Permission, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "view", "read", "ro", "readonly", "read-only":
		return PermView, nil
	case "upload", "upload-only", "uploadonly":
		return PermUpload, nil
	case "edit", "rw", "write", "read-write", "readwrite":
		return PermEdit, nil
	case "admin", "owner", "manage":
		return PermAdmin, nil
	case "none", "":
		// Treat empty as "caller didn't say" — the call site decides
		// the default. Returning PermNone with no error keeps the
		// flag parser simple; the cobra layer applies a per-command
		// fallback.
		return PermNone, nil
	}
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		switch Permission(n) {
		case PermNone, PermView, PermUpload, PermEdit, PermAdmin:
			return Permission(n), nil
		}
		return 0, fmt.Errorf("permission %d is out of range; expected 0..4 (none/view/upload/edit/admin)", n)
	}
	return 0, fmt.Errorf("unknown permission %q; expected one of: view, upload, edit, admin (or 0..4)", s)
}

// Target identifies a remote file or folder to share. Same shape as
// the cp / rename packages' Target — we keep it package-local so the
// share package stays free of a dep on cp / rename internals.
type Target struct {
	FileType    string
	Extend      string
	SubPath     string
	IsDirIntent bool
}

// CreateOptions captures every body field share.create accepts in the
// web app. Unset fields are omitted from the JSON body so the server
// applies its own defaults — this matters because (a) the backend
// rejects empty `expire_time` in some contexts but treats a missing
// key as "no expiration", and (b) Internal shares pass exactly
// {name, share_type, permission, password=""} and nothing else; any
// stray key would change semantics.
type CreateOptions struct {
	Name            string     `json:"name"`
	ShareType       Type       `json:"share_type"`
	Permission      Permission `json:"permission"`
	Password        string     `json:"password"`
	ExpireIn        int64      `json:"expire_in,omitempty"`
	ExpireTime      string     `json:"expire_time,omitempty"`
	Users           []SMBUser  `json:"users,omitempty"`
	PublicSMB       *bool      `json:"public_smb,omitempty"`
	UploadSizeLimit int64      `json:"upload_size_limit,omitempty"`
}

// SMBUser is one row of CreateOptions.Users — the SMB account ID +
// permission pair. Reused by UpdateSMBShareMember as well.
//
// Member kept as a separate type from SMBUser because Internal-share
// members address by NAME (`share_member`) while SMB members address
// by SMB-account ID (`id`); these aren't interchangeable on the wire.
type SMBUser struct {
	ID         string     `json:"id"`
	Permission Permission `json:"permission"`
}

// Member is the Internal-share counterpart of SMBUser: addresses by
// the user's name, not by an SMB-account ID. addMember / updateInternalShareMembers
// take a slice of these.
type Member struct {
	ShareMember string     `json:"share_member"`
	Permission  Permission `json:"permission"`
}

// Result is the {data: ...} payload share.create / share.query return.
// JSON tags lifted verbatim from the web app's ShareResult interface
// (apps/.../utils/interface/share.ts). Integer time fields are echoed
// as ISO strings by the server — kept as string here so we don't
// silently mangle a format the user might want to see verbatim.
type Result struct {
	ID              string `json:"id"`
	Owner           string `json:"owner"`
	FileType        string `json:"file_type"`
	Extend          string `json:"extend"`
	Path            string `json:"path"`
	ShareType       Type   `json:"share_type"`
	Name            string `json:"name"`
	ExpireIn        int64  `json:"expire_in"`
	ExpireTime      string `json:"expire_time"`
	Permission      Permission `json:"permission"`
	CreateTime      string `json:"create_time"`
	UpdateTime      string `json:"update_time"`
	SharedByMe      bool   `json:"shared_by_me"`
	SMBLink         string `json:"smb_link,omitempty"`
	SMBUser         string `json:"smb_user,omitempty"`
	SMBPassword     string `json:"smb_password,omitempty"`
	UploadSizeLimit int64  `json:"upload_size_limit,omitempty"`
	SyncRepoName    string `json:"sync_repo_name,omitempty"`
	Node            string `json:"node,omitempty"`
}

// ListParams is the GET /api/share/share_path/ query-string shape for
// listing and filtering shares. Matches the web app's getShareList
// params verbatim (share.ts L12-25). Empty values are omitted.
//
// SharedToMe / SharedByMe are *bool so the caller can pass "false"
// explicitly without it being indistinguishable from "unset" — the
// web app sends them on every list call.
type ListParams struct {
	SharedToMe *bool
	SharedByMe *bool
	ExpireIn   int64
	ExpireOver int64
	ShareType  string // comma-joined list of Type values
	Owner      string // comma-joined owner names
	Permission string // comma-joined permission integers
	PathID     string // single-share lookup; mutually exclusive with the filters above
}

// Client is the per-FilesURL handle for share calls. HTTPClient is
// the factory-provided client whose refreshingTransport injects
// X-Authorization (NOT Authorization: Bearer — see pkg/cmdutil/factory.go)
// and refreshes on 401/403.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// HTTPError carries the status + truncated body of a non-2xx response,
// same shape as the cp / rename HTTPErrors so the cobra layer can
// reuse the standard 401 / 403 / 404 reformatter pattern.
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

// envelope is the {code, message, data} wrapper the share endpoints
// always use. We keep Code as *int because the server occasionally
// omits the field on success (some endpoints return just {data: ...}),
// so a missing key shouldn't be confused with code:0.
//
// Generic over the data shape so callers can decode straight into
// Result / []Result / json.RawMessage etc. without an intermediate
// json.Unmarshal step.
type envelope[T any] struct {
	Code    *int   `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

// Create posts to /api/share/share_path/<fileType>/<extend><subPath>/
// with the CreateOptions body and returns the resulting share record.
// The caller is responsible for assembling type-specific options
// (e.g. setting CreateOptions.PublicSMB only for SMB shares); this
// function is intentionally type-agnostic so a hypothetical fourth
// share type wouldn't need a new method here.
//
// Wire path matches share.create at
// apps/.../api/files/v2/common/share.ts L45-52: per-segment percent-
// encoded fileType / extend / subPath, with an unconditional trailing
// '/' so a file-target share doesn't get routed through the directory
// handler.
func (c *Client) Create(ctx context.Context, t Target, opts CreateOptions) (*Result, error) {
	if t.FileType == "" || t.Extend == "" {
		return nil, fmt.Errorf("share Create: empty fileType or extend (got %q/%q)", t.FileType, t.Extend)
	}
	if opts.ShareType == "" {
		return nil, errors.New("share Create: ShareType is required (internal/external/smb)")
	}
	if opts.Name == "" {
		return nil, errors.New("share Create: Name is required (the human-readable share label)")
	}

	endpoint := c.BaseURL + buildSharePathURL(t)
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("share Create: marshal body: %w", err)
	}

	var env envelope[Result]
	if err := c.do(ctx, http.MethodPost, endpoint, body, &env); err != nil {
		return nil, err
	}
	if env.Code != nil && *env.Code != 0 {
		return nil, fmt.Errorf("share Create: server rejected (code %d): %s", *env.Code, env.Message)
	}
	// Defense in depth: an empty Result.ID points at a server-side
	// silent failure (the create endpoint is supposed to return the
	// canonical ID even on auto-rename). Surface it loudly rather
	// than letting the caller proceed with a useless handle.
	if env.Data.ID == "" {
		return nil, fmt.Errorf("share Create: server returned no share id (envelope: %+v)", env)
	}
	return &env.Data, nil
}

// AddInternalMembers posts to /api/share/share_member/. Used after
// Create for Internal shares to grant access to specific users —
// matches share.addMember at share.ts L56-66.
//
// Empty members is a no-op (returns nil) so the cobra layer can
// always call this after Create without checking the slice itself.
func (c *Client) AddInternalMembers(ctx context.Context, pathID string, members []Member) error {
	if pathID == "" {
		return errors.New("share AddInternalMembers: empty pathID")
	}
	if len(members) == 0 {
		return nil
	}
	body, err := json.Marshal(struct {
		PathID       string   `json:"path_id"`
		ShareMembers []Member `json:"share_members"`
	}{pathID, members})
	if err != nil {
		return fmt.Errorf("share AddInternalMembers: marshal body: %w", err)
	}
	var env envelope[json.RawMessage]
	if err := c.do(ctx, http.MethodPost, c.BaseURL+"/api/share/share_member/", body, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("share AddInternalMembers: server rejected (code %d): %s", *env.Code, env.Message)
	}
	return nil
}

// UpdateInternalMembers does a PUT against /api/share/share_path/share_members/.
// Used to bulk-set the member list of an Internal share — mirrors the
// web app's updateInternalShareMembers at share.ts L220-235.
//
// Unlike AddInternalMembers (which appends), this REPLACES the full
// member list. Pass an empty slice to clear all members.
func (c *Client) UpdateInternalMembers(ctx context.Context, pathID string, members []Member) error {
	if pathID == "" {
		return errors.New("share UpdateInternalMembers: empty pathID")
	}
	body, err := json.Marshal(struct {
		PathID       string   `json:"path_id"`
		ShareMembers []Member `json:"share_members"`
	}{pathID, members})
	if err != nil {
		return fmt.Errorf("share UpdateInternalMembers: marshal body: %w", err)
	}
	var env envelope[json.RawMessage]
	if err := c.do(ctx, http.MethodPut, c.BaseURL+"/api/share/share_path/share_members/", body, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("share UpdateInternalMembers: server rejected (code %d): %s", *env.Code, env.Message)
	}
	return nil
}

// UpdateSMBShareMember posts to /api/share/smb_share_member/. The
// body's `users` array is the SMB-account-ID list (NOT the user-name
// list AddInternalMembers uses), and `public_smb` toggles the
// "anyone on the local network" mode. Matches share.updateSMBShareMember
// at share.ts L198-210.
func (c *Client) UpdateSMBShareMember(ctx context.Context, pathID string, users []SMBUser, publicSMB bool) error {
	if pathID == "" {
		return errors.New("share UpdateSMBShareMember: empty pathID")
	}
	body, err := json.Marshal(struct {
		PathID    string    `json:"path_id"`
		Users     []SMBUser `json:"users"`
		PublicSMB bool      `json:"public_smb"`
	}{pathID, users, publicSMB})
	if err != nil {
		return fmt.Errorf("share UpdateSMBShareMember: marshal body: %w", err)
	}
	var env envelope[json.RawMessage]
	if err := c.do(ctx, http.MethodPost, c.BaseURL+"/api/share/smb_share_member/", body, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("share UpdateSMBShareMember: server rejected (code %d): %s", *env.Code, env.Message)
	}
	return nil
}

// ResetPassword does a PUT against /api/share/share_password/. Used
// to roll a Public-link share's access password without re-creating
// the share (the share id stays stable). Matches share.resetPassword
// at share.ts L212-218.
func (c *Client) ResetPassword(ctx context.Context, pathID, password string) error {
	if pathID == "" {
		return errors.New("share ResetPassword: empty pathID")
	}
	body, err := json.Marshal(struct {
		PathID   string `json:"path_id"`
		Password string `json:"password"`
	}{pathID, password})
	if err != nil {
		return fmt.Errorf("share ResetPassword: marshal body: %w", err)
	}
	var env envelope[json.RawMessage]
	if err := c.do(ctx, http.MethodPut, c.BaseURL+"/api/share/share_password/", body, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("share ResetPassword: server rejected (code %d): %s", *env.Code, env.Message)
	}
	return nil
}

// Remove deletes one or more share records by id. Matches share.remove
// at share.ts L77-82: comma-joined IDs in the `path_ids` query param.
//
// Empty IDs slice is rejected here (unlike AddInternalMembers's no-op
// behavior) — DELETE with an empty path_ids would be a wire bug, not
// a meaningful no-op.
func (c *Client) Remove(ctx context.Context, pathIDs []string) error {
	if len(pathIDs) == 0 {
		return errors.New("share Remove: empty pathIDs")
	}
	for _, id := range pathIDs {
		if id == "" {
			return errors.New("share Remove: pathIDs contains an empty entry")
		}
	}
	q := url.Values{}
	q.Set("path_ids", strings.Join(pathIDs, ","))
	endpoint := c.BaseURL + "/api/share/share_path/?" + q.Encode()
	var env envelope[json.RawMessage]
	if err := c.do(ctx, http.MethodDelete, endpoint, nil, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("share Remove: server rejected (code %d): %s", *env.Code, env.Message)
	}
	return nil
}

// Query fetches a single share by id. Matches share.query at
// share.ts L84-93: GET /api/share/share_path/?path_id=<id> and the
// response shape is {share_paths: [Result]} (NOT the {data:...}
// envelope the create endpoint uses) — the web app pulls the first
// element out, and we do the same.
//
// Returns nil result + nil error when the server replies with an
// empty share_paths slice (i.e. the id doesn't exist) so the caller
// can branch on absence without inspecting an error message.
func (c *Client) Query(ctx context.Context, pathID string) (*Result, error) {
	if pathID == "" {
		return nil, errors.New("share Query: empty pathID")
	}
	q := url.Values{}
	q.Set("path_id", pathID)
	endpoint := c.BaseURL + "/api/share/share_path/?" + q.Encode()
	var raw struct {
		SharePaths []Result `json:"share_paths"`
		Code       *int     `json:"code,omitempty"`
		Message    string   `json:"message,omitempty"`
	}
	if err := c.do(ctx, http.MethodGet, endpoint, nil, &raw); err != nil {
		return nil, err
	}
	if raw.Code != nil && *raw.Code != 0 {
		return nil, fmt.Errorf("share Query: server rejected (code %d): %s", *raw.Code, raw.Message)
	}
	if len(raw.SharePaths) == 0 {
		return nil, nil
	}
	return &raw.SharePaths[0], nil
}

// List does GET /api/share/share_path/ with the filter params from
// ListParams. Mirrors share.getShareList at share.ts L12-25 — the
// response shape is {share_paths: [Result], ...} on the wire (same as
// Query, just with multiple rows).
//
// Empty fields in ListParams are omitted from the query string so
// the server's defaults apply uniformly. SharedToMe / SharedByMe
// being *bool lets us send "false" explicitly when the caller cares.
func (c *Client) List(ctx context.Context, params ListParams) ([]Result, error) {
	q := url.Values{}
	if params.SharedToMe != nil {
		q.Set("shared_to_me", strconv.FormatBool(*params.SharedToMe))
	}
	if params.SharedByMe != nil {
		q.Set("shared_by_me", strconv.FormatBool(*params.SharedByMe))
	}
	if params.ExpireIn > 0 {
		q.Set("expire_in", strconv.FormatInt(params.ExpireIn, 10))
	}
	if params.ExpireOver > 0 {
		q.Set("expire_over", strconv.FormatInt(params.ExpireOver, 10))
	}
	if params.ShareType != "" {
		q.Set("share_type", params.ShareType)
	}
	if params.Owner != "" {
		q.Set("owner", params.Owner)
	}
	if params.Permission != "" {
		q.Set("permission", params.Permission)
	}
	if params.PathID != "" {
		q.Set("path_id", params.PathID)
	}

	endpoint := c.BaseURL + "/api/share/share_path/"
	if enc := q.Encode(); enc != "" {
		endpoint += "?" + enc
	}
	var raw struct {
		SharePaths []Result `json:"share_paths"`
		Code       *int     `json:"code,omitempty"`
		Message    string   `json:"message,omitempty"`
	}
	if err := c.do(ctx, http.MethodGet, endpoint, nil, &raw); err != nil {
		return nil, err
	}
	if raw.Code != nil && *raw.Code != 0 {
		return nil, fmt.Errorf("share List: server rejected (code %d): %s", *raw.Code, raw.Message)
	}
	return raw.SharePaths, nil
}

// SMBAccount is one row of GET /api/share/smb_share_user/ — the SMB
// account roster used to populate the Users field on SMB-share
// creation. Mirrors the web app's SMBUser type
// (apps/.../stores/files.ts).
type SMBAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListSMBAccounts fetches the current user's SMB-account roster.
// Returns an empty slice (NOT nil) when there are no accounts so
// callers can range over the result without a length check.
//
// Matches share.getSMBUsers at share.ts L138-148.
func (c *Client) ListSMBAccounts(ctx context.Context) ([]SMBAccount, error) {
	endpoint := c.BaseURL + "/api/share/smb_share_user/"
	var env envelope[[]SMBAccount]
	if err := c.do(ctx, http.MethodGet, endpoint, nil, &env); err != nil {
		return nil, err
	}
	if env.Code != nil && *env.Code != 0 {
		return nil, fmt.Errorf("share ListSMBAccounts: server rejected (code %d): %s", *env.Code, env.Message)
	}
	if env.Data == nil {
		return []SMBAccount{}, nil
	}
	return env.Data, nil
}

// CreateSMBAccount creates a new SMB user under the current Olares
// account. The server returns code:0 on success and a non-zero code
// (with message) on collisions — surface both. Matches
// share.createSMBUser at share.ts L150-167.
func (c *Client) CreateSMBAccount(ctx context.Context, user, password string) error {
	if user == "" {
		return errors.New("share CreateSMBAccount: empty user")
	}
	if password == "" {
		return errors.New("share CreateSMBAccount: empty password")
	}
	body, err := json.Marshal(struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}{user, password})
	if err != nil {
		return fmt.Errorf("share CreateSMBAccount: marshal body: %w", err)
	}
	endpoint := c.BaseURL + "/api/share/smb_share_user/"
	var env envelope[json.RawMessage]
	if err := c.do(ctx, http.MethodPost, endpoint, body, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("share CreateSMBAccount: server rejected (code %d): %s", *env.Code, env.Message)
	}
	return nil
}

// do is the shared HTTP machinery: marshal a body if present, set
// Content-Type / Accept, fire the request, decode the response into
// `out` (which must be a pointer to an envelope or a custom struct
// for non-envelope endpoints like Query / List).
//
// Splitting this out keeps every public method to a few lines of
// envelope-shape decoding without re-implementing the same six lines
// of HTTP boilerplate.
func (c *Client) do(ctx context.Context, method, endpoint string, body []byte, out any) error {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return &HTTPError{
			Status: resp.StatusCode,
			Body:   string(raw),
			URL:    endpoint,
			Method: method,
		}
	}
	// An empty body is acceptable for a few of the verbs (e.g. some
	// PUT/DELETE endpoints reply with just 200 OK). Skip decoding so
	// the envelope decoder doesn't surface "unexpected end of JSON
	// input" for what's actually a successful no-content response.
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w (body: %s)", err, truncateBody(raw))
	}
	return nil
}

// buildSharePathURL is the wire URL builder for the
// /api/share/share_path/<fileType>/<extend><subPath>/ endpoint, used
// by Create. Mirrors share.create at share.ts L45-52: per-segment
// JS-shape percent-encoding, with the trailing '/' unconditionally
// appended so the backend routes through the resource handler
// consistently.
//
// SubPath is normalized to start with '/' (the cobra-layer parser
// already guarantees this; the explicit prefix here is defense in
// depth so a future caller bypassing the parser doesn't generate
// "drive/HomeDocuments/foo" without a separator).
func buildSharePathURL(t Target) string {
	sub := t.SubPath
	if sub == "" {
		sub = "/"
	} else if !strings.HasPrefix(sub, "/") {
		sub = "/" + sub
	}
	plain := t.FileType + "/" + t.Extend + sub
	enc := encodepath.EncodeURL(plain)
	if !strings.HasSuffix(enc, "/") {
		enc += "/"
	}
	return "/api/share/share_path/" + enc
}

// truncateBody renders a (possibly large) response body for inclusion
// in error messages without dumping multi-MB blobs into the user's
// terminal. The 500-byte cap matches HTTPError.Error.
func truncateBody(b []byte) string {
	const max = 500
	if len(b) > max {
		return string(b[:max]) + "...(truncated)"
	}
	return string(b)
}
