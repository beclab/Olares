// Package clusterclient is the HTTP wrapper olares-cli's `cluster` command
// tree uses to talk to the per-user ControlHub BFF
// (`<rp.ControlHubURL>`). It is the moral counterpart of
// cli/cmd/ctl/market/client.go's MarketClient and
// cli/cmd/ctl/settings/client.go's SettingsClient — a thin Doer that
// delegates auth to the caller's http.Client (Factory's
// refreshingTransport injects X-Authorization and auto-rotates expired
// access_tokens via /api/refresh) and just maps Go method calls onto
// JSON HTTP requests.
//
// What this package deliberately does NOT do:
//
//   - It does NOT bake in any URL prefix. ControlHub fans out to multiple
//     prefixes ("/api/v1", "/apis/<group>/<version>", "/kapis/...",
//     "/capi/...", "/middleware/..." — see pkg/olares/id.go::ControlHubURL).
//     Callers pass full paths.
//
//   - It does NOT decode envelopes itself. ControlHub speaks at least
//     three wire shapes (KubeSphere {items,totalItems}, K8s native
//     {kind,apiVersion,metadata,...}, ControlHub /capi/* custom). Each
//     verb decides which envelope to expect and uses the helpers in
//     decode.go. The Client's DoJSON signature only knows "send JSON,
//     decode JSON into out" — same contract as whoami.Doer so the
//     `cluster context` command can hand a Client to clusterctx.Run.
//
//   - It does NOT do any client-side permission gating. ControlHub
//     scopes resources by the caller's token server-side; CLI verbs
//     MUST defer to that. 401/403 from the server are wrapped via
//     reformatClusterAuthErr to keep the recovery CTA consistent across
//     verbs (`olares-cli profile login`).
package clusterclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// Client is the per-process handle the `cluster` tree uses for ControlHub
// HTTP. Construct it once per command via NewClient(hc, rp); the http.Client
// must be the Factory-provided one whose RoundTripper is
// cmdutil.refreshingTransport so X-Authorization and refresh-on-401 happen
// transparently.
type Client struct {
	httpClient *http.Client
	baseURL    string

	// olaresID is captured for diagnostics (401/403 reformatter mentions
	// it in the CTA). Never used for permission decisions.
	olaresID string
}

// NewClient builds a Client targeting rp.ControlHubURL with the supplied
// http.Client. The base URL is normalized (no trailing slash) so callers
// can prepend "/capi/...", "/kapis/...", "/api/v1/..." without doubling
// slashes.
func NewClient(hc *http.Client, rp *credential.ResolvedProfile) *Client {
	return &Client{
		httpClient: hc,
		baseURL:    strings.TrimRight(rp.ControlHubURL, "/"),
		olaresID:   rp.OlaresID,
	}
}

// BaseURL exposes the resolved ControlHub base URL (no trailing slash).
// Useful for verbs that want to log the host they're talking to or build
// query strings on top.
func (c *Client) BaseURL() string { return c.baseURL }

// OlaresID is the resolved profile's OlaresID. Surfaced so verbs can put
// it into their own diagnostic messages without stashing it separately.
func (c *Client) OlaresID() string { return c.olaresID }

// DoJSON sends an HTTP request with an optional JSON body and decodes the
// JSON response body into `out` (when non-nil). It satisfies
// pkg/whoami.Doer so the `cluster context` command can reuse the whoami
// package's Run-driver shape.
//
// path is appended to BaseURL verbatim — caller picks "/capi/...",
// "/kapis/...", "/api/v1/..." based on the resource. body may be nil for
// GET/DELETE.
//
// Wire-format shape: this method only knows JSON. Envelope unwrapping
// (KubeSphere {items,totalItems}, K8s native, /capi/* custom) is the
// caller's responsibility — typically via the helpers in decode.go.
func (c *Client) DoJSON(ctx context.Context, method, path string, body, out interface{}) error {
	respBody, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s %s response: %w (body=%s)", method, c.baseURL+path, err, truncate(string(respBody), 200))
	}
	return nil
}

// DoRaw sends an HTTP request with an optional JSON body and returns the
// raw response body bytes. Used by decode helpers that need to peek into
// the wire shape before deciding how to parse (e.g. K8s API can return
// either a typed object or a metav1.Status on error), and by `cluster pod
// yaml` which forwards bytes to the user as-is.
//
// All non-2xx handling (401/403 reformat, generic HTTP error) happens
// before this returns, so a successful return always carries a 2xx body.
func (c *Client) DoRaw(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	return c.do(ctx, method, path, body)
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Origin / Referer left at Go defaults — see SettingsClient.DoJSON
	// (cli/cmd/ctl/settings/client.go) for the Authelia CSRF rationale.
	// ControlHub rides the same edge, so the same constraint applies.

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Factory's refreshingTransport surfaces typed credential
		// errors when /api/refresh itself fails. http.Client wraps
		// them in *url.Error; errors.As walks the Unwrap chain so
		// callers see the canonical "run profile login" CTA instead
		// of `Get "...": refresh token for ... became invalid at ...`.
		// Mirrors cli/cmd/ctl/{files/download.go,market/client.go,
		// settings/client.go}.
		var inv *credential.ErrTokenInvalidated
		if errors.As(err, &inv) {
			return nil, inv
		}
		var nli *credential.ErrNotLoggedIn
		if errors.As(err, &nli) {
			return nil, nli
		}
		return nil, fmt.Errorf("%s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, reformatClusterAuthErr(resp.StatusCode, respBody, c.olaresID)
	}
	if resp.StatusCode/100 != 2 {
		return nil, formatHTTPErr(method, url, resp.StatusCode, respBody)
	}
	return respBody, nil
}

// reformatClusterAuthErr mirrors files/download.go::reformatHTTPError
// 401/403 branch and market/client.go::reformatMarketAuthErr — same
// recovery CTA across all per-user ingress points so the user only has
// to memorize one trick (`olares-cli profile login`).
//
// Note: ControlHub itself does some path-level RBAC (a non-admin user
// hitting `/api/v1/nodes` will get 403 even with a valid token). We do
// NOT try to disambiguate "token bad" from "role insufficient" here:
// the body is included verbatim so the user can read it, and the action
// (`profile login` then `cluster context --refresh`) is appropriate for
// both modes.
func reformatClusterAuthErr(status int, respBody []byte, olaresID string) error {
	body := strings.TrimSpace(string(respBody))
	if len(body) > 200 {
		body = body[:200]
	}
	if olaresID != "" {
		if body != "" {
			return fmt.Errorf("server rejected the request (HTTP %d: %s); please run: olares-cli profile login --olares-id %s",
				status, body, olaresID)
		}
		return fmt.Errorf("server rejected the request (HTTP %d); please run: olares-cli profile login --olares-id %s",
			status, olaresID)
	}
	return fmt.Errorf("server rejected the request (HTTP %d); please re-run `olares-cli profile login`", status)
}

// formatHTTPErr handles non-401/403 non-2xx responses. ControlHub forwards
// upstream errors verbatim — KubeSphere returns `{message, status}`,
// kube-apiserver returns metav1.Status `{kind, apiVersion, status,
// message, reason, code}`, /capi/* returns plain text on some failures.
// Try the structured shapes first and fall back to a body-truncated dump.
func formatHTTPErr(method, url string, status int, body []byte) error {
	var k8sStatus struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
		Code    int    `json:"code"`
	}
	if err := json.Unmarshal(body, &k8sStatus); err == nil && k8sStatus.Message != "" {
		// metav1.Status path: prefer Reason+Message which together
		// describe both "what kind of failure" and "why".
		if k8sStatus.Reason != "" {
			return fmt.Errorf("%s %s: HTTP %d (%s): %s", method, url, status, k8sStatus.Reason, k8sStatus.Message)
		}
		return fmt.Errorf("%s %s: HTTP %d: %s", method, url, status, k8sStatus.Message)
	}
	var generic struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	if err := json.Unmarshal(body, &generic); err == nil {
		switch {
		case generic.Error != "":
			return fmt.Errorf("%s %s: HTTP %d: %s", method, url, status, generic.Error)
		case generic.Message != "":
			return fmt.Errorf("%s %s: HTTP %d (code=%d): %s", method, url, status, generic.Code, generic.Message)
		}
	}
	return fmt.Errorf("%s %s: HTTP %d: %s", method, url, status, truncate(string(body), 500))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
