// Package appearance hosts `olares-cli settings appearance`. Mirrors the
// SPA's Settings -> Appearance page: locale, theme, widget preferences,
// wallpaper, and the desktop layout reset.
//
// Most of it is backed by user-service (wallpaper.controller.ts,
// widget.controller.ts, desktop-layout.controller.ts) on the /api/*
// prefix, which re-wraps its upstream with returnSucceed(...) so the CLI
// sees a uniform BFL envelope. Two calls step outside that:
//
//   - the theme write goes to /api/env/userenvs, because the theme lives
//     in the OLARES_USER_THEME UserEnv (see theme.go);
//   - the wallpaper image upload goes to the settings ingress, because
//     /images is not proxied on the desktop host (see imageClient below).
//
// Same per-area common.go pattern as settings/me / users / apps / vpn /
// network — each subpackage owns its own decoder so per-endpoint quirks
// stay local.
package appearance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/bflenvelope"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// appearanceMinOlaresVersion is the first Olares line carrying the three
// Appearance surfaces that 1.12.5 does not have: the widget preferences
// API (/api/widget), the desktop layout reset route, and the
// OLARES_USER_THEME user env the theme is stored in. The rest of the
// subtree (locale, wallpaper) works on 1.12.5, so the gate is per-verb
// rather than on the whole area.
const appearanceMinOlaresVersion = "1.12.6"

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

func parseFormat(s string) (Format, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "", string(FormatTable):
		return FormatTable, nil
	case string(FormatJSON):
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported --output %q (allowed: table, json)", s)
	}
}

func addOutputFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "output", "o", "table", "output format: table, json")
}

type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

type preparedClient struct {
	profile *credential.ResolvedProfile
	doer    Doer
	images  *imageClient
}

func prepare(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: settings appearance not wired with cmdutil.Factory")
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	uploadHC, err := f.HTTPClientWithoutTimeout(ctx)
	if err != nil {
		return nil, err
	}
	return &preparedClient{
		profile: rp,
		doer:    whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID),
		images: &imageClient{
			hc:       uploadHC,
			baseURL:  strings.TrimRight(rp.SettingsURL, "/"),
			olaresID: rp.OlaresID,
		},
	}, nil
}

func doGetEnvelope(ctx context.Context, d Doer, path string, out interface{}) error {
	return doMutateEnvelope(ctx, d, "GET", path, nil, out)
}

// requireAppearanceBackendVersion is the client-side version preflight for
// the Appearance verbs whose upstream landed in 1.12.6, so an older backend
// gets the canonical upgrade message instead of an opaque 404 (widget,
// layout) or a missing-CR message (theme). Thin wrapper supplying this
// area's gate copy to the shared helper, like `settings compute`'s
// requireComputeBackendVersion.
func requireAppearanceBackendVersion(ctx context.Context, f *cmdutil.Factory, verb, reason string) error {
	return preflight.RequireMinVersion(ctx, f, preflight.MinVersionGate{
		Verb:       verb,
		MinVersion: appearanceMinOlaresVersion,
		Reason:     reason,
		Fallback:   "upgrade the Olares system to >= " + appearanceMinOlaresVersion,
	})
}

// doMutateEnvelope is the POST/PUT/DELETE counterpart of doGetEnvelope:
// fire the request with an optional JSON body, decode the BFL envelope,
// and return an error if the upstream code is not 0/200. user-service's
// wallpaper.controller.ts re-wraps the BFL response with
// returnSucceed(response.data.data), so the same envelope contract holds
// for writes.
func doMutateEnvelope(ctx context.Context, d Doer, method, path string, body, out interface{}) error {
	var env bflenvelope.Envelope
	if err := d.DoJSON(ctx, method, path, body, &env); err != nil {
		return err
	}
	return bflenvelope.Data(method, path, env, out)
}

// imageClient posts images to tapr's images-uploader.
//
// It rides SettingsURL, not DesktopURL: /images is proxied only by the
// settings and profile-editor ingresses (apps/docker/system-frontend/
// nginx/settings.conf), so this cannot share the base URL every /api/*
// call in this package uses.
type imageClient struct {
	hc       *http.Client
	baseURL  string
	olaresID string
}

// upload streams filePath as the multipart `image` part and decodes the
// envelope the uploader replies with. The client has no overall timeout;
// the request context stays the cancellation boundary.
func (c *imageClient) upload(ctx context.Context, path, filePath string, fields map[string]string, out interface{}) error {
	body, contentType, err := multipartImage(filePath, fields)
	if err != nil {
		return err
	}
	target := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", target, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("POST %s: read response body: %w", target, err)
	}
	switch {
	case resp.StatusCode == http.StatusUnauthorized,
		resp.StatusCode == http.StatusForbidden,
		resp.StatusCode == 459:
		return credential.FormatHTTPAuthError(resp.StatusCode, respBody, c.olaresID)
	case resp.StatusCode/100 != 2:
		return fmt.Errorf("POST %s: HTTP %d: %s", target, resp.StatusCode, truncate(string(respBody), 500))
	}
	var env bflenvelope.Envelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return fmt.Errorf("POST %s: decode response: %w (body=%s)", target, err, truncate(string(respBody), 200))
	}
	return bflenvelope.Data(http.MethodPost, path, env, out)
}

// multipartImage streams the file rather than buffering it, with the text
// fields written first so the receiver sees them before the image bytes.
// Only opening the file fails here; a read that fails mid-upload travels
// down the pipe and fails the request.
func multipartImage(filePath string, fields map[string]string) (io.Reader, string, error) {
	fh, err := os.Open(filePath)
	if err != nil {
		// os.Open's error already names the operation and the path.
		return nil, "", err
	}
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		defer fh.Close()
		_ = pw.CloseWithError(writeImageParts(mw, fh, filePath, fields))
	}()
	return pr, mw.FormDataContentType(), nil
}

func writeImageParts(mw *multipart.Writer, file io.Reader, filePath string, fields map[string]string) error {
	for k, v := range fields {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if err := mw.WriteField(k, v); err != nil {
			return fmt.Errorf("write field %s: %w", k, err)
		}
	}
	part, err := mw.CreateFormFile("image", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("create the image part: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("finish the upload: %w", err)
	}
	return nil
}

func printJSON(w io.Writer, v interface{}) error {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func nonEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
