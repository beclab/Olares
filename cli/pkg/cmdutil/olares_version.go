package cmdutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/olaresclient"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// Flag / viper keys for the version-compat layer. Registered as persistent
// flags on the root command (see cmd/ctl/root.go) and bound to viper there.
const (
	// FlagOlaresVersion overrides the detected backend version. Useful when
	// /api/olares-info is unreachable, or to force a specific code path for
	// debugging. Value is any Olares semver, e.g. 1.12.6 or 1.12.6-20260603.
	FlagOlaresVersion = "olares-version"
	// FlagRefreshVersion forces a fresh /api/olares-info read, bypassing the
	// persisted per-profile cache regardless of its TTL.
	FlagRefreshVersion = "refresh-version"
)

// backendVersionTTL is how long a cached backend version is trusted before it
// is re-fetched from /api/olares-info. The token self-heals on a 401, but the
// version has no such signal, so the cache must expire. One hour balances
// avoiding a round-trip on every command against noticing an out-of-band
// backend upgrade; the dispatch aspect (WithOlaresClient) additionally
// re-checks on version-suspect errors.
const backendVersionTTL = time.Hour

// olaresInfoEnvelope mirrors the BFL `{code, message, data}` envelope returned
// by /api/olares-info, decoding only the osVersion field the version-compat
// layer needs. Same endpoint `settings me version` reads.
type olaresInfoEnvelope struct {
	Code int `json:"code"`
	Data struct {
		OsVersion string `json:"osVersion"`
	} `json:"data"`
}

// OlaresBackendVersion returns the Olares OS version of the target instance,
// memoized for the lifetime of this Factory. Resolution order:
//
//  1. --olares-version flag (explicit override; never hits the network).
//  2. Per-profile cache in ~/.olares-cli/config.json, when present and within
//     backendVersionTTL and --refresh-version was not set.
//  3. A fresh GET {DesktopURL}/api/olares-info, whose osVersion is parsed and
//     written back to the cache.
//
// If the network read fails but a (stale) cached value exists, the stale value
// is returned with a warning rather than failing the command. Only when there
// is neither a reachable backend nor any cached value does this return an error.
func (f *Factory) OlaresBackendVersion(ctx context.Context) (*semver.Version, error) {
	f.backendVersionOnce.Do(func() {
		v, err := f.resolveBackendVersion(ctx)
		f.backendVersionMu.Lock()
		f.backendVersion, f.backendVersionErr = v, err
		f.backendVersionMu.Unlock()
	})
	f.backendVersionMu.Lock()
	defer f.backendVersionMu.Unlock()
	return f.backendVersion, f.backendVersionErr
}

// CachedOlaresBackendVersion returns the active profile's persisted backend
// version WITHOUT any network call and WITHOUT enforcing the TTL. It is the
// read used at command-tree construction time to drive version-aware help /
// subcommand visibility (e.g. `settings gpu`, `settings network overlay`):
// reading the local cache keeps `--help` and tab-completion offline and fast,
// while runtime dispatch (WithOlaresClient) still enforces correctness via the
// TTL-refreshing OlaresBackendVersion.
//
// Returns (nil, false) when there is no active profile or no cached version
// yet — callers should then present their forward-looking (newest) surface and
// let the runtime capability gate handle an actually-older backend. The cache
// is populated eagerly at `profile login` / `import` and refreshed on a TTL by
// normal version-aware commands, so this is accurate in the common case.
func (f *Factory) CachedOlaresBackendVersion() (*semver.Version, bool) {
	if f == nil {
		return nil, false
	}
	rp, err := f.ResolveProfile(context.Background())
	if err != nil || rp == nil {
		return nil, false
	}
	v, _ := loadCachedBackendVersion(rp.OlaresID)
	if v == nil {
		return nil, false
	}
	return v, true
}

// BackendVersionSource describes where the dispatched backend version came
// from, for the `settings me version --dispatch` transparency view.
type BackendVersionSource string

const (
	// BackendVersionFromFlag: resolved from the --olares-version override.
	BackendVersionFromFlag BackendVersionSource = "flag (--olares-version)"
	// BackendVersionFromCache: served from the per-profile cache (within TTL).
	BackendVersionFromCache BackendVersionSource = "cache (within TTL)"
	// BackendVersionFromFetch: freshly read from /api/olares-info.
	BackendVersionFromFetch BackendVersionSource = "fetched from /api/olares-info"
	// BackendVersionFromStaleCache: fetch failed; fell back to a stale cache.
	BackendVersionFromStaleCache BackendVersionSource = "stale cache (fetch failed)"
)

// BackendVersionInfo is the transparency view for version-compat dispatch: the
// effective version, where it came from, the persisted cache state, the TTL,
// and the client implementation it selects. Returned by
// OlaresBackendVersionInfo for `settings me version --dispatch`.
type BackendVersionInfo struct {
	Version        *semver.Version
	Source         BackendVersionSource
	CachedVersion  string
	RefreshedAt    int64
	TTL            time.Duration
	Implementation string
}

// OlaresBackendVersionInfo resolves the backend version the same way the
// dispatch path does (flag > fresh cache > /api/olares-info, with stale-cache
// fallback) but additionally reports the provenance, cache timestamp, TTL and
// the selected client implementation. It honors --refresh-version. Unlike
// OlaresBackendVersion it is not memoized — it is a diagnostic read meant to
// reflect live state when the user asks.
func (f *Factory) OlaresBackendVersionInfo(ctx context.Context) (*BackendVersionInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	info := &BackendVersionInfo{TTL: backendVersionTTL}

	// 1. Explicit override.
	if raw := strings.TrimSpace(viper.GetString(FlagOlaresVersion)); raw != "" {
		v, err := utils.ParseOlaresVersionString(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --%s %q: %w", FlagOlaresVersion, raw, err)
		}
		info.Version = v
		info.Source = BackendVersionFromFlag
		info.Implementation = olaresclient.SelectedImplementation(v)
		return info, nil
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	cached, cachedAt := loadCachedBackendVersion(rp.OlaresID)
	if cached != nil {
		info.CachedVersion = cached.Original()
		info.RefreshedAt = cachedAt
	}
	forceRefresh := viper.GetBool(FlagRefreshVersion)
	fresh := cached != nil && time.Since(time.Unix(cachedAt, 0)) <= backendVersionTTL

	// 2. Fresh cache.
	if !forceRefresh && fresh {
		info.Version = cached
		info.Source = BackendVersionFromCache
		info.Implementation = olaresclient.SelectedImplementation(cached)
		return info, nil
	}

	// 3. Fetch from the backend.
	fetched, fetchErr := f.fetchBackendVersion(ctx, rp)
	if fetchErr != nil {
		if cached != nil {
			info.Version = cached
			info.Source = BackendVersionFromStaleCache
			info.Implementation = olaresclient.SelectedImplementation(cached)
			return info, nil
		}
		return nil, fmt.Errorf("detect Olares backend version: %w (pass --%s <version> to set it manually)", fetchErr, FlagOlaresVersion)
	}
	if cfg, err := cliconfig.LoadMultiProfileConfig(); err == nil {
		now := time.Now().Unix()
		if _, serr := cfg.SetBackendVersion(rp.OlaresID, fetched.Original(), now); serr == nil {
			info.RefreshedAt = now
			info.CachedVersion = fetched.Original()
		}
	}
	info.Version = fetched
	info.Source = BackendVersionFromFetch
	info.Implementation = olaresclient.SelectedImplementation(fetched)
	return info, nil
}

// RefreshOlaresBackendVersion forces a fresh /api/olares-info read, updates the
// per-profile cache, and resets the in-process memoization so subsequent
// OlaresBackendVersion calls observe the new value. Used by the dispatch
// aspect's self-heal path when a command fails with a version-suspect error.
// Returns (newVersion, changed, err) where changed reports whether the version
// differs from what was previously cached.
func (f *Factory) RefreshOlaresBackendVersion(ctx context.Context) (*semver.Version, bool, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, false, err
	}
	fetched, err := f.fetchBackendVersion(ctx, rp)
	if err != nil {
		return nil, false, err
	}
	changed := false
	if cfg, lerr := cliconfig.LoadMultiProfileConfig(); lerr == nil {
		changed, _ = cfg.SetBackendVersion(rp.OlaresID, fetched.Original(), time.Now().Unix())
	}
	f.backendVersionMu.Lock()
	f.backendVersion = fetched
	f.backendVersionErr = nil
	f.backendVersionMu.Unlock()
	return fetched, changed, nil
}

func (f *Factory) resolveBackendVersion(ctx context.Context) (*semver.Version, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 1. Explicit override.
	if raw := strings.TrimSpace(viper.GetString(FlagOlaresVersion)); raw != "" {
		v, err := utils.ParseOlaresVersionString(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --%s %q: %w", FlagOlaresVersion, raw, err)
		}
		return v, nil
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	olaresID := rp.OlaresID
	forceRefresh := viper.GetBool(FlagRefreshVersion)

	// 2. Per-profile cache (subject to TTL unless a refresh is forced).
	cached, cachedAt := loadCachedBackendVersion(olaresID)
	if !forceRefresh && cached != nil && time.Since(time.Unix(cachedAt, 0)) <= backendVersionTTL {
		return cached, nil
	}

	// 3. Fetch from the backend.
	fetched, fetchErr := f.fetchBackendVersion(ctx, rp)
	if fetchErr != nil {
		// Degrade to the (stale) cache when one exists; only the first
		// detection ever fails hard.
		if cached != nil {
			fmt.Fprintf(os.Stderr, "warning: could not refresh Olares backend version (%v); using cached %s\n", fetchErr, cached)
			return cached, nil
		}
		return nil, fmt.Errorf("detect Olares backend version: %w (pass --%s <version> to set it manually)", fetchErr, FlagOlaresVersion)
	}

	// Persist best-effort; a write failure must not break the command.
	if cfg, err := cliconfig.LoadMultiProfileConfig(); err == nil {
		if _, serr := cfg.SetBackendVersion(olaresID, fetched.Original(), time.Now().Unix()); serr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not cache Olares backend version: %v\n", serr)
		}
	}
	return fetched, nil
}

// fetchBackendVersion reads osVersion from /api/olares-info on the profile's
// desktop ingress, reusing the Factory's auth-injecting http.Client.
func (f *Factory) fetchBackendVersion(ctx context.Context, rp *credential.ResolvedProfile) (*semver.Version, error) {
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	doer := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)

	var env olaresInfoEnvelope
	if err := doer.DoJSON(ctx, http.MethodGet, "/api/olares-info", nil, &env); err != nil {
		return nil, err
	}
	osVersion := strings.TrimSpace(env.Data.OsVersion)
	if osVersion == "" {
		return nil, fmt.Errorf("/api/olares-info returned an empty osVersion")
	}
	v, err := utils.ParseOlaresVersionString(osVersion)
	if err != nil {
		return nil, fmt.Errorf("parse backend osVersion %q: %w", osVersion, err)
	}
	return v, nil
}

func loadCachedBackendVersion(olaresID string) (*semver.Version, int64) {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return nil, 0
	}
	p := cfg.FindByOlaresID(olaresID)
	if p == nil || strings.TrimSpace(p.BackendVersion) == "" {
		return nil, 0
	}
	v, err := utils.ParseOlaresVersionString(p.BackendVersion)
	if err != nil {
		return nil, 0
	}
	return v, p.BackendVersionRefreshedAt
}
