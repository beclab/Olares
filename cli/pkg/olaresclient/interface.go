// Package olaresclient is the runtime version-compatibility layer for the
// remote / API commands of olares-cli. A single CLI binary may talk to Olares
// backends of different versions (e.g. 1.12.5 and 1.12.6); this package lets a
// command dispatch to a version-specific implementation chosen at runtime from
// the detected backend version, while version-agnostic transport (auth,
// envelope unwrapping, retries) stays in the calling package.
//
// Design mirrors pkg/upgrade's proven pattern: every version implementation
// lives in this single package (one file each) and self-registers via init().
// Keeping them in one package — rather than vX_Y_Z subpackages — is deliberate:
// subpackages would import this package for the registry while this package
// would have to blank-import them to trigger their init(), which is an import
// cycle. The upgrade package sidesteps the same problem the same way.
//
// Selection is by the backend's CORE version (Major.Minor.Patch, prerelease
// stripped) using a floor match: the registered implementation with the
// greatest core version that is <= the backend's core version wins. So:
//
//   - 1.12.6 / 1.12.6-20260603 (daily) / 1.12.6-alpha1 (prerelease) → 1.12.6
//   - 1.12.7-20260524 (newer than any known impl)                   → 1.12.6
//   - 1.12.5*                                                       → 1.12.5
//   - below the lowest registered impl / unknown                    → default
//
// Method-level fallback ("1.12.6 didn't change this op → use 1.12.5") is
// achieved with plain Go embedding: clientV1_12_6 embeds clientV1_12_5 and
// only overrides the methods whose wire format actually changed.
package olaresclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Doer executes a single JSON request against the per-user Olares backend and
// returns the raw `data` payload of the response envelope. The transport
// (token injection, refresh-on-401, envelope unwrapping, error reformatting)
// is owned by the caller — e.g. market.MarketClient implements this over its
// existing doRequest — so version implementations here only shape the request
// line (method / path / body) and never re-implement transport.
type Doer interface {
	Do(ctx context.Context, method, path string, body any) (json.RawMessage, error)
}

// VersionAware is implemented by every version client. Version reports the
// FULL backend version the client was created for (including any prerelease),
// so an implementation can branch internally on daily / alpha builds even
// though selection collapses to the core patch version.
type VersionAware interface {
	Version() *semver.Version
}

// MarketAppOps groups the market app-lifecycle operations whose wire format
// diverges across Olares versions. Per TermiPass PR #1162 the stop / resume /
// uninstall request bodies all moved from `{appName}` (and the old
// uninstall `{sync, all, deleteData}`) to a common `{app_name, source, ...}`
// shape — these are the representative version-specific operations.
//
// source is the market source the installed app belongs to (resolved by the
// caller from /market/state). 1.12.5 ignores it; 1.12.6 requires it.
type MarketAppOps interface {
	// StopApp suspends a running app (POST /apps/stop).
	StopApp(ctx context.Context, d Doer, appName, source string, all bool) (json.RawMessage, error)
	// ResumeApp resumes a suspended app (POST /apps/resume).
	ResumeApp(ctx context.Context, d Doer, appName, source string) (json.RawMessage, error)
	// UninstallApp uninstalls an app (DELETE /apps/{name}). version may be
	// empty for callers / versions that don't carry it.
	UninstallApp(ctx context.Context, d Doer, appName, source, version string, all, deleteData bool) (json.RawMessage, error)
}

// ComputeOps groups the GPU / compute-acceleration operations. This is the
// "complete replacement" version axis: Olares 1.12.5 exposed a GPU
// assign/mode model under /api/gpu/*, while 1.12.6 replaced it wholesale with
// a node/device/supportType/binding model under /api/compute-resources*
// (the old GPUController was deleted). The two share no wire format, so the
// version clients implement this interface independently (no embedding reuse).
//
// ListAccelerators exists on both lines (mapped to each version's list
// endpoint); the binding/support-type operations are 1.12.6+ only and return
// *ErrUnsupportedVersion on older lines. Every method returns the unwrapped
// `data` payload of the {code,message,data} envelope (the Doer owns transport
// + envelope unwrapping).
type ComputeOps interface {
	// ListAccelerators lists GPUs / compute resources. 1.12.5: GET
	// /api/gpu/list; 1.12.6: GET /api/compute-resources.
	ListAccelerators(ctx context.Context, d Doer) (json.RawMessage, error)
	// GetAppBindings returns an app's compute bindings (1.12.6+):
	// GET /api/apps/{app}/compute-resources/bindings.
	GetAppBindings(ctx context.Context, d Doer, appName string) (json.RawMessage, error)
	// ReleaseAppBindings releases an app's compute bindings (1.12.6+),
	// which suspends the app: DELETE /api/apps/{app}/compute-resources/bindings.
	ReleaseAppBindings(ctx context.Context, d Doer, appName string) (json.RawMessage, error)
	// SwitchSupportType changes a device's support type (1.12.6+):
	// PUT /api/compute-resources/nodes/{node}/devices/{device}/support-type.
	SwitchSupportType(ctx context.Context, d Doer, node, deviceID, supportType string) (json.RawMessage, error)
}

// OverlayOps groups the overlay-gateway operations. This is the "brand new
// feature" version axis: the overlay gateway (macvlan/bridge LAN-IP assignment
// for supported apps) landed in Olares 1.12.6; older backends have no such
// routes. All methods therefore return *ErrUnsupportedVersion below 1.12.6.
// Like ComputeOps, each method returns the unwrapped `data` payload.
type OverlayOps interface {
	// OverlayGatewayStatus reports gateway + per-app overlay status (1.12.6+):
	// GET /api/system/overlay-gateway-status/{user}.
	OverlayGatewayStatus(ctx context.Context, d Doer, user string) (json.RawMessage, error)
	// EnableOverlayGateway turns the gateway on (1.12.6+, owner-only):
	// POST /api/command/enable-overlay-gateway (no body).
	EnableOverlayGateway(ctx context.Context, d Doer) (json.RawMessage, error)
	// DisableOverlayGateway turns the gateway off (1.12.6+, owner-only):
	// POST /api/command/disable-overlay-gateway (no body).
	DisableOverlayGateway(ctx context.Context, d Doer) (json.RawMessage, error)
}

// OlaresClient is the umbrella interface a version implementation satisfies.
// It is intentionally composed of small capability interfaces (VersionAware,
// MarketAppOps, ...) so it can grow feature-by-feature without becoming a
// single monster interface — addressing the "interface bloat" risk.
type OlaresClient interface {
	VersionAware
	MarketAppOps
	ComputeOps
	OverlayOps
}

// ErrUnsupportedVersion is returned by a version implementation when the
// connected backend does not provide a capability the command needs (the
// "capability gate"). The dispatch aspect (cmdutil.Factory.WithOlaresClient)
// renders it into an actionable hint and does NOT retry — retrying cannot
// close a capability gap.
type ErrUnsupportedVersion struct {
	// Feature is a short, user-facing name of the operation, e.g.
	// "market clone --entrance".
	Feature string
	// MinVersion is the lowest Olares version that supports Feature.
	MinVersion *semver.Version
	// Current is the detected backend version (may be nil if unknown).
	Current *semver.Version
}

func (e *ErrUnsupportedVersion) Error() string {
	feature := e.Feature
	if feature == "" {
		feature = "this operation"
	}
	cur := "unknown"
	if e.Current != nil {
		cur = e.Current.String()
	}
	if e.MinVersion != nil {
		return fmt.Sprintf("%s requires Olares >= %s (connected backend is %s)", feature, e.MinVersion.String(), cur)
	}
	return fmt.Sprintf("%s is not supported by the connected backend (%s)", feature, cur)
}
