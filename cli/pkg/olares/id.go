// Package olares contains primitives shared across olares-cli that don't fit
// into a more specific subpackage.
//
// id.go: parse an Olares ID (e.g. "alice@olares.com") and derive the URLs of
// the per-user services that olares-cli talks to (auth / vault / desktop /
// files / market / dashboard).
//
// Background: every Olares user is identified by an "olaresId" of the form
// "<local>@<domain>". The terminus name is the same identity rendered with a
// dot ("<local>.<domain>"), and per-user service hostnames are constructed by
// prefixing that terminus name with the service subdomain (auth / vault /
// desktop / ...). The optional `localURLPrefix` is a dev-only knob that
// inserts an extra label between the service subdomain and the terminus name
// (used for staging / local DNS overrides). See pkg/wizard/user_store.go for
// the original derivation that the web app and CLI must stay in sync with.
package olares

import (
	"fmt"
	"strings"
)

// DefaultDomain is the fallback domain used when an olaresId has no `@<domain>`
// suffix. Mirrors `TerminusDefaultDomain` in pkg/wizard/user_store.go.
const DefaultDomain = "olares.com"

// ID is an opaque wrapper around an olaresId string. Construct one with
// ParseID; the zero value is invalid.
type ID string

// ParseID validates a raw olaresId string and returns it as an ID. An empty
// string or a value containing more than one `@` is rejected.
func ParseID(raw string) (ID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("olaresId is empty")
	}
	if strings.Count(raw, "@") > 1 {
		return "", fmt.Errorf("olaresId %q contains more than one '@'", raw)
	}
	return ID(raw), nil
}

// String returns the canonical olaresId string.
func (id ID) String() string { return string(id) }

// Local returns the part before `@`. For an unqualified id (no `@`) the whole
// id is returned.
func (id ID) Local() string {
	s := string(id)
	if i := strings.Index(s, "@"); i >= 0 {
		return s[:i]
	}
	return s
}

// Domain returns the part after `@`, falling back to DefaultDomain when absent.
func (id ID) Domain() string {
	s := string(id)
	if i := strings.Index(s, "@"); i >= 0 {
		return s[i+1:]
	}
	return DefaultDomain
}

// TerminusName renders the id with `.` instead of `@`, e.g. "alice.olares.com".
func (id ID) TerminusName() string {
	return id.Local() + "." + id.Domain()
}

// AuthURL returns the per-user Authelia base URL, e.g.
// "https://auth.alice.olares.com". `localPrefix` may be empty; when set it is
// inserted between the `auth.` subdomain and the terminus name (no trailing
// dot is added — callers pass e.g. "dev." for "auth.dev.alice.olares.com").
func (id ID) AuthURL(localPrefix string) string {
	return fmt.Sprintf("https://auth.%s%s", localPrefix, id.TerminusName())
}

// VaultURL returns the per-user vault base URL with the conventional `/server`
// suffix, e.g. "https://vault.alice.olares.com/server".
func (id ID) VaultURL(localPrefix string) string {
	return fmt.Sprintf("https://vault.%s%s/server", localPrefix, id.TerminusName())
}

// DesktopURL returns the per-user desktop base URL, e.g.
// "https://desktop.alice.olares.com".
//
// IMPORTANT: this is the desktop launcher SPA's origin (see
// apps/docker/system-frontend/nginx/desktop.conf). Its nginx exposes a
// minimal location set ("/api", "/server", "/notification", "/video",
// "/api/{device,logout,refresh}", "/kapis", "/seahub", "/ws"). Code that
// targets verbs covered by the Settings SPA — anything under "/headscale",
// "/apis/backup", "/admin", "/drive", "/vault", "/images",
// "/api/cloud/sign" — must use SettingsURL instead. See KNOWN_ISSUES.md
// KI-12 / KI-16 for the regression that surfaced this distinction.
func (id ID) DesktopURL(localPrefix string) string {
	return fmt.Sprintf("https://desktop.%s%s", localPrefix, id.TerminusName())
}

// SettingsURL returns the per-user settings SPA base URL, e.g.
// "https://settings.alice.olares.com".
//
// This is the origin the Settings SPA uses (see
// settings/src/application/settings.ts:42 — `tokenStore.setUrl(window.location.origin)`).
// Its nginx (apps/docker/system-frontend/nginx/settings.conf) is the
// superset that fans out to user-service ("/api", "/headscale",
// "/api/cloud/sign"), backup-server ("/apis/backup"), Infisical
// ("/admin"), files-service ("/drive", "/api/resources"),
// vault-admin-server ("/vault"), and the image proxy ("/images").
//
// olares-cli's `settings ...` command tree should always be built against
// this URL — DesktopURL appears to "work" for ~90% of verbs only because
// both nginx configs forward "/api/*" to the same user-service upstream.
// Verbs that hit any settings-only location regressed to HTML / 404 when
// they were hard-wired to DesktopURL (KI-12 vpn devices, KI-16 backup /
// restore plans, KI-6 apps secrets).
func (id ID) SettingsURL(localPrefix string) string {
	return fmt.Sprintf("https://settings.%s%s", localPrefix, id.TerminusName())
}

// FilesURL returns the per-user files-backend base URL, e.g.
// "https://files.alice.olares.com". Mirrors the web app's
// `getModuleSever('files')` derivation in
// apps/packages/app/src/stores/user.ts.
func (id ID) FilesURL(localPrefix string) string {
	return fmt.Sprintf("https://files.%s%s", localPrefix, id.TerminusName())
}

// MarketURL returns the per-user market base URL, e.g.
// "https://market.alice.olares.com". The Market app-store v2 API is reachable
// at `<MarketURL>/app-store/api/v2` — the same origin/path the Market SPA
// itself talks to via apps/packages/app/src/stores/market/center.ts. The same
// edge auth chain (Authelia + l4-bfl-proxy) accepts the `X-Authorization`
// header here as it does for files / vault / desktop.
func (id ID) MarketURL(localPrefix string) string {
	return fmt.Sprintf("https://market.%s%s", localPrefix, id.TerminusName())
}

// DashboardURL returns the per-user dashboard BFF base URL, e.g.
// "https://dashboard.alice.olares.com". The dashboard SPA itself is served
// here, and its Koa-based ks-apiserver-proxy backend is reachable at the
// same origin under `/capi/*` (custom aggregation), `/(k)?api(s)?/*`
// (wildcard kubesphere proxy), `/user-service/*` (BFL) and `/hami/*`
// (HAMI vGPU). Same edge auth chain as the other per-user URLs —
// X-Authorization is honored and 401/403 are recoverable via /api/refresh.
// olares-cli's `dashboard` command tree talks to this URL.
func (id ID) DashboardURL(localPrefix string) string {
	return fmt.Sprintf("https://dashboard.%s%s", localPrefix, id.TerminusName())
}
