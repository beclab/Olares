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
	"github.com/beclab/Olares/cli/pkg/olares"
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
	SettingsURL string
	FilesURL     string
	MarketURL    string
	DashboardURL string
	// ControlHubURL is the per-user ControlHub BFF base URL
	// ("https://control-hub.<terminus>"). The `cluster` command tree uses
	// this — see pkg/olares/id.go::ControlHubURL for the full description
	// of which path prefixes ride this origin.
	ControlHubURL string

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

	// Location is the CLI's network position relative to this Olares (see
	// pkg/olares.Location). It determined the URLs above and selects the
	// http.Transport resolver; it must travel with the resolved profile
	// because external/host/cluster share identical URLs and differ only in
	// transport. Empty means "not yet probed" — the Factory backfills it on
	// the first command that resolves this profile.
	Location olares.Location

	// LocalURLPrefix is the dev-only URL label (forwarded from ProfileConfig).
	// Kept on the resolved profile so the reprobe path can re-derive
	// endpoints for a new Location without reloading config.
	LocalURLPrefix string

	// AuthURLOverride, when non-empty, is the user-pinned auth URL. The
	// reprobe/relocation path checks it to avoid rewriting a deliberately
	// overridden endpoint.
	AuthURLOverride string
}

// ApplyLocation re-derives every per-service URL for loc and records it on rp.
// The pinned auth URL override (when set) is preserved. Used by the Factory's
// lazy backfill and reprobe paths to switch a resolved profile to a freshly
// detected Location without reloading config. A malformed OlaresID is a
// silent no-op (it would already have failed earlier resolution).
func (rp *ResolvedProfile) ApplyLocation(loc olares.Location) {
	id, err := olares.ParseID(rp.OlaresID)
	if err != nil {
		return
	}
	ep := id.Endpoints(loc, rp.LocalURLPrefix)
	if rp.AuthURLOverride == "" {
		rp.AuthURL = ep.Auth
	}
	rp.VaultURL = ep.Vault
	rp.DesktopURL = ep.Desktop
	rp.SettingsURL = ep.Settings
	rp.FilesURL = ep.Files
	rp.MarketURL = ep.Market
	rp.DashboardURL = ep.Dashboard
	rp.ControlHubURL = ep.ControlHub
	rp.Location = loc
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
