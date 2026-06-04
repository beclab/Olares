package olaresclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/Masterminds/semver/v3"
)

var version_1_12_6 = semver.MustParse("1.12.6")

// clientV1_12_6 embeds clientV1_12_5 so any capability NOT overridden here
// transparently delegates to the 1.12.5 implementation — the "1.12.6 didn't
// change this operation, so reuse 1.12.5" method-level fallback. Only the
// market operations whose wire format changed in TermiPass PR #1162 are
// overridden below.
type clientV1_12_6 struct {
	clientV1_12_5
}

func newClientV1_12_6(backendVersion *semver.Version) (OlaresClient, error) {
	return &clientV1_12_6{clientV1_12_5{baseClient{version: backendVersion}}}, nil
}

// StopApp on 1.12.6: body moved to {app_name, source, all} (TermiPass
// PR #1162 — StopRequest extends BaseOperationRequest{app_name, source}).
func (c *clientV1_12_6) StopApp(ctx context.Context, d Doer, appName, source string, all bool) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodPost, "/apps/stop", map[string]any{
		"app_name": appName,
		"source":   source,
		"all":      all,
	})
}

// ResumeApp on 1.12.6: body moved to {app_name, source} (TermiPass PR #1162 —
// Resume2RestartRequest extends BaseOperationRequest; computeBinding is
// optional and not exposed by the CLI).
func (c *clientV1_12_6) ResumeApp(ctx context.Context, d Doer, appName, source string) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodPost, "/apps/resume", map[string]any{
		"app_name": appName,
		"source":   source,
	})
}

// UninstallApp on 1.12.6: the request body now carries app_name + source
// (UninstallRequest extends BaseOperationRequest) alongside sync / all /
// deleteData (DELETE /apps/{name}). version is included only when the caller
// supplies one.
func (c *clientV1_12_6) UninstallApp(ctx context.Context, d Doer, appName, source, version string, all, deleteData bool) (json.RawMessage, error) {
	body := map[string]any{
		"app_name":   appName,
		"source":     source,
		"sync":       true,
		"all":        all,
		"deleteData": deleteData,
	}
	if version != "" {
		body["version"] = version
	}
	return d.Do(ctx, http.MethodDelete, "/apps/"+appName, body)
}

// --- ComputeOps (compute-resources model, replaces the 1.12.5 /api/gpu/*) ---
//
// clientV1_12_6 embeds clientV1_12_5 only to reuse the unchanged MarketAppOps
// methods. ComputeOps is a COMPLETE replacement (different endpoints + data
// model), so every method is overridden here — none of the 1.12.5 /api/gpu/*
// behavior must leak through.

// ListAccelerators on 1.12.6: GET /api/compute-resources (nodes → devices →
// bindings), replacing the legacy /api/gpu/list.
func (c *clientV1_12_6) ListAccelerators(ctx context.Context, d Doer) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodGet, "/api/compute-resources", nil)
}

// GetAppBindings on 1.12.6: GET /api/apps/{app}/compute-resources/bindings.
func (c *clientV1_12_6) GetAppBindings(ctx context.Context, d Doer, appName string) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodGet, "/api/apps/"+url.PathEscape(appName)+"/compute-resources/bindings", nil)
}

// ReleaseAppBindings on 1.12.6: DELETE /api/apps/{app}/compute-resources/bindings.
// The backend implements this as a suspend to free the GPU binding.
func (c *clientV1_12_6) ReleaseAppBindings(ctx context.Context, d Doer, appName string) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodDelete, "/api/apps/"+url.PathEscape(appName)+"/compute-resources/bindings", nil)
}

// SwitchSupportType on 1.12.6: PUT
// /api/compute-resources/nodes/{node}/devices/{device}/support-type with body
// {supportType}. The result carries a status discriminant
// (switched/unchanged/bound-apps-stop-blocked) the caller inspects.
func (c *clientV1_12_6) SwitchSupportType(ctx context.Context, d Doer, node, deviceID, supportType string) (json.RawMessage, error) {
	path := "/api/compute-resources/nodes/" + url.PathEscape(node) + "/devices/" + url.PathEscape(deviceID) + "/support-type"
	return d.Do(ctx, http.MethodPut, path, map[string]any{"supportType": supportType})
}

// --- OverlayOps (overlay gateway, new in 1.12.6) ---

// OverlayGatewayStatus on 1.12.6: GET /api/system/overlay-gateway-status/{user}.
func (c *clientV1_12_6) OverlayGatewayStatus(ctx context.Context, d Doer, user string) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodGet, "/api/system/overlay-gateway-status/"+url.PathEscape(user), nil)
}

// EnableOverlayGateway on 1.12.6: POST /api/command/enable-overlay-gateway (no body).
func (c *clientV1_12_6) EnableOverlayGateway(ctx context.Context, d Doer) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodPost, "/api/command/enable-overlay-gateway", nil)
}

// DisableOverlayGateway on 1.12.6: POST /api/command/disable-overlay-gateway (no body).
func (c *clientV1_12_6) DisableOverlayGateway(ctx context.Context, d Doer) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodPost, "/api/command/disable-overlay-gateway", nil)
}

func init() {
	registerClientFactory(version_1_12_6, newClientV1_12_6)
}
