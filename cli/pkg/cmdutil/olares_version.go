package cmdutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// Flag / viper keys for command-side version awareness. They are registered
// on the profile command tree (see cmd/ctl/profile/root.go); other command
// trees consume the persisted per-profile version cache.
const (
	// FlagOlaresVersion overrides detection within the profile command tree.
	// Tests may also seed the viper key directly to exercise version branches.
	FlagOlaresVersion = "olares-version"
	// FlagRefreshVersion forces a fresh /api/olares-info read, bypassing the
	// persisted per-profile cache.
	FlagRefreshVersion = "refresh-version"

	// OlaresVersionRefreshHint is the recovery path for version-gated command
	// trees. Those trees do not expose --olares-version themselves.
	OlaresVersionRefreshHint = "ensure the active profile is logged in (`olares-cli profile login`) and refresh its cached version with `olares-cli profile list --refresh-version`"
)

// OlaresBackendVersion returns the Olares OS version of the target instance,
// memoized for the lifetime of this Factory. Resolution order:
//
//  1. The profile tree's --olares-version override (or the same viper key in
//     tests; never hits the network).
//  2. Per-profile cache in ~/.olares-cli/config.json, when present and
//     --refresh-version was not set. There is no TTL: the cache is populated
//     eagerly at login and only re-read on demand (--refresh-version) or when
//     it is empty.
//  3. A fresh GET {DesktopURL}/api/olares-info via the shared whoami fetch,
//     whose osVersion is parsed and written back to the cache.
//
// Returns an error only when there is neither a usable override/cache nor a
// reachable backend.
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

// OlaresBackendAtLeast reports whether the detected backend version is >= min
// (an Olares semver such as "1.12.6"). It is the small selector that
// command-side version branching uses instead of hand-writing a comparison on
// every call:
//
//	atLeast126, err := f.OlaresBackendAtLeast(ctx, "1.12.6")
//	if atLeast126 { /* v1_12_6 path */ } else { /* v1_12_5 path */ }
//
// Comparison is done on the core (major.minor.patch) level — prerelease /
// build qualifiers are stripped before comparing — so a daily build like
// 1.12.6-20260327 counts as >= 1.12.6 (it IS the 1.12.6 line). This mirrors
// how pkg/upgrade/version.go normalizes versions with
// semver.New(major,minor,patch,"",""). `min` is expected to be a plain
// x.y.z patch (its own prerelease, if any, is likewise ignored).
func (f *Factory) OlaresBackendAtLeast(ctx context.Context, min string) (bool, error) {
	v, err := f.OlaresBackendVersion(ctx)
	if err != nil {
		return false, fmt.Errorf("%w; %s", err, OlaresVersionRefreshHint)
	}
	if v == nil {
		return false, fmt.Errorf("Olares backend version is unknown; %s", OlaresVersionRefreshHint)
	}
	minV, err := semver.NewVersion(min)
	if err != nil {
		return false, fmt.Errorf("invalid minimum version %q: %w", min, err)
	}
	coreV := semver.New(v.Major(), v.Minor(), v.Patch(), "", "")
	coreMin := semver.New(minV.Major(), minV.Minor(), minV.Patch(), "", "")
	return coreV.Compare(coreMin) >= 0, nil
}

// CachedOlaresBackendVersion returns the active profile's persisted backend
// version WITHOUT any network call. It is the read used at command-tree
// construction time to drive version-aware help / subcommand visibility (e.g.
// `settings gpu`, `settings network overlay`): reading the local cache keeps
// `--help` and tab-completion offline and fast, while runtime branching still
// resolves the live value via OlaresBackendVersion.
//
// Returns (nil, false) when there is no active profile or no cached version
// yet — callers should then present their forward-looking (newest) surface and
// let the runtime check handle an actually-older backend. The cache is
// populated eagerly at `profile login` / `import`, so this is accurate in the
// common case.
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

// RefreshOlaresBackendVersion forces a fresh /api/olares-info read, updates the
// per-profile cache, and resets the in-process memoization so subsequent
// OlaresBackendVersion calls observe the new value. Returns (newVersion,
// changed, err) where changed reports whether the version differs from what
// was previously cached.
func (f *Factory) RefreshOlaresBackendVersion(ctx context.Context) (*semver.Version, bool, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, false, err
	}
	prev, _ := loadCachedBackendVersion(rp.OlaresID)
	fetched, err := f.fetchBackendVersion(ctx, rp)
	if err != nil {
		return nil, false, err
	}
	changed := prev == nil || prev.Original() != fetched.Original()
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

	// 2. Per-profile cache (no TTL; skipped only when a refresh is forced).
	if !viper.GetBool(FlagRefreshVersion) {
		if cached, _ := loadCachedBackendVersion(rp.OlaresID); cached != nil {
			return cached, nil
		}
	}

	// 3. Fetch from the backend (writes the cache as a side effect).
	fetched, err := f.fetchBackendVersion(ctx, rp)
	if err != nil {
		return nil, fmt.Errorf("detect Olares backend version: %w", err)
	}
	return fetched, nil
}

// fetchBackendVersion reads osVersion from /api/olares-info on the profile's
// desktop ingress (through the shared whoami.FetchOlaresInfo path) using the
// Factory's auth-injecting http.Client, then best-effort caches it.
func (f *Factory) fetchBackendVersion(ctx context.Context, rp *credential.ResolvedProfile) (*semver.Version, error) {
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	doer := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)

	info, err := whoami.FetchOlaresInfo(ctx, doer)
	if err != nil {
		return nil, err
	}
	osVersion := strings.TrimSpace(info.OsVersion)
	if osVersion == "" {
		return nil, fmt.Errorf("/api/olares-info returned an empty osVersion")
	}
	v, err := utils.ParseOlaresVersionString(osVersion)
	if err != nil {
		return nil, fmt.Errorf("parse backend osVersion %q: %w", osVersion, err)
	}

	// Persist best-effort; a write failure (e.g. an env-only profile that
	// isn't in config.json) must not break the command.
	if cfg, lerr := cliconfig.LoadMultiProfileConfig(); lerr == nil {
		if _, serr := cfg.SetBackendVersion(rp.OlaresID, v.Original(), time.Now().Unix()); serr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not cache Olares backend version: %v\n", serr)
		}
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
