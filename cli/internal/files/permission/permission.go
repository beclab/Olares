// Package permission implements the wire side of `olares-cli files chown`,
// which talks to the per-user files-backend's `/api/permission/...`
// surface to read and update the POSIX owner UID of a file or
// directory.
//
// On the LarePass web app this lives in the file-properties dialog
// (apps/packages/app/src/components/files/prompts/InfoDialog.vue) under
// the "Permission" tab — a `q-select` with two options (Root=0,
// User=1000) and a "recursive" toggle. The submit handler calls
// `operationStore.setPermission(file, uid, recursive)` (operation.ts
// L912-930), which boils down to:
//
//	GET  /api/permission/<fileType>/<extend><subPath>          → {uid: <int>}
//	PUT  /api/permission/<fileType>/<extend><subPath>?uid=<int>[&recursive=1]
//
// The "permission" segment after `/api/` matches every other
// CommonUrlApiType in the web app (resources, paste, raw, md5,
// preview, repos, ...) — see commonUrlPrefix in
// apps/.../api/files/v2/common/utils.ts.
//
// Per-namespace URL shape mirrors the per-driveType `getDiskPath`
// helpers:
//
//   - drive (Home):  /api/permission/drive/Home/<sub>
//   - drive (Data):  /api/permission/drive/Data/<sub>
//   - cache:         /api/permission/cache/<node>/<sub>
//
// Other namespaces (sync / external / awss3 / dropbox / google /
// tencent) are gated client-side (see `permissionInDriveType` in the
// web app — it lists Drive, Data, Cache only). Their `getDiskPath`
// implementations may technically build a URL on the wire, but the
// GUI never exposes the affordance and the server's behavior at those
// paths is not part of the contract this CLI implements.
//
// Wire shape vs. share package: this surface uses bare HTTP — no
// {code, message, data} envelope. GET returns a flat `{uid}` body and
// PUT returns a small status object. We keep decoding minimal so a
// future server-side field addition doesn't fail to round-trip.
package permission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/beclab/Olares/cli/internal/files/encodepath"
)

// SupportedFileTypes is the set of front-end fileType values this
// surface is willing to address. Mirrors `permissionInDriveType` in
// apps/.../components/files/prompts/InfoDialog.vue (DriveType.Drive,
// DriveType.Data, DriveType.Cache; Drive and Data both wire to
// fileType="drive" so the set collapses to two entries here).
//
// External / sync / cloud are intentionally absent: the LarePass GUI
// hides the Permission tab for those namespaces, and the server-side
// behavior under `/api/permission/external/...` / `/api/permission/sync/...`
// is not a documented part of the contract. The cobra layer rejects
// these client-side with a self-describing error.
var SupportedFileTypes = map[string]struct{}{
	"drive": {},
	"cache": {},
}

// Target identifies the resource whose ownership we want to read or
// write. Same shape as the share / cp packages' Target so the cobra
// layer can pass through a parsed FrontendPath without an
// intermediate conversion struct.
//
// IsDirIntent is preserved purely to keep the wire URL agreeing with
// what the caller typed (a trailing '/' on the input survives into
// the request URL); the per-resource permission endpoint does not
// distinguish file vs. dir routing the way /api/resources does, but
// keeping the slash semantically consistent avoids surprises if the
// backend ever does start to care.
type Target struct {
	FileType    string
	Extend      string
	SubPath     string
	IsDirIntent bool
}

// String renders the target as `<fileType>/<extend><subPath>` —
// suitable for error messages and progress output. Matches the way
// FrontendPath.String() formats the same data.
func (t Target) String() string {
	sub := t.SubPath
	if sub == "" {
		sub = "/"
	} else if !strings.HasPrefix(sub, "/") {
		sub = "/" + sub
	}
	return t.FileType + "/" + t.Extend + sub
}

// HTTPError carries the status + truncated body of a non-2xx
// response, same shape as share.HTTPError / rename.HTTPError so the
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

// Client is the per-FilesURL handle for permission calls. HTTPClient
// is the factory-provided client whose refreshingTransport injects
// X-Authorization and refreshes on 401/403, same as the share / cp
// clients in this codebase.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// getResponse mirrors the body shape of GET /api/permission/<...>
// observed via `operationStore.getPermission` in the web app:
// the server replies with `{ uid: <int> }`. Server-side
// implementations occasionally surface additional fields (e.g.
// `gid`, `mode`); we read only `uid` so future additions don't break
// the decode, and ignore extras.
type getResponse struct {
	UID int `json:"uid"`
}

// Get reads the current owner UID of the target. The wire shape is
//
//	GET /api/permission/<fileType>/<extend><subPath>
//
// — same path the PUT endpoint uses minus the query string.
//
// Returns the parsed UID. A non-200 response is surfaced as
// *HTTPError so the cobra layer can branch on auth / not-found.
func (c *Client) Get(ctx context.Context, t Target) (int, error) {
	if t.FileType == "" || t.Extend == "" {
		return 0, fmt.Errorf("permission Get: empty fileType or extend (got %q/%q)", t.FileType, t.Extend)
	}
	endpoint := c.BaseURL + buildPermissionURL(t)
	var resp getResponse
	if err := c.do(ctx, http.MethodGet, endpoint, nil, &resp); err != nil {
		return 0, err
	}
	return resp.UID, nil
}

// Set writes a new owner UID to the target. Wire shape:
//
//	PUT /api/permission/<fileType>/<extend><subPath>?uid=<int>[&recursive=1]
//
// The request body is an empty JSON object `{}` — the server expects
// the actual parameters in the query string, matching the LarePass
// web app's `setPermission` helper exactly (operation.ts L912-930).
//
// `recursive=1` (note: the GUI sends literal "1", not "true") applies
// the change to every entry beneath the target as well; without it,
// only the named entry's UID is updated. The server treats anything
// truthy the same way, but we mimic the GUI verbatim for symmetry.
//
// Returns nil on 2xx, *HTTPError otherwise.
func (c *Client) Set(ctx context.Context, t Target, uid int, recursive bool) error {
	if t.FileType == "" || t.Extend == "" {
		return fmt.Errorf("permission Set: empty fileType or extend (got %q/%q)", t.FileType, t.Extend)
	}
	q := url.Values{}
	q.Set("uid", strconv.Itoa(uid))
	if recursive {
		q.Set("recursive", "1")
	}
	endpoint := c.BaseURL + buildPermissionURL(t) + "?" + q.Encode()
	// The GUI sends an explicit empty object — keep it byte-for-byte
	// so any server-side body length checks behave the same way as
	// they do for LarePass requests.
	return c.do(ctx, http.MethodPut, endpoint, []byte("{}"), nil)
}

// do is the shared HTTP machinery: marshal a body if present, set
// Content-Type / Accept, fire the request, decode the response into
// `out` (a pointer to a struct, or nil to discard the body).
//
// Cribbed from share.Client.do — the per-resource permission
// endpoints are simpler (no envelope) so we collapse the decoding
// to a single optional Unmarshal.
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
	if out == nil || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w (body: %s)", err, truncateBody(raw))
	}
	return nil
}

// buildPermissionURL is the wire URL builder for the
// /api/permission/<fileType>/<extend><subPath> endpoint, used by both
// Get and Set.
//
// Per-segment percent-encoding via encodepath.EncodeURL keeps the
// CLI byte-identical to LarePass — the same `encodeUrl` helper from
// apps/packages/app/src/utils/encode.ts that share / cp / rm /
// upload all defer to.
//
// The trailing slash policy mirrors `/api/share/share_path/...`:
// always end the path in '/' so the backend's resource handler
// matches the directory route shape it uses elsewhere. Even for
// file targets the GUI sends a trailing slash on this endpoint
// (the path comes out of `getDiskPath` which preserves the
// caller's slash); we do the same so a hypothetical future
// server-side router doesn't diverge.
//
// Defense in depth: if SubPath is empty (caller bypassed the
// FrontendPath parser) we synthesize "/" so we never emit
// "drive/HomeDocuments" without a separator.
func buildPermissionURL(t Target) string {
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
	return "/api/permission/" + enc
}

// IsSupported reports whether `fileType` is one of the namespaces
// the LarePass Permission tab is willing to talk to (drive / cache).
// Callers use this to fail fast client-side before assembling a
// Target — symmetric with share.go's `shareFlavorAllowedNamespaces`.
func IsSupported(fileType string) bool {
	_, ok := SupportedFileTypes[fileType]
	return ok
}

// SupportedFileTypesList returns a stable, sorted, comma-joined
// rendering of SupportedFileTypes for error messages.
func SupportedFileTypesList() string {
	out := make([]string, 0, len(SupportedFileTypes))
	for k := range SupportedFileTypes {
		out = append(out, k)
	}
	sortStrings(out)
	return strings.Join(out, ", ")
}

// sortStrings is a tiny dependency-free shim so this package doesn't
// pull in `sort` for the single error-path call site. Kept private.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
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

// Compile-time assertion: keep the public error surface stable so the
// cobra layer's `errors.As(err, &*HTTPError{})` branches don't break
// under refactors.
var _ error = (*HTTPError)(nil)
