// frp.go: fetch the public Olares-tunnel (FRP) registry by olaresId.
//
// Mirrors the TS reference in
// TermiPass/packages/app/src/stores/wizard-step.ts (getFrpList) and the
// shared host map in TermiPass/packages/core/src/global.ts:
//
//   - olaresId ending with `olares.cn` → https://api.olares.cn/frp
//   - everything else                  → https://api.olares.com/frp
//
// We POST `/v2/servers` with `{"name": "<olaresId>"}` and decode the
// returned `OlaresTunneV2Interface[]` shape:
//
//   [{ region: string, name: { "en-US": ..., "zh-CN": ... }, machine: [{ host: string }] }]
//
// This endpoint is public (no auth header) and is the same one the
// activation wizard's "select tunnel" step calls before the user has
// finished binding to a per-user namespace.
package wizard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FrpEnvironment selects which public Olares API host to talk to.
type FrpEnvironment string

const (
	FrpEnvCN FrpEnvironment = "cn"
	FrpEnvEN FrpEnvironment = "en"
)

// frpListURLs mirrors GolbalHost.FRP_LIST_URL in
// TermiPass/packages/core/src/global.ts.
var frpListURLs = map[FrpEnvironment]string{
	FrpEnvEN: "https://api.olares.com/frp",
	FrpEnvCN: "https://api.olares.cn/frp",
}

// FrpServer is one entry of the public Olares-tunnel registry. The
// `Name` field is a {locale: label} map (e.g. `en-US`, `zh-CN`) — the
// caller picks which locale to render.
type FrpServer struct {
	Region  string            `json:"region"`
	Name    map[string]string `json:"name"`
	Machine []FrpMachine      `json:"machine"`
}

// FrpMachine is one of the (potentially many) reachable hosts for a
// region. The TS code uses `machine[0].host` as the default selection;
// we surface every host so callers can decide.
type FrpMachine struct {
	Host string `json:"host"`
}

// FrpListOptions tunes the FetchFrpList call. The zero value is valid
// and uses sane defaults.
type FrpListOptions struct {
	// Environment forces the API host. When empty, FetchFrpList falls
	// back to FrpEnvironmentForOlaresID(olaresID).
	Environment FrpEnvironment
	// HTTPClient overrides the default 10s-timeout client. Useful for
	// tests; production callers should leave this nil.
	HTTPClient *http.Client
	// Timeout overrides the default 10s timeout when HTTPClient is
	// nil. Ignored when HTTPClient is set.
	Timeout time.Duration
}

// FrpEnvironmentForOlaresID picks the right environment based on the
// olaresId suffix, matching TS userNameToEnvironment().
func FrpEnvironmentForOlaresID(olaresID string) FrpEnvironment {
	if strings.HasSuffix(strings.ToLower(strings.TrimSpace(olaresID)), "olares.cn") {
		return FrpEnvCN
	}
	return FrpEnvEN
}

// FrpListBaseURL returns the public Olares API base URL the FRP list
// call targets for `env`. Unknown envs fall back to the EN endpoint
// (same defensive default as the TS code).
func FrpListBaseURL(env FrpEnvironment) string {
	if u, ok := frpListURLs[env]; ok {
		return u
	}
	return frpListURLs[FrpEnvEN]
}

// FetchFrpList calls POST <FrpListBaseURL>/v2/servers and returns the
// decoded server list. The endpoint is unauthenticated; the only input
// is the olaresId, which the registry uses to scope the response.
func FetchFrpList(ctx context.Context, olaresID string, opts FrpListOptions) ([]FrpServer, error) {
	if strings.TrimSpace(olaresID) == "" {
		return nil, fmt.Errorf("olaresId is required")
	}

	env := opts.Environment
	if env == "" {
		env = FrpEnvironmentForOlaresID(olaresID)
	}
	endpoint := FrpListBaseURL(env) + "/v2/servers"

	client := opts.HTTPClient
	if client == nil {
		timeout := opts.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}

	body, err := json.Marshal(map[string]string{"name": olaresID})
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("POST %s: HTTP %d: %s", endpoint, resp.StatusCode, truncateFRPBody(raw))
	}

	var list []FrpServer
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("decode response: %w (body=%s)", err, truncateFRPBody(raw))
	}
	return list, nil
}

// LocalizedName returns the label for `locale`, falling back to en-US
// then any non-empty entry. Mirrors TS olaresTunnelsV2Options() in
// stores/settings/network.ts.
func (s FrpServer) LocalizedName(locale string) string {
	if s.Name == nil {
		return ""
	}
	if v, ok := s.Name[locale]; ok && v != "" {
		return v
	}
	if v, ok := s.Name["en-US"]; ok && v != "" {
		return v
	}
	for _, v := range s.Name {
		if v != "" {
			return v
		}
	}
	return ""
}

// FirstHost returns the first reachable host (matching TS
// `machine[0].host`), or "" when the entry has no machines.
func (s FrpServer) FirstHost() string {
	if len(s.Machine) == 0 {
		return ""
	}
	return s.Machine[0].Host
}

func truncateFRPBody(b []byte) string {
	const max = 256
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
}
