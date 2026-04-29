package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// debugHTTPEnabled returns true when the operator opted into HTTP wire-
// dumping via `OLARES_CLI_DEBUG_HTTP=1`. Used by `do` to emit one stderr
// line per request and one per response (status + truncated body) so a
// human can compare the CLI's actual wire shape against a SPA-side
// devtools capture without rebuilding the binary.
//
// The check is per-call rather than process-wide so a parent shell can
// flip it on / off mid-session via `env`.
func debugHTTPEnabled() bool {
	v := strings.TrimSpace(os.Getenv("OLARES_CLI_DEBUG_HTTP"))
	return v != "" && v != "0" && strings.ToLower(v) != "false"
}

// Client is a thin wrapper around the factory's authenticated *http.Client.
// Every command in the dashboard tree obtains one via the area-local
// `prepareClient` declared in each cmd subpackage's `common.go`
// (cli/cmd/ctl/dashboard/<area>/common.go) and calls DoJSON / DoEmpty
// / Get* helpers; the underlying refreshingTransport handles X-Authorization
// injection + transparent /api/refresh on 401/403.
//
// Why a typed Client (vs. a raw http.Client + URL constants per file)?
// Because we need to reformat 4xx/5xx into agent-friendly errors, expose an
// `EnsureUser` cache (used by every --user-aware command), and keep base-URL
// stitching in one place for the schemas / e2e tests to mock.
type Client struct {
	hc       *http.Client
	baseURL  string // <DashboardURL>, no trailing slash
	olaresID string

	// userOnce / userInfo / userErr cache the result of EnsureUser. We
	// only ever talk to /capi/app/detail once per command invocation
	// regardless of how many `--user`-aware leaf commands run inside an
	// aggregated `dashboard overview` invocation.
	userOnce sync.Once
	userInfo *UserDetail
	userErr  error

	// systemStatusOnce / systemStatus / systemStatusErr cache the
	// result of EnsureSystemStatus. The dashboard fan / gpu subtrees
	// gate themselves on the device profile (Olares One vs. generic
	// box, CUDA-capable vs. not). Cached for the lifetime of the
	// Client to keep `dashboard overview` aggregations cheap.
	systemStatusOnce sync.Once
	systemStatus     *SystemStatus
	systemStatusErr  error
}

// NewClient builds a Client from a factory-provided http.Client (already
// wired with refreshingTransport) and the resolved profile. Strips a
// trailing "/" from rp.DashboardURL so callers can concatenate paths
// without juggling slashes.
func NewClient(hc *http.Client, rp *credential.ResolvedProfile) *Client {
	return &Client{
		hc:       hc,
		baseURL:  strings.TrimRight(rp.DashboardURL, "/"),
		olaresID: rp.OlaresID,
	}
}

// BaseURL returns the dashboard-BFF root, sans trailing slash.
func (c *Client) BaseURL() string { return c.baseURL }

// OlaresID returns the OlaresID of the active profile (for Meta.Profile and
// reformatted error messages).
func (c *Client) OlaresID() string { return c.olaresID }

// HTTPClient returns the underlying authenticated *http.Client. Surface area
// for callers that need to issue raw requests bypassing DoJSON/DoRaw — kept
// small on purpose so the auth + error-reformatting story stays centralised.
func (c *Client) HTTPClient() *http.Client { return c.hc }

// ----------------------------------------------------------------------------
// HTTP helpers
// ----------------------------------------------------------------------------

// HTTPError is the typed error every Do* helper returns for non-200
// responses. ErrorKind is the small enum surfaced via Meta.ErrorKind so
// agents can branch without parsing free-form text.
type HTTPError struct {
	Status    int
	URL       string
	Body      string // truncated to 512 bytes
	ErrorKind string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("HTTP %d from %s", e.Status, e.URL)
	}
	return fmt.Sprintf("HTTP %d from %s: %s", e.Status, e.URL, e.Body)
}

// IsHTTPError unwraps err looking for *HTTPError. Returns the typed error
// + true on hit; nil + false otherwise.
func IsHTTPError(err error) (*HTTPError, bool) {
	var he *HTTPError
	if errors.As(err, &he) {
		return he, true
	}
	return nil, false
}

// ClassifyTransportErr maps an arbitrary error returned by hc.Do into an
// ErrorKind enum. Used by runner.go to populate Meta.ErrorKind for failed
// iterations. Order matters: typed credential errors first, then HTTPError,
// then generic transport.
func ClassifyTransportErr(err error) string {
	if err == nil {
		return ""
	}
	var inv *credential.ErrTokenInvalidated
	if errors.As(err, &inv) {
		return "auth"
	}
	var nli *credential.ErrNotLoggedIn
	if errors.As(err, &nli) {
		return "auth"
	}
	var he *HTTPError
	if errors.As(err, &he) && he.ErrorKind != "" {
		return he.ErrorKind
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) {
		return "transport"
	}
	return "transport"
}

// DoJSON issues `method <baseURL><path>?<query>` with optional JSON body
// and decodes the 2xx response into `out`. `out` may be nil to discard.
//
// Behaviour on non-2xx:
//
//   - 401/403 → reformatted as "server rejected the access token (HTTP X)"
//     with the standard CTA. The factory's refreshingTransport already had
//     a chance to refresh+retry; reaching us means the grant is dead.
//   - 4xx     → *HTTPError with ErrorKind="http_4xx".
//   - 5xx     → *HTTPError with ErrorKind="http_5xx".
//
// A response body up to 512 bytes is captured into HTTPError.Body so error
// messages stay actionable without leaking arbitrary upstream payloads.
func (c *Client) DoJSON(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	resp, raw, err := c.do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := classifyStatus(resp.StatusCode, resp.Request.URL.String(), raw, c.olaresID); err != nil {
		return err
	}

	if out == nil {
		return nil
	}
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response from %s: %w", resp.Request.URL, err)
	}
	return nil
}

// DoEmpty is DoJSON for callers that just want to check the status, e.g.
// health probes. Body is read+discarded for connection re-use.
func (c *Client) DoEmpty(ctx context.Context, method, path string, query url.Values, body any) error {
	return c.DoJSON(ctx, method, path, query, body, nil)
}

// DoRaw is the escape hatch for endpoints that don't return JSON or where
// the caller wants to peek at the status before deciding how to decode
// (used by the GPU `404=no-integration` three-state). The returned body is
// already drained — caller treats it as []byte.
func (c *Client) DoRaw(ctx context.Context, method, path string, query url.Values, body any) (status int, payload []byte, err error) {
	resp, raw, err := c.do(ctx, method, path, query, body)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, raw, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any) (*http.Response, []byte, error) {
	endpoint, err := c.urlFor(path, query)
	if err != nil {
		return nil, nil, err
	}
	var (
		reader  io.Reader
		bodyBuf []byte
	)
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("encode request body: %w", err)
		}
		bodyBuf = buf
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	if reader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	debug := debugHTTPEnabled()
	if debug {
		// Outgoing line. Body is the marshalled payload pre-transport;
		// it does NOT yet carry refreshingTransport's X-Authorization
		// header (that's injected just before the wire). We log only
		// the body so the SPA-vs-CLI shape comparison is one-line.
		fmt.Fprintf(os.Stderr, "[olares-cli debug-http] → %s %s body=%s\n",
			method, endpoint, debugTruncate(bodyBuf, 512))
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "[olares-cli debug-http] ← %s %s ERROR=%v\n",
				method, endpoint, err)
		}
		// refreshingTransport surfaces typed credential errors here.
		// Surface them verbatim so the caller's reformatter sees the
		// canonical CTA instead of a wrapped url.Error.
		var inv *credential.ErrTokenInvalidated
		if errors.As(err, &inv) {
			return nil, nil, inv
		}
		var nli *credential.ErrNotLoggedIn
		if errors.As(err, &nli) {
			return nil, nil, nli
		}
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		// Best effort: drop the response, surface the read error.
		resp.Body.Close()
		if debug {
			fmt.Fprintf(os.Stderr, "[olares-cli debug-http] ← %s %s status=%d READ-ERROR=%v\n",
				method, endpoint, resp.StatusCode, readErr)
		}
		return nil, nil, fmt.Errorf("read response from %s: %w", endpoint, readErr)
	}
	if debug {
		fmt.Fprintf(os.Stderr, "[olares-cli debug-http] ← %s %s status=%d body=%s\n",
			method, endpoint, resp.StatusCode, debugTruncate(raw, 512))
	}
	return resp, raw, nil
}

// debugTruncate renders a payload as a single-line string (newlines
// collapsed, max `n` bytes) for the debug-http log. Stays under one
// terminal line even when HAMI returns a fat JSON document.
func debugTruncate(b []byte, n int) string {
	if len(b) == 0 {
		return "<empty>"
	}
	s := string(b)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		s = s[:n] + "…(truncated)"
	}
	return s
}

func (c *Client) urlFor(path string, query url.Values) (string, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	full := c.baseURL + path
	u, err := url.Parse(full)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint %s: %w", full, err)
	}
	if query != nil {
		// Merge with any query already on `path`. Last writer wins, which
		// matches what `req.URL.RawQuery = query.Encode()` would do.
		existing := u.Query()
		for k, vs := range query {
			existing[k] = vs
		}
		u.RawQuery = existing.Encode()
	}
	return u.String(), nil
}

func classifyStatus(status int, endpoint string, raw []byte, olaresID string) error {
	if status >= 200 && status < 300 {
		return nil
	}
	body := strings.TrimSpace(string(raw))
	if len(body) > 512 {
		body = body[:512]
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return reformatAuthErr(status, body, olaresID, endpoint)
	}
	kind := "http_5xx"
	if status >= 400 && status < 500 {
		kind = "http_4xx"
	}
	return &HTTPError{
		Status:    status,
		URL:       endpoint,
		Body:      body,
		ErrorKind: kind,
	}
}

// reformatAuthErr matches the wording used by the files / market trees so
// the troubleshooting story stays uniform across CLI commands.
func reformatAuthErr(status int, body, olaresID, endpoint string) error {
	if olaresID != "" {
		if body != "" {
			return fmt.Errorf("server rejected the access token (HTTP %d from %s: %s); please run: olares-cli profile login --olares-id %s",
				status, endpoint, body, olaresID)
		}
		return fmt.Errorf("server rejected the access token (HTTP %d from %s); please run: olares-cli profile login --olares-id %s",
			status, endpoint, olaresID)
	}
	return fmt.Errorf("server rejected the access token (HTTP %d from %s); please re-run `olares-cli profile login`", status, endpoint)
}

// ----------------------------------------------------------------------------
// EnsureUser — globalrole + user identity, lazily fetched once per Client.
// ----------------------------------------------------------------------------

// UserDetail captures the subset of /capi/app/detail the CLI cares about.
// The wire payload nests the identity inside `.user.{username,globalrole}`
// (matching `AppDetailResponse` in
// `controlPanelCommon/network/network.ts`); EnsureUser does the lifting so
// the rest of the package keeps a flat view.
type UserDetail struct {
	Name       string
	OlaresID   string
	GlobalRole string
}

// IsAdmin reports whether the resolved user has the platform-admin role.
// The server canonical value is `platform-admin`; we tolerate `admin` too
// for forward-compat with the role rename done in early Olares releases.
func (u *UserDetail) IsAdmin() bool {
	if u == nil {
		return false
	}
	switch u.GlobalRole {
	case "platform-admin", "admin":
		return true
	}
	return false
}

// EnsureUser fetches /capi/app/detail and caches the result for the lifetime
// of the Client. Subsequent callers — including the watch loop iterating —
// reuse the same result; the cache is never invalidated mid-process.
//
// Any HTTP error from /capi/app/detail is cached too; callers should not
// retry by re-calling EnsureUser. (Re-running the whole command is the
// supported path for a transient outage.)
func (c *Client) EnsureUser(ctx context.Context) (*UserDetail, error) {
	c.userOnce.Do(func() {
		var raw struct {
			ClusterRole string `json:"clusterRole"`
			User        struct {
				Username   string `json:"username"`
				GlobalRole string `json:"globalrole"`
				Email      string `json:"email"`
			} `json:"user"`
		}
		if err := c.DoJSON(ctx, http.MethodGet, "/capi/app/detail", nil, nil, &raw); err != nil {
			c.userErr = err
			return
		}
		c.userInfo = &UserDetail{
			Name:       raw.User.Username,
			GlobalRole: raw.User.GlobalRole,
		}
	})
	return c.userInfo, c.userErr
}

// RequireAdmin is a guard for `--user`-aware commands. It calls EnsureUser
// and returns a friendly error if the active profile is not an admin.
func (c *Client) RequireAdmin(ctx context.Context) (*UserDetail, error) {
	u, err := c.EnsureUser(ctx)
	if err != nil {
		return nil, err
	}
	if !u.IsAdmin() {
		return u, fmt.Errorf("--user requires platform-admin; %s does not have that role", u.Name)
	}
	return u, nil
}

// ----------------------------------------------------------------------------
// EnsureSystemStatus — device profile (Olares One vs. generic), cached.
// ----------------------------------------------------------------------------

// SystemStatus is the subset of `/user-service/api/system/status`'s
// payload the CLI uses to decide whether fan / gpu subtrees apply on
// this device. The wire shape is `data.{device_name, ...}`; we only
// keep the bits that drive the gates.
//
// Mirrors `TerminusStatus` in
// `packages/app/src/services/abstractions/mdns/service.ts:237` and the
// `device_name === 'Olares One'` check in
// `packages/app/src/apps/dashboard/stores/Fan.ts:67`.
type SystemStatus struct {
	DeviceName string
	HostName   string
	CPUInfo    string
	GPUInfo    string
}

// IsOlaresOne reports whether this Olares instance is running on an
// Olares One device. Returns the cached value from EnsureSystemStatus.
// On error from the underlying call we report `false` so the gates
// fail open (no fan section) rather than blocking with stale data.
func (s *SystemStatus) IsOlaresOne() bool {
	if s == nil {
		return false
	}
	return s.DeviceName == "Olares One"
}

// EnsureSystemStatus fetches `/user-service/api/system/status` once per
// Client and caches the trimmed `SystemStatus` view. Subsequent callers
// — including any nested `overview fan / overview gpu` invocations
// inside an aggregated `dashboard overview` — reuse the result.
//
// Behaviour on error mirrors EnsureUser: the error is cached so a
// failure is sticky for the rest of the process; re-running the whole
// command is the supported path for a transient outage.
func (c *Client) EnsureSystemStatus(ctx context.Context) (*SystemStatus, error) {
	c.systemStatusOnce.Do(func() {
		var raw struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				DeviceName string `json:"device_name"`
				HostName   string `json:"host_name"`
				CPUInfo    string `json:"cpu_info"`
				GPUInfo    string `json:"gpu_info"`
			} `json:"data"`
		}
		if err := c.DoJSON(ctx, http.MethodGet, "/user-service/api/system/status", nil, nil, &raw); err != nil {
			c.systemStatusErr = err
			return
		}
		c.systemStatus = &SystemStatus{
			DeviceName: raw.Data.DeviceName,
			HostName:   raw.Data.HostName,
			CPUInfo:    raw.Data.CPUInfo,
			GPUInfo:    raw.Data.GPUInfo,
		}
	})
	return c.systemStatus, c.systemStatusErr
}

// IsOlaresOne is a convenience wrapper for callers that just need the
// boolean. Returns false on any underlying error so gates fail open.
func (c *Client) IsOlaresOne(ctx context.Context) (bool, error) {
	s, err := c.EnsureSystemStatus(ctx)
	if err != nil {
		return false, err
	}
	return s.IsOlaresOne(), nil
}
