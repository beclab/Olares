package clusterctx

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

// HTTPClient is a Doer that talks to the per-user ControlHub BFF
// directly. It exists for the same reason pkg/whoami.HTTPClient does:
// the auth flow (`profile login` / `profile import`) wants to fetch
// `/capi/app/detail` with a freshly-minted token but can't import
// cli/cmd/ctl/cluster (which would in turn need profile types) without
// inviting an import cycle.
//
// For the regular `cluster context` command, callers reach into
// pkg/clusterclient.Client instead — that one shares the Factory's
// refreshingTransport so X-Authorization auto-rotation is transparent.
// HTTPClient here is purely the standalone variant.
//
// Auth + 401/403 reformatting are handled by the upstream http.Client
// (factory.refreshingTransport when the caller used NewHTTPClient with
// a Factory client) and by formatBackendErr below. The wording matches
// pkg/whoami.HTTPClient verbatim so users see one CTA across all
// per-user verbs.
type HTTPClient struct {
	hc       *http.Client
	baseURL  string
	olaresID string
}

// NewHTTPClient builds a clusterctx Doer pointed at <controlHubURL>,
// reusing a caller-supplied http.Client. Use this when the caller has a
// Factory http.Client whose RoundTripper already injects
// X-Authorization (and auto-rotates expired tokens) for the active
// profile. olaresID is included only for diagnostic messages.
func NewHTTPClient(hc *http.Client, controlHubURL, olaresID string) *HTTPClient {
	return &HTTPClient{
		hc:       hc,
		baseURL:  strings.TrimRight(controlHubURL, "/"),
		olaresID: olaresID,
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

	resp, err := c.hc.Do(req)
	if err != nil {
		// Surface typed credential errors directly so the caller sees
		// the canonical "run profile login" CTA. Mirrors
		// cli/pkg/whoami/client.go and cli/pkg/clusterclient/client.go.
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

// formatBackendAuthErr keeps the 401/403 CTA word-for-word identical to
// pkg/whoami.formatBackendAuthErr so users see the same message
// regardless of which per-user origin rejected them.
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

// formatBackendErr handles non-401/403 non-2xx responses. ControlHub
// proxies several upstream services (kube-apiserver, KubeSphere, /capi/*
// aggregator), each with its own error shape; we try the structured
// shapes and fall back to a body-truncated raw dump.
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
