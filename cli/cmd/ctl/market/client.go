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
	"net/url"
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

// ChartPackage is an open stream of a chart .tgz served by
// GET /apps/{name}/package. Body is the caller's to close.
type ChartPackage struct {
	Body io.ReadCloser

	// Filename is the name the backend suggested through
	// Content-Disposition (`<chart>-<version>.tgz`), empty when the header
	// is absent or unparseable.
	Filename string

	// Size is the announced Content-Length, or -1 when the response was
	// chunked.
	Size int64
}

// downloadStream issues a GET whose success body is bytes rather than the
// app-store JSON envelope, so it cannot go through executeRequest: that
// helper reads the whole body into memory and insists on parsing it as
// APIResponse. Failures still arrive as the envelope, and are decoded here
// with the same wording so a 404 reads the same whichever verb hit it.
//
// The caller closes the returned body.
func (c *MarketClient) downloadStream(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// The error path answers JSON, the success path octet-stream.
	req.Header.Set("Accept", "application/octet-stream, application/json")

	// The no-timeout client, for the same reason uploads use it: a chart is
	// large enough that the standard 30s deadline can cut a healthy transfer
	// off partway through.
	resp, err := c.uploadClient.Do(req)
	if err != nil {
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

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, errorBodyLimit))
		resp.Body.Close()
		return nil, reformatMarketAuthErr(resp.StatusCode, body, c.olaresID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, errorBodyLimit))
		resp.Body.Close()
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, envelopeMessage(body, resp.StatusCode))
	}
	// A 200 carrying the JSON envelope means the backend answered a
	// different question than the one asked. Writing that to a .tgz would
	// produce a file that only fails later, at `helm install` time.
	if ct := resp.Header.Get("Content-Type"); strings.Contains(strings.ToLower(ct), "application/json") {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, errorBodyLimit))
		resp.Body.Close()
		return nil, fmt.Errorf("expected chart bytes but got a JSON response: %s", envelopeMessage(body, http.StatusOK))
	}
	return resp, nil
}

// errorBodyLimit caps how much of a failure body is read for the message.
// The edge proxy can answer an HTML page, and a whole one in a terminal
// buries the line that matters.
const errorBodyLimit = 8 << 10

// envelopeMessage extracts the human-readable part of an app-store error
// body: the envelope's `message` when it parses, the trimmed body when it
// doesn't, and the bare status when there is nothing at all.
func envelopeMessage(body []byte, status int) string {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err == nil {
		if msg := strings.TrimSpace(apiResp.Message); msg != "" {
			return msg
		}
	}
	if msg := strings.TrimSpace(string(body)); msg != "" {
		return msg
	}
	return fmt.Sprintf("HTTP %d", status)
}

// DownloadChart opens GET /apps/{name}/package for reading. An empty version
// lets the backend pick the current one; an empty source falls back to the
// client's default, and then to the backend's (`upload`).
//
// The returned ChartPackage owns an open connection — close its Body.
func (c *MarketClient) DownloadChart(ctx context.Context, appName, source, version string) (*ChartPackage, error) {
	if source == "" {
		source = c.source
	}
	query := url.Values{}
	if source != "" {
		query.Set("source", source)
	}
	if version != "" {
		query.Set("version", version)
	}
	resp, err := c.downloadStream(ctx, "/apps/"+url.PathEscape(appName)+"/package", query)
	if err != nil {
		return nil, err
	}
	return &ChartPackage{
		Body:     resp.Body,
		Filename: parseContentDispositionFilename(resp.Header.Get("Content-Disposition")),
		Size:     resp.ContentLength,
	}, nil
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

// InstallApp issues POST /apps/{name}/install. selectedGpuType pins the
// compute mode on Olares 1.12.6+ (InstallRequest.SelectedGpuType, omitempty):
// callers pass "" on 1.12.5 so the wire stays byte-identical to before the
// field existed, and the 1.12.6 path passes either the --compute-mode value or
// the mode resolved from a computeModeSelect 422 retry.
func (c *MarketClient) InstallApp(ctx context.Context, appName, version, source, selectedGpuType string, envs []AppEnvVar) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPost, "/apps/"+appName+"/install", InstallRequest{
		Source:          source,
		AppName:         appName,
		Version:         version,
		Sync:            true,
		Envs:            envs,
		SelectedGpuType: selectedGpuType,
	})
}

// CloneApp issues POST /apps/{name}/clone. selectedGpuType pins the compute
// mode on Olares 1.12.6+ (CloneRequest.SelectedGpuType, omitempty), mirroring
// InstallApp: callers pass "" on 1.12.5 so the wire stays byte-identical, and
// the 1.12.6 path passes either the --compute-mode value or the mode resolved
// from a computeModeSelect 422 retry.
func (c *MarketClient) CloneApp(ctx context.Context, appName, source, title, selectedGpuType string, envs []AppEnvVar, entrances []AppEntrance, templateClone bool) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPost, "/apps/"+appName+"/clone", CloneRequest{
		Source:          source,
		AppName:         appName,
		Title:           title,
		Sync:            true,
		Envs:            envs,
		Entrances:       entrances,
		SelectedGpuType: selectedGpuType,
		TemplateClone:   templateClone,
	})
}

// UpgradeApp issues PUT /apps/{name}/upgrade. The payload intentionally
// omits env vars: the Market SPA's upgradeApp() never sends them either
// (existing values are preserved server-side from the prior install).
// Use `olares-cli market env` to update env values out-of-band when
// upgrading isn't the right tool.
func (c *MarketClient) UpgradeApp(ctx context.Context, appName, version, source string) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPut, "/apps/"+appName+"/upgrade", UpgradeRequest{
		Source:  source,
		AppName: appName,
		Version: version,
		Sync:    true,
	})
}

// StopApp / ResumeApp / UninstallApp / CancelOperation used to build their
// request bodies here. Their wire format diverges across Olares versions
// (TermiPass PR #1162: stop/resume/uninstall all moved to {app_name, source,
// ...}; cancel moved to {app_name, source, version}), so body shaping now lives
// in per-command, per-version builder sub-packages
// (market/{stop,resume,uninstall,cancel}/{1_12_5,1_12_6}). Those builders
// return (method, path, body); the runners pick the version with
// cmdutil.Factory.OlaresBackendAtLeast and send through doRequest, so transport
// (auth injection, refresh-on-401, envelope validation) is never re-implemented.
// UninstallRequest is retained in types.go as the 1.12.5 body shape reference.
