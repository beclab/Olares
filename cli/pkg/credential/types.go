// Package credential is the orchestration layer that turns a
// cliconfig.ProfileConfig + a stored token into a fully-resolved view that
// command code can consume without touching disk directly.
//
// The package is intentionally small in Phase 1: a Provider interface, a
// chained CredentialProvider, a DefaultProvider that reads
// ~/.olares-cli/{config,tokens}.json, and an EnvProvider stub for the future
// in-cluster (sandbox) scenario. Phase 2 adds keychain + automatic refresh
// inside DefaultProvider; the interface stays stable.
package credential

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// ResolvedProfile is the "ready to make an API call" view of a profile —
// analogous to lark-cli's CliConfig. Command code interacts only with this
// struct so that swapping in an EnvProvider later requires zero changes
// upstream.
type ResolvedProfile struct {
	Name       string // alias, falls back to OlaresID
	OlaresID   string
	UserUID    string

	AuthURL    string
	VaultURL   string
	DesktopURL string
	// SettingsURL is the per-user Settings SPA origin
	// (https://settings.<terminus>). The Settings CLI subtree must use
	// this rather than DesktopURL because the desktop nginx
	// (apps/docker/system-frontend/nginx/desktop.conf) does NOT forward
	// "/headscale/*", "/apis/backup/*", "/admin/*" etc. to the
	// settings/backup/secret backends — only the settings nginx
	// (settings.conf) does. See olares.ID.SettingsURL for the full
	// rationale and KNOWN_ISSUES.md KI-12 / KI-16 for the regression
	// that exposed this distinction.
	SettingsURL string
	FilesURL    string
	MarketURL   string

	AccessToken string
	// ExpiresAt is the unix-seconds expiry decoded from AccessToken's `exp`
	// claim. Zero means "no exp claim found" and is treated as "trust the
	// token until the server says otherwise".
	ExpiresAt int64

	// Source identifies which Provider produced this ResolvedProfile (for
	// diagnostics: "default", "env", ...).
	Source string

	// InsecureSkipVerify is forwarded from the underlying ProfileConfig so
	// HTTP clients constructed against this profile honor the dev override.
	InsecureSkipVerify bool
}

// Provider is implemented by anything that can turn a ProfileConfig (which
// may be nil for env-driven providers) into a ResolvedProfile. Returning
// (nil, nil) means "I don't claim this profile, try the next provider".
//
// The `profile` argument is provided by the orchestrating CredentialProvider:
// it's the currently-selected ProfileConfig from cliconfig (or nil when none
// exists). EnvProvider may ignore it entirely; DefaultProvider requires it.
type Provider interface {
	Name() string
	Resolve(ctx context.Context, profile *cliconfig.ProfileConfig) (*ResolvedProfile, error)
}
