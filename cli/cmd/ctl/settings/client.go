package settings

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

// SettingsClient talks to user-service / BFL via the desktop ingress at
// https://desktop.<terminusName>. It is the moral counterpart of
// cli/cmd/ctl/market/client.go's MarketClient: a thin HTTP wrapper that
// delegates auth to the caller's http.Client (Factory's refreshingTransport
// injects X-Authorization and auto-rotates expired access_tokens via
// /api/refresh — see cli/pkg/cmdutil/factory.go) and otherwise just maps
// Go method calls onto JSON HTTP requests.
//
// Two URL prefixes ride this same host:
//   - "/api/...", "/server/...", "/headscale/..." → user-service :3010
//   - "/apis/backup/v1/..."                       → BFL backup-server
//
// Both are reached via the same desktop origin (see
// apps/packages/app/src/stores/desktop/token.ts and the SPA's backup store at
// apps/packages/app/src/api/settings/backup.ts) — no extra host derivation
// needed, just two path prefixes. We don't bake either prefix into the
// client; callers pass full paths so backup/restore code in Phase 6 doesn't
// need a second client type.
type SettingsClient struct {
	httpClient *http.Client
	baseURL    string

	// Identity bits captured from the resolved profile, used by:
	//   - 401/403 reformatting (CTA mentions OlaresID),
	//   - whoami's eventual cache write in Phase 0b.
	olaresID    string
	accessToken string
}

// NewSettingsClient builds a SettingsClient from a factory-provided
// http.Client (already wired with X-Authorization injection) and a resolved
// profile. The base URL is rp.DesktopURL — derived once by buildResolved in
// pkg/credential/default_provider.go from the OlaresID's terminus name.
func NewSettingsClient(hc *http.Client, rp *credential.ResolvedProfile) *SettingsClient {
	return &SettingsClient{
		httpClient:  hc,
		baseURL:     strings.TrimRight(rp.DesktopURL, "/"),
		olaresID:    rp.OlaresID,
		accessToken: rp.AccessToken,
	}
}

// BaseURL exposes the desktop ingress base URL for callers that build their
// own paths (e.g. Phase 0b's whoami helper, Phase 6's backup-server helpers).
// Returned without a trailing slash so callers can prepend "/api/..." or
// "/apis/backup/v1/..." without doubling slashes.
func (c *SettingsClient) BaseURL() string { return c.baseURL }

// OlaresID is the resolved profile's OlaresID. Surfaced for diagnostics —
// e.g. the 401/403 reformatter wants to suggest "olares-cli profile login
// --olares-id <id>".
func (c *SettingsClient) OlaresID() string { return c.olaresID }

// DoJSON sends a JSON request and decodes a JSON response body into `out` (or
// discards the body if out is nil). On 401/403 it surfaces the standard
// "profile login" CTA via reformatSettingsAuthErr; other non-2xx responses
// are returned as a generic HTTP error wrapping the body for triage.
//
// path is appended to BaseURL verbatim — caller picks "/api/..." vs
// "/apis/backup/v1/..." based on the resource. body may be nil for GET/DELETE.
func (c *SettingsClient) DoJSON(ctx context.Context, method, path string, body, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// 2026-04-28 KI-12 / KI-15 / KI-16 reverted: an earlier draft
	// also injected `Origin: <baseURL>` and `Referer: <baseURL>/`
	// to mimic the browser/SPA shape. That caused a global
	// regression — Authelia's CSRF/Origin gate on the desktop
	// ingress takes the presence of an Origin header on a Go-style
	// non-browser request as a sign of a forged request and
	// redirects to login → 400 Bad Request HTML on every path that
	// previously worked (31/34 verbs flipped from ok to fail). Until
	// we have an actual Authelia config dump showing what shape it
	// expects, we leave Origin/Referer unset (Go's default) so the
	// /api/* paths keep working.

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// The Factory's refreshingTransport may surface a typed
		// credential error when /api/refresh itself fails (the grant is
		// dead, or no token is stored at all). http.Client wraps it
		// inside *url.Error, but errors.As walks the Unwrap chain — pull
		// it out so the caller sees the canonical "run profile login"
		// CTA instead of `GET https://...: Get "https://...": refresh
		// token for ... became invalid at ...`. Mirrors the unwrapping
		// done in cli/cmd/ctl/files/download.go and market/client.go.
		var inv *credential.ErrTokenInvalidated
		if errors.As(err, &inv) {
			return inv
		}
		var nli *credential.ErrNotLoggedIn
		if errors.As(err, &nli) {
			return nli
		}
		return fmt.Errorf("%s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return reformatSettingsAuthErr(resp.StatusCode, respBody, c.olaresID)
	}

	if resp.StatusCode/100 != 2 {
		return formatHTTPErr(method, url, resp.StatusCode, respBody)
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s %s response: %w (body=%s)", method, url, err, truncate(string(respBody), 200))
	}
	return nil
}

// reformatSettingsAuthErr mirrors files/download.go's formatHTTPError 401/403
// branch and market/client.go's reformatMarketAuthErr: turn the edge proxy's
// "you are not authenticated" into the same actionable CTA the rest of the
// CLI uses, so the user's troubleshooting story is consistent across verbs.
func reformatSettingsAuthErr(status int, respBody []byte, olaresID string) error {
	body := strings.TrimSpace(string(respBody))
	if len(body) > 200 {
		body = body[:200]
	}
	if olaresID != "" {
		if body != "" {
			return fmt.Errorf("server rejected the access token (HTTP %d: %s); please run: olares-cli profile login --olares-id %s",
				status, body, olaresID)
		}
		return fmt.Errorf("server rejected the access token (HTTP %d); please run: olares-cli profile login --olares-id %s",
			status, olaresID)
	}
	return fmt.Errorf("server rejected the access token (HTTP %d); please re-run `olares-cli profile login`", status)
}

// formatHTTPErr handles non-401/403 non-2xx responses. user-service and BFL
// both speak loose JSON — sometimes {"error":"..."}, sometimes
// {"code":1,"message":"..."}, sometimes plain text from the edge proxy. We
// try the structured shapes first and fall through to a body-truncated raw
// dump so the user always gets *something* to grep.
func formatHTTPErr(method, url string, status int, body []byte) error {
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
