// Package download implements the per-file and recursive directory
// download path for `olares-cli files download` and `files cat`. It
// talks to the same per-user files-backend endpoints the LarePass web
// app uses:
//
//   - GET /api/resources/<fileType>/<extend><encPath>     → metadata / listing
//     (`Stat` returns the envelope shape used by the web app's
//     `getFileInfo` / `formatRequestUrl` helpers; see
//     [cli/cmd/ctl/files/ls.go] for the precedent).
//   - GET /api/raw/<fileType>/<extend><encPath>           → raw bytes
//     The file-server supports `Range: bytes=N-` (raw_service.go's
//     parseRangeHeader), so single-file resume is server-driven and we
//     don't need a sidecar progress file.
//
// Same X-Authorization injection convention as the upload package and
// the rest of the CLI (see pkg/cmdutil/factory.go's authTransport for
// the rationale): Olares' edge stack only forwards X-Authorization to
// per-user services, so the standard `Authorization: Bearer ...` would
// silently drop the credential on the way through.
package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beclab/Olares/cli/pkg/files/encodepath"
)

// Client is the per-FilesURL handle used by Stat / List / DownloadFile
// and by the cobra command. It is cheap to construct; reuse one per
// `files download` / `files cat` invocation.
//
// AccessToken is sent as `X-Authorization` (not `Authorization: Bearer`),
// because Olares' edge stack only forwards the X-Authorization header to
// per-user services. See pkg/cmdutil/factory.go for the full rationale.
type Client struct {
	HTTPClient  *http.Client
	BaseURL     string // FilesURL, e.g. https://files.alice.olares.com
	AccessToken string
}

// HTTPError carries the status + truncated body of a non-2xx response so
// callers can branch on the status code without stringly-typed error
// parsing (the same shape as upload.HTTPError; we keep an in-package
// type so download has no leaky abstractions in its public surface).
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

// resourcesURL returns `<BaseURL>/api/resources/<encPlainPath>`. The
// caller's trailing `/` (if any) is preserved — the backend uses it as
// a "this is a directory" hint, see files/pkg/models/file_param.go's
// FileParam.convert (it splits on '/' and rejects len < 3 for resource
// listings). plainPath looks like `drive/Home/Documents` or
// `drive/Home/Documents/`.
func (c *Client) resourcesURL(plainPath string) string {
	return c.BaseURL + "/api/resources/" + encodepath.EncodeURL(plainPath)
}

// rawURL returns `<BaseURL>/api/raw/<encPlainPath>`. Mirrors the web
// app's `driveCommonUrl('raw', filePath)` (data.ts in v2/drive). The
// raw endpoint refuses non-file paths with a 400, so callers should
// Stat first when the user-supplied path could be either.
func (c *Client) rawURL(plainPath string) string {
	return c.BaseURL + "/api/raw/" + encodepath.EncodeURL(plainPath)
}

// do performs a single HTTP request with the configured access token
// injected as `X-Authorization`, and returns the response body on 2xx.
// Non-2xx responses surface as *HTTPError so callers can branch on
// status (e.g. 404 from Stat is meaningful: "not found" vs. "auth
// problem").
//
// `body` may be nil. Extra headers (Range, Accept) ride on `headers`.
// We deliberately do NOT stream the body here — Stat / List responses
// are small JSON envelopes and the caller wants them whole. The
// downloader path bypasses do() and streams resp.Body directly so we
// never buffer a multi-GB file in memory.
func (c *Client) do(
	ctx context.Context,
	method, endpoint string,
	body io.Reader,
	headers http.Header,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if c.AccessToken != "" {
		req.Header.Set("X-Authorization", c.AccessToken)
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, &HTTPError{
			Status: resp.StatusCode,
			Body:   string(respBody),
			URL:    endpoint,
			Method: method,
		}
	}
	return respBody, nil
}

// trimResourcesPlainPath drops a trailing '/' from a resource path so
// Stat probes hit the "metadata for X" form rather than the "listing
// of directory X/" form. Callers pre-validate that the path is
// non-empty (the cobra cmd does this via ParseFrontendPath).
func trimResourcesPlainPath(p string) string {
	return strings.TrimRight(p, "/")
}
