// Package api is a thin HTTP client for the local olaresd daemon.
//
// olaresd listens on TCP port 18088 by default and exposes JSON
// endpoints under paths like /system/status, /system/ifs, etc. All
// endpoints are loopback-only (RequireLocal middleware on the daemon
// side), so the client expects to talk to 127.0.0.1.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/daemon/state"
)

// DefaultEndpoint is the loopback URL where olaresd listens by
// default. Callers normally do not need to override this.
const DefaultEndpoint = "http://127.0.0.1:18088"

// DefaultTimeout is the per-request timeout used when the caller
// does not supply one.
const DefaultTimeout = 5 * time.Second

// Client is a small HTTP client for the olaresd daemon. The zero
// value uses DefaultEndpoint and DefaultTimeout; callers may
// override either field before issuing requests.
type Client struct {
	// Endpoint is the base URL of olaresd, e.g.
	// "http://127.0.0.1:18088". Trailing slashes are stripped.
	Endpoint string

	// Timeout bounds the total time spent on a single HTTP request,
	// including dialing, TLS, and reading the response body.
	Timeout time.Duration

	// HTTPClient lets tests substitute a custom transport. When
	// nil, the client constructs one on demand using Timeout.
	HTTPClient *http.Client
}

// NewClient returns a Client that hits the given endpoint with the
// supplied timeout. Pass an empty endpoint or zero timeout to fall
// back to DefaultEndpoint / DefaultTimeout.
func NewClient(endpoint string, timeout time.Duration) *Client {
	return &Client{Endpoint: endpoint, Timeout: timeout}
}

func (c *Client) endpoint() string {
	ep := c.Endpoint
	if ep == "" {
		ep = DefaultEndpoint
	}
	return strings.TrimRight(ep, "/")
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	t := c.Timeout
	if t <= 0 {
		t = DefaultTimeout
	}
	return &http.Client{Timeout: t}
}

// GetSystemStatus calls GET /system/status and returns:
//   - the parsed State struct for programmatic consumption;
//   - the raw `data` JSON bytes so callers that want to forward
//     the response (for example, the CLI's --json mode) can do so
//     without a re-marshal round trip.
//
// The error message always contains the endpoint URL and the HTTP
// status (when applicable) so users can tell the difference between
// "olaresd is down" and "olaresd returned an error".
func (c *Client) GetSystemStatus(ctx context.Context) (*state.State, []byte, error) {
	u, err := url.Parse(c.endpoint())
	if err != nil {
		return nil, nil, fmt.Errorf("invalid olaresd endpoint %q: %w", c.endpoint(), err)
	}
	u.Path = "/system/status"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("build request to %s: %w", u.String(), err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request olaresd at %s (is olaresd running?): %w", u.String(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response from %s: %w", u.String(), err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("olaresd %s returned HTTP %d: %s", u.String(), resp.StatusCode, truncate(string(body), 200))
	}

	// olaresd wraps payloads as {code, message, data}. We need the
	// raw `data` slice both to populate State and to expose it
	// verbatim to --json callers.
	var raw struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, nil, fmt.Errorf("decode response envelope from %s: %w", u.String(), err)
	}
	if raw.Code != http.StatusOK {
		return nil, nil, fmt.Errorf("olaresd returned code=%d message=%q", raw.Code, raw.Message)
	}

	var s state.State
	if err := json.Unmarshal(raw.Data, &s); err != nil {
		return nil, nil, fmt.Errorf("decode state payload from %s: %w", u.String(), err)
	}

	return &s, []byte(raw.Data), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
