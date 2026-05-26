package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// isUpgradableStates is the verbatim mirror of the SPA's
// `isUpgradableAppStates` set in apps/.../constant/config.ts:
//
//	APP_STATUS.RUNNING            -> "running"
//	APP_STATUS.STOP.COMPLETED     -> "stopped"
//	APP_STATUS.STOP.FAILED        -> "stopFailed"
//	APP_STATUS.UPGRADE.FAILED     -> "upgradeFailed"
//	APP_STATUS.ENV.APPLY_FAILED   -> "applyEnvFailed"
//
// These are the only five states the SPA's `canUpgrade()` predicate
// (also config.ts) will accept before showing the "Upgrade" button.
// The CLI's `market upgrade` preflight refuses every other state up
// front so the call never reaches the backend in a knowingly-bad
// state — exactly the same surface area the user sees in the Market
// UI. Keep this map in lockstep with the SPA: any reshuffle of the
// SPA's upgradable states must be reflected here and in
// TestIsUpgradableState.
var isUpgradableStates = map[string]struct{}{
	"running":        {},
	"stopped":        {},
	"stopFailed":     {},
	"upgradeFailed":  {},
	"applyEnvFailed": {},
}

// isUpgradable reports whether an app currently in `state` can have
// `market upgrade` issued against it per the SPA's `canUpgrade` gate.
// Empty / unknown states return false (same as the SPA's button being
// hidden) so the preflight bails with an actionable error rather than
// firing an upgrade we know the backend will reject.
func isUpgradable(state string) bool {
	if state == "" {
		return false
	}
	_, ok := isUpgradableStates[state]
	return ok
}

// upgradableStateList returns the SPA's allowed states in a stable
// human-friendly order, used to render preflight errors like
// "allowed states: running, stopped, stopFailed, upgradeFailed,
// applyEnvFailed".
func upgradableStateList() string {
	return "running, stopped, stopFailed, upgradeFailed, applyEnvFailed"
}

// installedAppRow captures the subset of /market/state fields the
// upgrade preflight needs. Built by lookupInstalledApp from the raw
// MarketStateResponse — kept narrow so the predicate logic stays
// straightforward.
type installedAppRow struct {
	Name    string
	RawName string
	Source  string
	State   string
	Version string
}

// lookupInstalledApp finds the active profile's per-user state row
// for appName via /market/state. Returns nil when the row genuinely
// doesn't exist (vs an HTTP / parse error, which is surfaced).
//
// MATCHING RULE — strict on Name, NOT RawName:
//
// We match only `row.Name == appName`. We deliberately do NOT match
// `row.RawName == appName`, because RawName aliases every clone back
// to its source app (a clone like `windowsefe992` carries
// `RawName=windows`). If the user types `windows` AND both the
// primary `windows` row and one or more clones of `windows` are
// installed, a RawName-match would also catch each clone — and the
// function would return whichever the slice / map iteration happened
// to surface first. That's non-deterministic across slice order
// inside one source and across the `Sources` map iteration (Go map
// iteration is randomized). The consequence is that callers like
// `preflightUpgrade` (gate 2 / gate 3 read State + Version) and
// `shouldAutoCascade` (reads Source for the catalog probe) silently
// see the WRONG row's data — upgrades fail with a misleading state
// message, auto-cascade looks up isCSV2 in the wrong source's
// catalog, etc.
//
// The SPA never sends source-app names from clone cards either:
// clicking "Upgrade" / "Uninstall" / "Stop" on a clone card invokes
// the corresponding endpoint with the per-instance name
// (`windowsefe992`). The CLI's contract is the same: type
// `windowsefe992` to operate on the clone, `windows` to operate on
// the primary. Anything else is the wrong row.
//
// The returned struct still surfaces `RawName` (we read it from the
// matched row's status) so callers needing catalog metadata can look
// up the source app for clones — same trick fetchInstalledApps and
// preflightUpgrade use for the /apps lookup key. RawName is data on
// the return path, never a match key on the lookup path.
func lookupInstalledApp(ctx context.Context, mc *MarketClient, appName string) (*installedAppRow, error) {
	resp, err := mc.GetMarketState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read user's market state: %w", err)
	}

	var data MarketStateResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse market state response: %w", err)
	}
	if data.UserData == nil {
		return nil, nil
	}

	for sourceName, sourceData := range data.UserData.Sources {
		sourceName = strings.TrimSpace(sourceName)
		if sourceName == "" || sourceData == nil {
			continue
		}
		for _, appState := range sourceData.AppStateLatest {
			rowName := strings.TrimSpace(appState.Status.Name)
			rawName := strings.TrimSpace(appState.Status.RawName)
			if rowName != appName {
				// Defensive: legacy / malformed rows that omit Name
				// but populate RawName can still be identified IF
				// no other row in the same source is going to claim
				// this name first. We only accept that fallback
				// when Name is empty — never when Name is populated
				// with something else (which would mean this is a
				// clone whose source app happens to match the
				// user's input, the exact ambiguity documented in
				// the lookupInstalledApp doc above).
				if rowName != "" || rawName != appName {
					continue
				}
			}
			canonical := rowName
			if canonical == "" {
				canonical = rawName
			}
			return &installedAppRow{
				Name:    canonical,
				RawName: rawName,
				Source:  sourceName,
				State:   appState.Status.State,
				Version: strings.TrimSpace(appState.Version),
			}, nil
		}
	}
	return nil, nil
}

// isAppSuspended mirrors the SPA's `suspendApp(simpleLatest)` predicate
// in apps/.../constant/config.ts:
//
//	app_simple_info.app_labels.includes('suspend')
//	|| app_simple_info.app_labels.includes('remove')
//
// The SPA refuses to render the Upgrade button when either label is
// present (the app is marked as withdrawn / suspended in the
// catalog). The CLI preflight mirrors that so a `market upgrade <X>`
// against a removed chart fails locally with a clear message instead
// of producing a confusing backend error.
//
// The label list lives under `app_simple_info.app_labels` in BOTH
// catalog responses the CLI consumes — /market/data
// (`AppSimpleInfoLatest`) and /apps (`AppFullInfoLatest`) — because
// `app_simple_info` is a shared sub-document in both shapes (see
// `AppFullInfoLatest` in apps/.../constant/constants.ts). We read
// from the /apps response in the preflight since the upgrade path
// already fetches that document for version resolution.
func isAppSuspended(appInfo map[string]interface{}) bool {
	for _, label := range appLabels(appInfo) {
		switch label {
		case "suspend", "remove":
			return true
		}
	}
	return false
}

func appLabels(appInfo map[string]interface{}) []string {
	if appInfo == nil {
		return nil
	}
	raw, ok := getNestedValue(appInfo, "app_simple_info", "app_labels").([]interface{})
	if !ok {
		return nil
	}
	labels := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				labels = append(labels, s)
			}
		}
	}
	return labels
}

// preflightUpgrade mirrors the SPA's `canUpgrade(statusLatest, appId,
// sourceId)` predicate in apps/.../constant/config.ts — same four
// gates, same order:
//
//  1. state ∈ isUpgradableStates  (`isUpgradable`)
//  2. installed + target version both present
//  3. target version > installed (strict semver)
//  4. catalog row is NOT marked `suspend` / `remove` (`isAppSuspended`)
//
// Returns nil on pass. On any explicit gate failure returns a typed
// error with a self-contained message — failOp will format it in
// table mode or surface it as the `message` field in -o json mode.
//
// `source` is the catalog source the CLI is about to send the
// /apps/{name}/upgrade request to, NOT the source the app was
// installed from. They are usually the same (resolveCatalogSource
// defaults to market.olares and the SPA always installs from there
// too) but the user can `-s` to point upgrade at a different source.
// When the installed row's source disagrees with the upgrade source
// we surface a warning rather than bailing — sometimes legitimate
// (chart was moved between sources) but worth flagging.
//
// On transient catalog lookup errors (network blip, /apps shape
// surprise) we soft-fail: log a warning, skip the suspend-label
// gate, and let the upgrade proceed. Same soft-fail philosophy as
// shouldAutoCascade in uninstall.go — the backend has the final
// say and we don't want a flaky probe to block the user. The hard
// gates (state / version) never soft-fail because their inputs come
// from the /market/state response we already have in hand for the
// row lookup.
func preflightUpgrade(ctx context.Context, opts *MarketOptions, mc *MarketClient, appName, targetVersion, source string) error {
	row, err := lookupInstalledApp(ctx, mc, appName)
	if err != nil {
		return fmt.Errorf("preflight: %w", err)
	}
	if row == nil {
		return fmt.Errorf("cannot upgrade '%s': app is not installed (no per-user state row); use 'olares-cli market install %s' first", appName, appName)
	}

	if !isUpgradable(row.State) {
		return fmt.Errorf(
			"cannot upgrade '%s' in state '%s': upgrade is only allowed from %s; current state may be transient (in-flight install / uninstall) — re-run 'olares-cli market status %s' to confirm",
			appName, row.State, upgradableStateList(), appName,
		)
	}

	if row.Version == "" {
		return fmt.Errorf("cannot upgrade '%s': no version recorded on the state row (mid-flight install or older backend) — re-run 'olares-cli market status %s --watch' until the row stabilizes, then retry", appName, appName)
	}
	if targetVersion == "" {
		return fmt.Errorf("cannot upgrade '%s': target version is empty (internal error — version resolution did not produce a version string)", appName)
	}

	cmp, err := compareSemver(targetVersion, row.Version)
	if err != nil {
		return fmt.Errorf("cannot upgrade '%s': version comparison failed: %w (installed %q, target %q)", appName, err, row.Version, targetVersion)
	}
	if cmp == 0 {
		return fmt.Errorf("cannot upgrade '%s': target version '%s' is already installed — nothing to do", appName, targetVersion)
	}
	if cmp < 0 {
		return fmt.Errorf("cannot upgrade '%s': target version '%s' is older than installed version '%s'; downgrade via upgrade is rejected — uninstall and reinstall the older version instead", appName, targetVersion, row.Version)
	}

	if strings.TrimSpace(source) != "" && row.Source != "" && source != row.Source {
		opts.info("warning: upgrade is targeting source '%s' but app is currently installed from source '%s'", source, row.Source)
	}

	// Clones (where rawAppName != name, see isCloneApp in SPA) live under
	// per-instance names like `windowsefe992` but their catalog entry is
	// under the source app `windows`. Mirror fetchInstalledApps's lookup:
	// prefer RawName when present, fall back to the canonical row name.
	lookupName := row.RawName
	if lookupName == "" {
		lookupName = row.Name
	}
	if lookupName == "" {
		lookupName = appName
	}

	appInfo, err := fetchAppInfo(ctx, mc, lookupName, source)
	if err != nil {
		// Soft-fail: a transient /apps blip shouldn't block an upgrade
		// when the hard gates above (state + strict-newer version) all
		// passed. Surface as a warning so it's visible without
		// blocking the operation.
		opts.info("warning: preflight could not read catalog metadata for '%s' from source '%s' (%v); skipping suspend-label check", lookupName, source, err)
		return nil
	}
	if isAppSuspended(appInfo) {
		return fmt.Errorf("cannot upgrade '%s': chart is marked 'suspend' or 'remove' in source '%s' (the SPA hides the Upgrade button for the same reason); upstream has withdrawn this app", lookupName, source)
	}

	return nil
}

// compareSemver returns -1 / 0 / 1 comparing target vs installed using
// the same semver semantics validateVersion uses (Masterminds v3,
// strict). Strips an optional leading `v` for parity with how
// `validateVersion` already normalizes /api/users-supplied input.
func compareSemver(target, installed string) (int, error) {
	t, err := semver.StrictNewVersion(strings.TrimPrefix(target, "v"))
	if err != nil {
		return 0, fmt.Errorf("invalid target version '%s'", target)
	}
	i, err := semver.StrictNewVersion(strings.TrimPrefix(installed, "v"))
	if err != nil {
		return 0, fmt.Errorf("invalid installed version '%s'", installed)
	}
	return t.Compare(i), nil
}
