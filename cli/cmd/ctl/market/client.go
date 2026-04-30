package market

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// apiPrefix is the app-store v2 path the Market SPA also uses; see
// apps/packages/app/src/stores/market/center.ts (`appUrl`).
const apiPrefix = "/app-store/api/v2"

// APIResponse is the canonical envelope the app-store v2 backend wraps every
// response in (success/message/data). We keep it identical to the shape the
// SPA's axios layer parses so the CLI's diagnostics can use the same fields.
type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// MarketClient talks to the per-user app-store v2 API at
// `<MarketURL>/app-store/api/v2`. It is the moral counterpart of `files`'s
// `download.Client`: a thin HTTP wrapper that delegates auth to the caller's
// http.Client (Factory's refreshingTransport injects `X-Authorization` and
// transparently refreshes on 401/403) and otherwise just maps Go method
// calls to JSON requests.
//
// Two HTTP clients are stored:
//   - httpClient is the factory's standard 30s-timeout client; used for
//     short JSON requests.
//   - uploadClient is the factory's no-timeout client for multipart chart
//     uploads. Both share the same refreshingTransport instance under the
//     hood, so a refresh triggered through one is immediately visible to
//     the other.
type MarketClient struct {
	httpClient   *http.Client
	uploadClient *http.Client
	baseURL      string
	source       string

	// olaresID is captured for OperationResult.User (diagnostics /
	// scripting) and for reformatting 401/403 messages with the user's
	// own ID in the CTA.
	olaresID string
}

// NewMarketClient builds a MarketClient from factory-provided http.Clients
// (both already wired with refreshingTransport so X-Authorization injection
// + refresh-on-401 happen transparently) and a resolved profile. The base
// URL is `<rp.MarketURL>/app-store/api/v2`.
//
// hc is the standard timed client used for JSON requests; uploadHC is the
// no-timeout client used for multipart chart uploads. Pass the same client
// for both if streaming uploads aren't expected.
func NewMarketClient(hc, uploadHC *http.Client, rp *credential.ResolvedProfile, source string) *MarketClient {
	base := strings.TrimRight(rp.MarketURL, "/") + apiPrefix
	return &MarketClient{
		httpClient:   hc,
		uploadClient: uploadHC,
		baseURL:      base,
		source:       source,
		olaresID:     rp.OlaresID,
	}
}

func (c *MarketClient) doRequest(ctx context.Context, method, path string, body interface{}) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.executeRequest(c.httpClient, req)
}

func (c *MarketClient) doMultipart(ctx context.Context, path, filename string, data io.Reader, source string) (*APIResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("chart", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, data); err != nil {
		return nil, fmt.Errorf("failed to copy chart data: %w", err)
	}
	if err := writer.WriteField("source", source); err != nil {
		return nil, fmt.Errorf("failed to write source field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	// X-Authorization is injected by the factory's refreshingTransport;
	// we just pick the no-timeout client so large pushes aren't killed
	// by the default 30s deadline. The body is a *bytes.Buffer so
	// http.NewRequest sets GetBody automatically — refresh+retry on 401
	// works here.
	return c.executeRequest(c.uploadClient, req)
}

func (c *MarketClient) executeRequest(hc *http.Client, req *http.Request) (*APIResponse, error) {
	resp, err := hc.Do(req)
	if err != nil {
		// The factory's refreshingTransport returns a typed
		// credential error when /api/refresh itself fails (the grant
		// is dead, or no token is stored at all). http.Client wraps
		// it inside *url.Error, but errors.As walks the Unwrap chain
		// — surface the typed error directly so the caller sees the
		// canonical "run profile login" CTA instead of
		// `request failed: Get "https://...": refresh token for ...`.
		var inv *credential.ErrTokenInvalidated
		if errors.As(err, &inv) {
			return nil, inv
		}
		var nli *credential.ErrNotLoggedIn
		if errors.As(err, &nli) {
			return nil, nli
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 401/403 reaches us only when the factory's refreshingTransport
	// already attempted a refresh+retry and STILL got rejected (the
	// server is consistently saying no — usually a server-side state
	// drift the user can't recover from). Reformat with the standard
	// `profile login` CTA so users hit the same wording they get from
	// `files ls`/`files cat`. The body may not be JSON (the edge proxy
	// can short-circuit to a plaintext page), so the JSON parse below
	// is best-effort.
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, reformatMarketAuthErr(resp.StatusCode, respBody, c.olaresID)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !apiResp.Success {
		message := strings.TrimSpace(apiResp.Message)
		if message == "" {
			message = strings.TrimSpace(string(respBody))
		}
		if message == "" {
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return &apiResp, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, message)
	}

	return &apiResp, nil
}

// reformatMarketAuthErr mirrors reformatHTTPErr in cmd/ctl/files/download.go:
// turn 401/403 into the same `olares-cli profile login --olares-id <id>` CTA
// users see from the files verbs, so the troubleshooting story is consistent.
func reformatMarketAuthErr(status int, respBody []byte, olaresID string) error {
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

func (c *MarketClient) GetMarketData(ctx context.Context) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/market/data", nil)
}

func (c *MarketClient) GetMarketState(ctx context.Context) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/market/state", nil)
}

func (c *MarketClient) GetAppsInfo(ctx context.Context, apps []AppQueryInfo) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/apps", map[string]interface{}{
		"apps": apps,
	})
}

func (c *MarketClient) UploadChart(ctx context.Context, filePath, source string) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	return c.doMultipart(ctx, "/apps/upload", file.Name(), file, source)
}

func (c *MarketClient) UploadChartFromReader(ctx context.Context, filename string, data io.Reader, source string) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doMultipart(ctx, "/apps/upload", filename, data, source)
}

func (c *MarketClient) DeleteLocalApp(ctx context.Context, appName, appVersion, sourceID string) (*APIResponse, error) {
	if sourceID == "" {
		sourceID = c.source
	}
	return c.doRequest(ctx, http.MethodDelete, "/local-apps/delete", map[string]string{
		"app_name":    appName,
		"app_version": appVersion,
		"source":      sourceID,
	})
}

func (c *MarketClient) InstallApp(ctx context.Context, appName, version, source string, envs []AppEnvVar) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPost, "/apps/"+appName+"/install", InstallRequest{
		Source:  source,
		AppName: appName,
		Version: version,
		Sync:    true,
		Envs:    envs,
	})
}

func (c *MarketClient) CloneApp(ctx context.Context, appName, source, title string, envs []AppEnvVar, entrances []AppEntrance) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPost, "/apps/"+appName+"/clone", CloneRequest{
		Source:    source,
		AppName:   appName,
		Title:     title,
		Sync:      true,
		Envs:      envs,
		Entrances: entrances,
	})
}

func (c *MarketClient) UninstallApp(ctx context.Context, appName string, all, deleteData bool) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, "/apps/"+appName, UninstallRequest{
		Sync:       true,
		All:        all,
		DeleteData: deleteData,
	})
}

func (c *MarketClient) UpgradeApp(ctx context.Context, appName, version, source string, envs []AppEnvVar) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPut, "/apps/"+appName+"/upgrade", InstallRequest{
		Source:  source,
		AppName: appName,
		Version: version,
		Sync:    true,
		Envs:    envs,
	})
}

func (c *MarketClient) CancelOperation(ctx context.Context, appName string) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, "/apps/"+appName+"/install", map[string]interface{}{
		"sync": true,
	})
}

func (c *MarketClient) ResumeApp(ctx context.Context, appName string) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/apps/resume", map[string]string{
		"appName": appName,
	})
}

func (c *MarketClient) StopApp(ctx context.Context, appName string, all bool) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/apps/stop", map[string]interface{}{
		"appName": appName,
		"all":     all,
	})
}
