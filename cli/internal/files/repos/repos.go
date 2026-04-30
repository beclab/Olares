// Package repos implements the wire side of `olares-cli files repos`,
// the listing of Sync (Seafile / Seahub) repositories under the user's
// account.
//
// The endpoint is the same per-user files-backend the rest of the CLI
// talks to:
//
//	GET /api/repos/                       → repos owned by the user
//	GET /api/repos/?type=share_to_me      → repos shared TO the user
//	GET /api/repos/?type=shared           → repos the user shared with others
//
// All three flavors return the same envelope shape, `{repos: [Repo,
// ...]}`, with the per-row schema mirroring the LarePass web app's
// SyncRepoItemType / SyncRepoSharedItemType (apps/packages/app/src/api/
// files/v2/sync/type.ts). The "mine" / "shared_to_me" / "shared" split
// surfaces extra columns (share_permission, user_email, etc.) only on
// the shared variants, so we keep all of them as optional fields on a
// single Repo struct rather than carrying three near-duplicate types.
//
// Why a dedicated package (rather than e.g. shoving this into
// internal/files/share or internal/files/upload):
//
//   - The repos surface is unrelated to /api/share/ — it's the Sync
//     storage backend's own catalog, accessed via /api/repos/. Keeping
//     it separate avoids confusion between "share records" (cross-user
//     ACL grants on any Drive/Sync path) and "sync repos" (Seafile
//     libraries underneath the sync/ fileType).
//   - Upload/download for Sync paths use the existing /api/resources/
//     and /upload/* surfaces — the only thing they need from this
//     package is the repo id, which the user types directly into the
//     `sync/<repo_id>/...` path argument. So `repos` stays a leaf
//     package with no inbound callers from upload/.
//
// Authentication transport: the factory's HTTPClient injects
// `X-Authorization` and refreshes on 401/403; same convention as every
// other internal/files/* package — see pkg/cmdutil/factory.go's
// authTransport for why X-Authorization (not Authorization: Bearer)
// is the right header for Olares.
package repos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Type selects which slice of the user's repo catalog to fetch.
// Matches the `type` query param accepted by /api/repos/:
//
//	"" (default)     → repos the user owns (used by fetchMineRepo in
//	                   apps/.../api/files/v2/sync/utils.ts L93-101)
//	"share_to_me"    → repos shared TO the user (fetchtosharedRepo,
//	                   utils.ts L103-118)
//	"shared"         → repos the user shared with others (fetchsharedRepo,
//	                   utils.ts L120-135)
//
// Wire values are LOWERCASE strings; the caller-facing CLI uses
// hyphenated forms ("share-to-me") that we translate before fanning
// out — see ParseType.
type Type string

const (
	// TypeMine is the default: repos owned by the current user.
	// On the wire this is sent as an absent `type` query param,
	// matching the web app's fetchMineRepo (no params).
	TypeMine Type = "mine"
	// TypeSharedToMe lists repos that other users have shared with
	// the caller. Wire value: `type=share_to_me`.
	TypeSharedToMe Type = "share_to_me"
	// TypeShared lists repos the caller has shared OUT with other
	// users (i.e. repos where the current user is the source of the
	// share). Wire value: `type=shared`.
	TypeShared Type = "shared"
)

// ParseType maps the user-facing CLI strings (mine / share-to-me /
// shared / share-with-me / shared-to-me) onto canonical Type values.
// Returns an error for unknown inputs so the cobra layer can surface
// a clean "expected one of: ..." message instead of silently sending
// the wrong filter.
//
// The empty string is treated as TypeMine — same default the web app
// uses when no type is specified.
func ParseType(raw string) (Type, error) {
	switch raw {
	case "", "mine":
		return TypeMine, nil
	case "share-to-me", "share_to_me", "shared-to-me", "shared_to_me", "share-with-me":
		return TypeSharedToMe, nil
	case "shared", "shared-by-me", "share-by-me":
		return TypeShared, nil
	default:
		return "", fmt.Errorf("unknown repos type %q (expected one of: mine, share-to-me, shared)", raw)
	}
}

// flexBool decodes the `encrypted` field when the Seahub / gateway
// adapter emits either a JSON bool or the strings "true" / "false" —
// both shapes appear in the wild (see e.g. encrypted as string in
// GET /api/repos/ responses from some nodes).
type flexBool bool

func (b *flexBool) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch t := v.(type) {
	case bool:
		*b = flexBool(t)
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		*b = flexBool(s == "true" || s == "1" || s == "yes" || s == "on")
	case float64:
		*b = flexBool(t != 0)
	case nil:
		*b = false
	default:
		return fmt.Errorf("json: cannot unmarshal %T into flexBool", v)
	}
	return nil
}

// flexInt64 decodes the `size` field when the wire uses either a JSON
// number or a base-10 string (same adapter inconsistency as flexBool).
type flexInt64 int64

func (i *flexInt64) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch t := v.(type) {
	case float64:
		*i = flexInt64(int64(t))
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			*i = 0
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("json: unmarshal flexInt64: %w", err)
		}
		*i = flexInt64(n)
	case nil:
		*i = 0
	default:
		return fmt.Errorf("json: cannot unmarshal %T into flexInt64", v)
	}
	return nil
}

// Repo is the union projection of /api/repos/'s row shapes, covering
// both the "mine" variant (SyncRepoItemType) and the shared variants
// (SyncRepoSharedItemType). All shared-only fields use omitempty so
// the JSON shape round-trips for either flavor.
//
// Field naming follows the wire (snake_case) precisely; the JSON tags
// are the source of truth — the Go field names are the camelCase
// versions for readability. Don't rename without changing the tags
// (the encoding/json default is case-insensitive, so missing tags
// would still mostly work, but explicit tags survive future
// refactors and make the wire shape obvious at a glance).
type Repo struct {
	// --- Common columns (present for every flavor).

	// RepoID is the Seafile library identifier — used as the
	// `<extend>` segment in `sync/<repo_id>/...` front-end paths.
	// Stable across renames.
	RepoID string `json:"repo_id"`
	// RepoName is the display name. Mutable: rename via PATCH
	// /api/repos/?repoId=... (renameRepo in sync/utils.ts).
	RepoName string `json:"repo_name"`
	// Encrypted is true when the repo is client-side encrypted; CLI
	// upload/download against an encrypted repo fail at the Seahub
	// layer until the user unlocks it via the web app (the CLI has
	// no equivalent unlock flow). The `flex` wrapper tolerates
	// non-standard string encodings from some gateways.
	Encrypted flexBool `json:"encrypted"`
	// LastModified is an ISO-8601 timestamp string, NOT a Unix
	// seconds value — keep it as a string here so the cobra layer
	// can render it without a conversion step (matches what
	// the web app shows in the file list).
	LastModified string `json:"last_modified"`
	// Permission is the caller's effective permission on the repo:
	// "rw" for read+write, "r" for read-only, "" for the
	// "type=share_to_me" / "shared" variants (which carry the
	// permission under SharePermission instead).
	Permission string `json:"permission,omitempty"`
	// Size is the total bytes used by the repo. Reported as a 64-bit
	// integer; some nodes send it as a JSON string (flexInt64
	// accepts number or string).
	Size flexInt64 `json:"size,omitempty"`
	// IsVirtual is a Seafile-internal flag; surfaced for parity with
	// the LarePass UI, otherwise ignored by the CLI.
	IsVirtual bool `json:"is_virtual,omitempty"`
	// Type is the server-side classification echoed in the response
	// (e.g. "mine"). Keep it as an opaque string — the CLI doesn't
	// branch on its value.
	Type string `json:"type,omitempty"`
	// Status indicates Seafile's per-repo state ("normal", "broken",
	// ...). Surfaced verbatim so the user can see what the web app
	// would see.
	Status string `json:"status,omitempty"`

	// --- Owner columns (mine + shared variants).

	// OwnerEmail is the canonical owner identifier (an Olares email
	// or a Seafile internal address).
	OwnerEmail string `json:"owner_email,omitempty"`
	// OwnerName is the display name of the owner.
	OwnerName string `json:"owner_name,omitempty"`
	// OwnerContactEmail is the owner's contact email (may differ
	// from OwnerEmail in some deployments where the canonical id is
	// a UUID).
	OwnerContactEmail string `json:"owner_contact_email,omitempty"`

	// --- Modifier columns (mine + shared variants).

	// ModifierEmail is the canonical id of the user who most
	// recently modified the repo.
	ModifierEmail string `json:"modifier_email,omitempty"`
	// ModifierName is the display name of the most recent modifier.
	ModifierName string `json:"modifier_name,omitempty"`
	// ModifierContactEmail is the modifier's contact email; same
	// caveat as OwnerContactEmail.
	ModifierContactEmail string `json:"modifier_contact_email,omitempty"`

	// --- Mine-only flags.

	// Monitored is true when the user has opted in to per-event
	// notifications for the repo (web-app feature; surfaced for
	// completeness).
	Monitored bool `json:"monitored,omitempty"`
	// Starred is true when the user has favorited the repo. Affects
	// the LarePass left-nav ordering only; the CLI doesn't use it.
	Starred bool `json:"starred,omitempty"`
	// Salt is a Seafile internal field; non-empty for encrypted
	// repos. Pass-through.
	Salt string `json:"salt,omitempty"`

	// --- Shared-variant columns (only populated for type=share_to_me
	// or type=shared).

	// SharePermission is the permission granted via the share, in
	// "rw" / "r" form. For shared variants this replaces Permission.
	SharePermission string `json:"share_permission,omitempty"`
	// ShareType is "personal" / "group" / etc. — the Seafile share
	// taxonomy. Pass-through.
	ShareType string `json:"share_type,omitempty"`
	// UserEmail is the email of the OTHER party in the share — the
	// user it was shared TO (for type=shared) or the user who shared
	// it with the caller (for type=share_to_me).
	UserEmail string `json:"user_email,omitempty"`
	// UserName is the display name of the other party.
	UserName string `json:"user_name,omitempty"`
	// ContactEmail is the other party's contact email.
	ContactEmail string `json:"contact_email,omitempty"`
	// IsAdmin is true when the share grants admin rights (rare;
	// usually only set for repos shared from a group the caller
	// administers).
	IsAdmin bool `json:"is_admin,omitempty"`
}

// Client is the per-FilesURL handle for repos calls. HTTPClient is
// the factory-provided client whose refreshingTransport injects
// X-Authorization (NOT Authorization: Bearer — see pkg/cmdutil/factory.go)
// and refreshes on 401/403.
//
// Cheap to construct; reuse one per `files repos` invocation.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string // FilesURL, e.g. https://files.alice.olares.com
}

// HTTPError carries the status + truncated body of a non-2xx response
// in the same shape the upload / download / share packages use, so the
// cobra layer can re-use its 401 / 403 / 404 reformatter pattern.
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

// reposEnvelope is the {repos: [...]} response shape. The endpoint
// can also surface a top-level error envelope (`{code, message, ...}`
// with non-zero code) — we honor that here so a server-side 200 with
// a business-logic failure reads as an error instead of silently
// returning an empty list.
type reposEnvelope struct {
	Repos   []Repo `json:"repos"`
	Code    *int   `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// List fetches the slice of repos selected by `kind`. Returns an
// empty (non-nil) slice when the server replies with no repos so
// callers can range over the result without a length check.
//
// Wire shape:
//
//   - kind == TypeMine        → GET /api/repos/        (no `type` query)
//   - kind == TypeSharedToMe  → GET /api/repos/?type=share_to_me
//   - kind == TypeShared      → GET /api/repos/?type=shared
//
// Mirrors fetchMineRepo / fetchtosharedRepo / fetchsharedRepo at
// apps/.../api/files/v2/sync/utils.ts L93-135.
func (c *Client) List(ctx context.Context, kind Type) ([]Repo, error) {
	endpoint := c.BaseURL + "/api/repos/"
	if kind != "" && kind != TypeMine {
		q := url.Values{}
		q.Set("type", string(kind))
		endpoint += "?" + q.Encode()
	}
	var env reposEnvelope
	if err := c.do(ctx, http.MethodGet, endpoint, &env); err != nil {
		return nil, err
	}
	if env.Code != nil && *env.Code != 0 {
		return nil, fmt.Errorf("repos List: server rejected (code %d): %s", *env.Code, env.Message)
	}
	if env.Repos == nil {
		return []Repo{}, nil
	}
	return env.Repos, nil
}

// ListAll fans out to all three flavors and concatenates the results,
// in the order: mine → shared_to_me → shared. Useful for the CLI's
// `--type all` mode.
//
// On any error we return whatever we collected so far AND the error,
// so a partial fan-out (e.g. shared_to_me succeeds but `shared`
// times out) still surfaces useful data — the caller can choose to
// print + warn vs. abort hard.
func (c *Client) ListAll(ctx context.Context) ([]Repo, error) {
	var out []Repo
	for _, kind := range []Type{TypeMine, TypeSharedToMe, TypeShared} {
		batch, err := c.List(ctx, kind)
		if err != nil {
			return out, fmt.Errorf("list %s: %w", kind, err)
		}
		out = append(out, batch...)
	}
	if out == nil {
		return []Repo{}, nil
	}
	return out, nil
}

// Get returns the single repo whose RepoID matches the given id, or
// (nil, nil) if none of the three lists contains it. Implemented as a
// fan-out + filter rather than a dedicated endpoint because the
// per-user files-backend doesn't expose a "single repo" GET — the web
// app does the same client-side filter.
//
// Searches in the natural order (mine → shared_to_me → shared) and
// returns on the first hit so we don't keep pulling lists once the
// answer is found.
func (c *Client) Get(ctx context.Context, repoID string) (*Repo, error) {
	if repoID == "" {
		return nil, errors.New("repos Get: empty repoID")
	}
	for _, kind := range []Type{TypeMine, TypeSharedToMe, TypeShared} {
		batch, err := c.List(ctx, kind)
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", kind, err)
		}
		for i := range batch {
			if batch[i].RepoID == repoID {
				return &batch[i], nil
			}
		}
	}
	return nil, nil
}

// mutationEnvelope is the response shape for the write verbs
// (POST / PATCH / DELETE on /api/repos/). The per-user files-backend
// is consistent across these endpoints:
//
//   - On success the response either carries a {code: 0, ...payload}
//     envelope (most common — the LarePass front-end relies on it for
//     the response interceptor's error toast at fetch.ts L118-133),
//     or omits `code` entirely (which we treat as success).
//   - On business-logic failure the server returns HTTP 200 with a
//     non-zero `code` and a human-readable `message`. We promote
//     these to a Go error so the cobra layer can surface them
//     verbatim.
//
// The embedded Repo lets Create decode the new library's metadata
// directly (the create response carries repo_id / repo_name on the
// happy path), without a second envelope type. For rename / delete
// we only consult the code/message fields and ignore the payload.
type mutationEnvelope struct {
	Code    *int   `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Repo
}

// Create provisions a new Sync (Seafile) library named `name` on the
// per-user files-backend.
//
// Wire shape (mirrors createLibrary in apps/.../api/files/v2/sync/utils.ts):
//
//	POST /api/repos/?repoName=<url-encoded-name>
//	body: empty
//
// Returns the freshly created repo's metadata (notably RepoID, which
// the caller will use as the `<extend>` segment of subsequent
// sync/<repo_id>/ paths). Empty `name` is rejected client-side so a
// trivial typo doesn't round-trip to the server.
//
// Encryption is NOT supported here: the back-end's createLibrary call
// has no password / encryption flags, and the LarePass UI doesn't
// expose them either. Encrypted libraries must be created from the
// LarePass app or the Seahub web UI.
func (c *Client) Create(ctx context.Context, name string) (*Repo, error) {
	if name == "" {
		return nil, errors.New("repos Create: empty repo name")
	}
	q := url.Values{}
	q.Set("repoName", name)
	endpoint := c.BaseURL + "/api/repos/?" + q.Encode()

	var env mutationEnvelope
	if err := c.do(ctx, http.MethodPost, endpoint, &env); err != nil {
		return nil, err
	}
	if env.Code != nil && *env.Code != 0 {
		return nil, fmt.Errorf("repos Create: server rejected (code %d): %s", *env.Code, env.Message)
	}
	if env.Repo.RepoID == "" {
		// Server returned success but no repo metadata — rare but
		// possible if the Seahub adapter strips the payload. Surface
		// a clean error so the caller can re-list to recover the id
		// rather than silently returning an empty Repo.
		return nil, fmt.Errorf("repos Create: server accepted the request but did not return a repo_id (try `files repos list` to confirm)")
	}
	out := env.Repo
	return &out, nil
}

// Rename changes the display name of `repoID` to `newName`. The
// repo's UUID (`<extend>` segment of sync/<repo_id>/...) is stable
// across renames — already-cached front-end paths keep working.
//
// Wire shape (mirrors renameRepo in apps/.../api/files/v2/sync/utils.ts):
//
//	PATCH /api/repos/?destination=<new-name>&repoId=<repo-id>
//	body: empty
//
// Both arguments are required and rejected client-side when empty so
// we don't silently send `?destination=&repoId=...` and trip an
// unhelpful server-side error.
func (c *Client) Rename(ctx context.Context, repoID, newName string) error {
	if repoID == "" {
		return errors.New("repos Rename: empty repoID")
	}
	if newName == "" {
		return errors.New("repos Rename: empty new name")
	}
	q := url.Values{}
	q.Set("destination", newName)
	q.Set("repoId", repoID)
	endpoint := c.BaseURL + "/api/repos/?" + q.Encode()

	var env mutationEnvelope
	if err := c.do(ctx, http.MethodPatch, endpoint, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("repos Rename %s -> %q: server rejected (code %d): %s",
			repoID, newName, *env.Code, env.Message)
	}
	return nil
}

// Delete tears down `repoID` and all of its contents on the per-user
// files-backend. This is destructive: there is no client-side
// undo, and the server-side trash window depends on the Seafile
// deployment's retention policy (the CLI does not expose a restore
// verb).
//
// Wire shape (mirrors deleteRepo in apps/.../api/files/v2/sync/utils.ts):
//
//	DELETE /api/repos/?repoId=<repo-id>
//
// Empty repoID is rejected client-side — sending DELETE
// /api/repos/?repoId= on the wire is a meaningless request and would
// produce an opaque 4xx from the server.
func (c *Client) Delete(ctx context.Context, repoID string) error {
	if repoID == "" {
		return errors.New("repos Delete: empty repoID")
	}
	q := url.Values{}
	q.Set("repoId", repoID)
	endpoint := c.BaseURL + "/api/repos/?" + q.Encode()

	var env mutationEnvelope
	if err := c.do(ctx, http.MethodDelete, endpoint, &env); err != nil {
		return err
	}
	if env.Code != nil && *env.Code != 0 {
		return fmt.Errorf("repos Delete %s: server rejected (code %d): %s",
			repoID, *env.Code, env.Message)
	}
	return nil
}

// do is the shared HTTP machinery for repos calls: fire the request,
// surface non-2xx as *HTTPError, decode the response into out (which
// must be a pointer to reposEnvelope / mutationEnvelope or a
// structurally compatible type). When out is nil the body is
// drained and discarded — useful for callers that don't care about
// the payload (none of the current verbs hit this path, but it
// keeps `do` honest as a primitive).
//
// The body is read fully even on the discard path so the underlying
// HTTP/1.1 connection can be returned to the keep-alive pool; with
// Resumable.js' chunked uploads + repos calls happening over the
// same client, leaking connections here adds up fast.
func (c *Client) do(ctx context.Context, method, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

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
	trim := strings.TrimSpace(string(raw))
	if out == nil || trim == "" {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		if acceptPlainTextMutationSuccess(method, raw) {
			// Some Seahub / proxy deployments return HTTP 200 with a bare
			// ASCII word such as "success" instead of a JSON object for
			// PATCH/DELETE/POST on /api/repos/ — the axios stack in the
			// web app never hit this in dev (response.data is a string?),
			// but the CLI's strict JSON decode must tolerate it.
			return nil
		}
		return fmt.Errorf("decode response: %w (body: %s)", err, truncateBody(raw))
	}
	return nil
}

// acceptPlainTextMutationSuccess returns true when the body is a
// non-JSON success marker that some gateways emit for successful
// write calls. Only GET /api/repos/ (list) is required to be JSON; we
// must not use this for GET, or a misconfigured server could mask a
// broken list response.
func acceptPlainTextMutationSuccess(method string, raw []byte) bool {
	if method == http.MethodGet {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(string(raw)))
	switch s {
	case "success", "ok", "true", "1":
		return true
	default:
		return false
	}
}

// truncateBody renders a (possibly large) response body for inclusion
// in error messages without dumping multi-KB blobs into the user's
// terminal. The 500-byte cap matches HTTPError.Error.
func truncateBody(b []byte) string {
	const max = 500
	if len(b) > max {
		return string(b[:max]) + "...(truncated)"
	}
	return string(b)
}
