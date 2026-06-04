package olaresclient

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Masterminds/semver/v3"
)

var version_1_12_5 = semver.MustParse("1.12.5")

// clientV1_12_5 is the baseline implementation: the market app-lifecycle wire
// format as it existed on Olares 1.12.5 (pre TermiPass PR #1162). It provides
// the full MarketAppOps surface so newer clients only need to override the
// methods that actually changed.
type clientV1_12_5 struct {
	baseClient
}

func newClientV1_12_5(backendVersion *semver.Version) (OlaresClient, error) {
	return &clientV1_12_5{baseClient{version: backendVersion}}, nil
}

// StopApp on 1.12.5: POST /apps/stop with body {appName, all}. The source
// argument is ignored on this line — 1.12.5 does not send it.
func (c *clientV1_12_5) StopApp(ctx context.Context, d Doer, appName, source string, all bool) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodPost, "/apps/stop", map[string]any{
		"appName": appName,
		"all":     all,
	})
}

// ResumeApp on 1.12.5: POST /apps/resume with body {appName}. source ignored.
func (c *clientV1_12_5) ResumeApp(ctx context.Context, d Doer, appName, source string) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodPost, "/apps/resume", map[string]any{
		"appName": appName,
	})
}

// UninstallApp on 1.12.5: DELETE /apps/{name}; the app name rides the path and
// the body carries only {sync, all, deleteData}. The source / version
// arguments are ignored on this line — 1.12.5 does not send them.
func (c *clientV1_12_5) UninstallApp(ctx context.Context, d Doer, appName, source, version string, all, deleteData bool) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodDelete, "/apps/"+appName, map[string]any{
		"sync":       true,
		"all":        all,
		"deleteData": deleteData,
	})
}

// --- ComputeOps (GPU / compute acceleration) ---

// ListAccelerators on 1.12.5: GET /api/gpu/list (the legacy HAMI-backed GPU
// list; the compute-resources model did not exist yet).
func (c *clientV1_12_5) ListAccelerators(ctx context.Context, d Doer) (json.RawMessage, error) {
	return d.Do(ctx, http.MethodGet, "/api/gpu/list", nil)
}

// GetAppBindings is a 1.12.6+ capability; 1.12.5 has no compute-resources API.
func (c *clientV1_12_5) GetAppBindings(_ context.Context, _ Doer, _ string) (json.RawMessage, error) {
	return nil, c.computeUnsupported("settings gpu bindings")
}

// ReleaseAppBindings is a 1.12.6+ capability.
func (c *clientV1_12_5) ReleaseAppBindings(_ context.Context, _ Doer, _ string) (json.RawMessage, error) {
	return nil, c.computeUnsupported("settings gpu unbind")
}

// SwitchSupportType is a 1.12.6+ capability.
func (c *clientV1_12_5) SwitchSupportType(_ context.Context, _ Doer, _, _, _ string) (json.RawMessage, error) {
	return nil, c.computeUnsupported("settings gpu support-type set")
}

// --- OverlayOps (overlay gateway) — entirely 1.12.6+ ---

func (c *clientV1_12_5) OverlayGatewayStatus(_ context.Context, _ Doer, _ string) (json.RawMessage, error) {
	return nil, c.overlayUnsupported("settings network overlay status")
}

func (c *clientV1_12_5) EnableOverlayGateway(_ context.Context, _ Doer) (json.RawMessage, error) {
	return nil, c.overlayUnsupported("settings network overlay enable")
}

func (c *clientV1_12_5) DisableOverlayGateway(_ context.Context, _ Doer) (json.RawMessage, error) {
	return nil, c.overlayUnsupported("settings network overlay disable")
}

// computeUnsupported / overlayUnsupported build the capability-gate error for
// operations that only exist on Olares 1.12.6+. Current reflects the real
// detected backend version (from baseClient), so the message is accurate even
// for the default/fallback client which embeds this type.
func (c *clientV1_12_5) computeUnsupported(feature string) error {
	return &ErrUnsupportedVersion{Feature: feature, MinVersion: version_1_12_6, Current: c.Version()}
}

func (c *clientV1_12_5) overlayUnsupported(feature string) error {
	return &ErrUnsupportedVersion{Feature: feature, MinVersion: version_1_12_6, Current: c.Version()}
}

func init() {
	registerClientFactory(version_1_12_5, newClientV1_12_5)
}
