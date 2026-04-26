package whoami

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient is a Doer that talks to the desktop ingress directly. It's
// used by both `profile whoami` and `profile login`'s eager fetch (which
// can't import cli/cmd/ctl/settings because settings imports profile),
// and by the settings tree's me/users-me wrappers (which would otherwise
// have to thread settings.SettingsClient through what is fundamentally a
// trivial GET).
//
// Why a free standalone client instead of reusing settings.SettingsClient:
// keeping HTTPClient in pkg/whoami breaks the otherwise-natural cycle
// (profile imports whoami; settings imports profile.NewWhoamiCommand;
// settings.SettingsClient lives in settings → profile would have to
// import settings) and lets login.go fetch the role without pulling the
// entire settings package into the auth flow.
//
// Auth + 401/403 reformatting are handled by the upstream http.Client
// (factory.authTransport injects X-Authorization) and by formatBackendErr
// below — kept consistent with files/download.go's reformatHTTPError so
// users see one message for all "your token doesn't work" cases.
type HTTPClient struct {
	hc       *http.Client
	baseURL  string
	olaresID string
	// accessToken, when non-empty, is set on every outbound request as
	// `X-Authorization`. This is for the `profile login` / `profile import`
	// eager-fetch path, where we have a freshly-minted token but Factory's
	// authTransport may be tied to a stale resolved profile (or to none).
	// When empty, we trust the underlying http.Client's RoundTripper to
	// inject the header (factory.authTransport does this).
	accessToken string
}

// NewHTTPClient builds a whoami Doer pointed at <desktopURL>, reusing a
// caller-supplied http.Client. Use this when the caller has a Factory
// http.Client whose RoundTripper already injects X-Authorization for the
// active profile (e.g. `profile whoami` after a successful ResolveProfile).
// olaresID is included only for diagnostic messages.
func NewHTTPClient(hc *http.Client, desktopURL, olaresID string) *HTTPClient {
	return &HTTPClient{
		hc:       hc,
		baseURL:  strings.TrimRight(desktopURL, "/"),
		olaresID: olaresID,
	}
}

// NewHTTPClientWithToken is the eager-fetch variant: it builds a fresh
// http.Client (no Factory memoization, no shared timeout) and injects the
// supplied access token on every request. Used by `profile login` /
// `profile import` to call /api/backend/v1/user-info with the token they
// just persisted, avoiding the chicken-and-egg "Factory's resolved profile
// is still the previous one" hazard.
//
// insecureSkipVerify mirrors the profile's TLS knob — we honor whatever
// the user opted into for this profile rather than re-derive it.
func NewHTTPClientWithToken(desktopURL, olaresID, accessToken string, insecureSkipVerify bool) *HTTPClient {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if insecureSkipVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	hc := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
	return &HTTPClient{
		hc:          hc,
		baseURL:     strings.TrimRight(desktopURL, "/"),
		olaresID:    olaresID,
		accessToken: accessToken,
	}
}

// DoJSON satisfies Doer. body / out may be nil.
func (c *HTTPClient) DoJSON(ctx context.Context, method, path string, body, out interface{}) error {
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
	if c.accessToken != "" {
		req.Header.Set("X-Authorization", c.accessToken)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return formatBackendAuthErr(resp.StatusCode, respBody, c.olaresID)
	}
	if resp.StatusCode/100 != 2 {
		return formatBackendErr(method, url, resp.StatusCode, respBody)
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s %s response: %w (body=%s)", method, url, err, truncate(string(respBody), 200))
	}
	return nil
}

// formatBackendAuthErr mirrors files/download.go's 401/403 branch and
// market/client.go's reformatMarketAuthErr — turn the edge proxy's "you
// are not authenticated" into the same actionable CTA the rest of the CLI
// uses, so the user's troubleshooting story is consistent across verbs.
func formatBackendAuthErr(status int, respBody []byte, olaresID string) error {
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

// formatBackendErr handles non-401/403 non-2xx responses. user-service and
// BFL both speak loose JSON; we try the structured shapes and fall through
// to a body-truncated raw dump so the user always gets something to grep.
func formatBackendErr(method, url string, status int, body []byte) error {
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
