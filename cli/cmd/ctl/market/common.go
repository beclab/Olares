package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	defaultCatalogSource = "market.olares"
	// chartUploadSource is the hard-coded local source for `market
	// upload` and `market delete`. The CLI used to accept
	// `-s {upload|studio|cli}` here, but in practice every user-driven
	// chart push belongs in the SPA's "Local Sources → Upload" bucket
	// (`upload`), and offering the other two surfaced an avoidable
	// foot-gun: a chart uploaded to `cli` was invisible to the SPA
	// despite using the same backend. We collapse to `upload` so
	// `upload` / `delete` and the SPA's Local Sources tab refer to the
	// exact same bucket; `studio` / `cli` are still valid source ids
	// for read-only verbs (`market list -s cli`) but no longer
	// reachable as a write target through the CLI.
	chartUploadSource = "upload"
)

func resolveCatalogSource(opts *MarketOptions) string {
	if s := strings.TrimSpace(opts.Source); s != "" {
		return s
	}
	return defaultCatalogSource
}

// resolveInstalledSource determines which market source an installed app
// belongs to. The 1.12.6 stop/resume/uninstall wire format requires `source`
// in the request body (TermiPass PR #1162), but those verbs don't expose
// `-s` (source is implicit). An explicit --source wins when present;
// otherwise it is read from the per-user state row via /market/state. Returns
// a clear error when the app has no installed row, so the caller can fail
// fast instead of sending a request the backend will reject for a missing
// source.
//
// "Installed" mirrors the SPA's appStore.findAppByName(): it skips rows whose
// state is in `uninstalledAppStates` (`!uninstalledApp(status)`), so a row
// that lingers in /market/state in a terminal not-installed state
// (installFailed, uninstalled, downloadFailed, the *Canceled variants — see
// notInstalledStates / isInstalledState in types.go) is treated as "not
// installed" rather than yielding a source for a request the backend will
// reject. An explicit --source bypasses this guard (the user is asserting the
// source themselves).
func resolveInstalledSource(ctx context.Context, opts *MarketOptions, mc *MarketClient, appName string) (string, error) {
	if s := strings.TrimSpace(opts.Source); s != "" {
		return s, nil
	}
	row, err := resolveInstalledRow(ctx, mc, appName)
	if err != nil {
		return "", err
	}
	return row.Source, nil
}

// resolveInstalledRow returns the full per-user state row for an installed
// app, applying the same "must be installed" guards as resolveInstalledSource
// (no --source override path — callers that need the whole row, like restart's
// --watch baseline capture, always resolve the row from /market/state). It is
// the single source of truth for the not-installed / wrong-state error
// messages so resolveInstalledSource and restart stay in lockstep.
func resolveInstalledRow(ctx context.Context, mc *MarketClient, appName string) (*installedAppRow, error) {
	row, err := lookupInstalledApp(ctx, mc, appName)
	if err != nil {
		return nil, err
	}
	if row == nil || strings.TrimSpace(row.Source) == "" {
		return nil, fmt.Errorf("%q is not installed for this user (run `olares-cli market list --mine` to see installed apps)", appName)
	}
	if !isInstalledState(row.State) {
		return nil, fmt.Errorf("%q is not an installed app (state %q); nothing to operate on (run `olares-cli market list --mine` to see installed apps)", appName, row.State)
	}
	return row, nil
}

// resolveUpgradeSource picks the source `market upgrade` should target.
//
// Unlike stop / resume / uninstall, upgrade does expose `-s`, so an explicit
// flag always wins: pointing an upgrade at a different source is legitimate
// when a chart has been moved between them. But defaulting to the catalog is
// not, which is what upgrade used to do. An app installed from `upload` has
// no row in market.olares, so the default made version resolution read the
// wrong source ("no versions found", or worse, some unrelated app's version)
// and addressed the upgrade to a source that never held the chart.
//
// Lookup failures fall back to the catalog default rather than erroring:
// preflightUpgrade re-reads the row immediately afterwards and owns the
// not-installed / wrong-state messages, and duplicating them here would only
// make the two disagree.
func resolveUpgradeSource(ctx context.Context, opts *MarketOptions, mc *MarketClient, appName string) string {
	if s := strings.TrimSpace(opts.Source); s != "" {
		return s
	}
	if row, err := resolveInstalledRow(ctx, mc, appName); err == nil {
		opts.info("Using source: %s (the source '%s' is installed from)", row.Source, appName)
		return row.Source
	}
	source := resolveCatalogSource(opts)
	opts.info("Using source: %s", source)
	return source
}

func validateVersion(version string) error {
	if _, err := semver.StrictNewVersion(strings.TrimPrefix(version, "v")); err != nil {
		return fmt.Errorf("invalid version '%s': must be a valid semver (e.g. 1.0.0, 1.2.3)", version)
	}
	return nil
}

var envNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func validateEnvName(name string) error {
	if !envNamePattern.MatchString(name) {
		return fmt.Errorf("invalid env name '%s': must start with a letter and contain only letters, digits, and underscores", name)
	}
	return nil
}

// unknownSourceError says a -s value names no source this cluster has, and
// lists the ones it does. Without it a typo read as "app not found in source
// 'bogussource'", which sends the reader looking for a missing app.
type unknownSourceError struct {
	Source string
	Known  []string
}

func (e *unknownSourceError) Error() string {
	return fmt.Sprintf("unknown market source '%s': this cluster has %s",
		e.Source, strings.Join(e.Known, ", "))
}

// appNotInSourceError says the source exists and holds no such app. It is a
// type rather than a message because the verbs that resolve a version have to
// tell it apart from a version lookup that failed for some other reason: no
// version of an absent app is reachable, so "use --version to specify" is not
// the next step.
type appNotInSourceError struct {
	App    string
	Source string
}

func (e *appNotInSourceError) Error() string {
	return fmt.Sprintf("app '%s' not found in source '%s'", e.App, e.Source)
}

// knownSources returns the source ids this cluster actually has, read from the
// per-user source map of /market/data -- the same keys `market list -a`
// iterates. Nothing is hardcoded on purpose: --help used to publish a closed
// list that omitted market.test while that source held 247 apps, so a
// hand-written list is what produced the wrong error to begin with.
//
// Called only after a lookup has already failed, so the extra request never
// lands on a working path.
func knownSources(ctx context.Context, mc *MarketClient) []string {
	resp, err := mc.GetMarketData(ctx)
	if err != nil {
		return nil
	}
	var data MarketDataResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil || data.UserData == nil {
		return nil
	}
	sources := make([]string, 0, len(data.UserData.Sources))
	for name := range data.UserData.Sources {
		if name = strings.TrimSpace(name); name != "" {
			sources = append(sources, name)
		}
	}
	sort.Strings(sources)
	return sources
}

// resolveVersionInSource returns the version the catalog holds for an app.
//
// versionFlagNarrows says whether passing --version instead would change what
// the caller's verb does, which decides whether the failure may suggest it. It
// does for install and upgrade. It does not for delete, where --version names
// the request but removes every version regardless -- and following the
// suggestion there turns a correct failure on an absent app into a success,
// because the backend treats deleting nothing as done.
func resolveVersionInSource(mc *MarketClient, appName, source string, versionFlagNarrows bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	appInfo, err := fetchAppInfo(ctx, mc, appName, source)
	if err == nil {
		if version, _ := appInfo["version"].(string); version != "" {
			return version, nil
		}
		err = fmt.Errorf("app '%s' version not found in source '%s'", appName, source)
	}

	// A missing source or a missing app is the whole reason no version came
	// back, so it is reported as itself. Wrapping it in "cannot determine
	// version" buries the cause under a symptom.
	var unknownSource *unknownSourceError
	var notInSource *appNotInSourceError
	if errors.As(err, &unknownSource) || errors.As(err, &notInSource) {
		return "", err
	}
	if versionFlagNarrows {
		return "", fmt.Errorf("cannot determine version in source '%s': %w (use --version to specify)", source, err)
	}
	return "", fmt.Errorf("cannot determine version in source '%s': %w", source, err)
}

func fetchAppInfo(ctx context.Context, mc *MarketClient, appName, source string) (map[string]interface{}, error) {
	resp, err := mc.GetAppsInfo(ctx, []AppQueryInfo{{AppID: appName, SourceDataName: source}})
	if err != nil {
		return nil, fmt.Errorf("failed to query app info: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse app info response: %w", err)
	}

	apps, _ := result["apps"].([]interface{})
	if len(apps) == 0 {
		// An empty result is the same reply for "no such source" and "no such
		// app in it", so ask the cluster which of the two it was. A source
		// list we cannot obtain leaves the app-level answer, which is the
		// safer of the two to guess.
		if known := knownSources(ctx, mc); len(known) > 0 && !containsSource(known, source) {
			return nil, &unknownSourceError{Source: source, Known: known}
		}
		return nil, &appNotInSourceError{App: appName, Source: source}
	}

	appInfo, ok := apps[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse app '%s' info", appName)
	}
	return appInfo, nil
}

func containsSource(sources []string, source string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}
	return false
}

// appSupportsClone reports whether an app can be cloned. A regular
// multi-instance app advertises allowMultipleInstall (isMultiInstanceApp); a
// template body advertises templateOnly (isTemplateApp) and is cloned to create
// instances even though its body is never installed. Either qualifies.
func appSupportsClone(appInfo map[string]interface{}) bool {
	if supported, ok := deepFindBoolValue(appInfo, "allowMultipleInstall"); ok && supported {
		return true
	}
	return appIsTemplateOnly(appInfo)
}

// isCSV2 mirrors apps/packages/app/src/constant/constants.ts `isCSV2(fullInfo)`:
//
//	fullInfo.app_info.app_entry.apiVersion === 'v2'
//	&& fullInfo.app_info.app_entry.subCharts?.length > 0
//
// This is the SAME predicate the Market SPA uses in csAppUninstall()
// (apps/.../stores/market/appService.ts) to gate the "also uninstall the
// shared server" cascade — i.e. whether to default `all: true` on the
// DELETE payload. "C/S" in this codebase means a v2 multi-chart bundle
// where the user's own chart shares server-side sub-charts with other
// users; it is NOT the same thing as "cluster-scoped" (which is gated by
// `options.appScope.clusterScoped` and used elsewhere). Keep the
// predicate in lockstep with the SPA: if the SPA tweaks isCSV2 or
// renames the apiVersion / subCharts JSON keys, update this function and
// the TestIsCSV2 table together.
func isCSV2(appInfo map[string]interface{}) bool {
	if appInfo == nil {
		return false
	}
	entry, ok := getNestedValue(appInfo, "app_info", "app_entry").(map[string]interface{})
	if !ok {
		return false
	}
	apiVersion, _ := entry["apiVersion"].(string)
	if apiVersion != "v2" {
		return false
	}
	subCharts, ok := entry["subCharts"].([]interface{})
	if !ok {
		return false
	}
	return len(subCharts) > 0
}

// isCsOrSharedFromSimple mirrors the 1.12.6 SPA's CS/shared detection, which
// moved off the full-info app_entry onto the per-app simpleInfo
// (apps/.../constant/constants.ts):
//
//	isCSV2(simple)        -> simple.app_simple_info.apiVersion === 'v2'
//	isSharedV3(simple)    -> simple.app_simple_info.shared === true
//	isCsOrSharedApp(simple) -> apiVersion === 'v2' || shared
//
// Note this is a DIFFERENT predicate from the 1.12.5 isCSV2() above, which
// reads app_info.app_entry.{apiVersion,subCharts} from the full info. On
// 1.12.6 the "also tear down / stop the shared server" cascade (the DELETE /
// stop `all` flag) is gated on this simpleInfo predicate instead. The full
// info response the CLI already fetches via fetchAppInfo carries app_simple_info
// as a sibling of app_info, so no extra endpoint is needed.
func isCsOrSharedFromSimple(appInfo map[string]interface{}) bool {
	simple, ok := appInfo["app_simple_info"].(map[string]interface{})
	if !ok {
		return false
	}
	if apiVersion, _ := simple["apiVersion"].(string); apiVersion == "v2" {
		return true
	}
	shared, _ := simple["shared"].(bool)
	return shared
}

// appIsTemplateOnly mirrors the SPA's isTemplateApp(simple) predicate
// (apps/.../constant/config.ts): a "template body" advertises
// app_simple_info.templateOnly === true. Template apps have no installable
// body — instances are created from them via clone — so they support the
// clone/create flow even when allowMultipleInstall is not set.
func appIsTemplateOnly(appInfo map[string]interface{}) bool {
	simple, ok := appInfo["app_simple_info"].(map[string]interface{})
	if !ok {
		// Fall back to a top-level flag in case the catalog item is the
		// bare simpleInfo rather than the full-info envelope.
		if v, ok2 := appInfo["templateOnly"].(bool); ok2 {
			return v
		}
		return false
	}
	v, _ := simple["templateOnly"].(bool)
	return v
}

func getNestedValue(m map[string]interface{}, keys ...string) interface{} {
	var current interface{} = m
	for _, key := range keys {
		cm, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = cm[key]
	}
	return current
}

func getNestedString(m map[string]interface{}, keys ...string) string {
	v := getNestedValue(m, keys...)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getStringValue(m map[string]interface{}, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}

func newOperationResult(mc *MarketClient, op, app, source, version, message string, resp *APIResponse) OperationResult {
	result := OperationResult{
		App:       app,
		Operation: op,
		Status:    "accepted",
		Message:   message,
		Source:    source,
		Version:   version,
	}
	if mc != nil {
		result.User = mc.olaresID
	}

	data := parseResponseData(resp)
	result.TargetApp = deepFindStringValue(data, "app_name", "appName", "uid")
	if result.TargetApp == result.App {
		result.TargetApp = ""
	}

	return result
}

func finishOperation(opts *MarketOptions, _ *MarketClient, result OperationResult) error {
	if opts.Quiet {
		return nil
	}
	opts.printResult(result)
	return nil
}

func parseEnvFlags(rawEnvs []string) ([]AppEnvVar, error) {
	if len(rawEnvs) == 0 {
		return nil, nil
	}

	var envs []AppEnvVar
	for _, raw := range rawEnvs {
		parts := strings.SplitN(raw, "=", 2)
		key := strings.TrimSpace(parts[0])
		if len(parts) != 2 || key == "" {
			return nil, fmt.Errorf("invalid env format '%s': expected KEY=VALUE", raw)
		}
		if err := validateEnvName(key); err != nil {
			return nil, err
		}
		envs = append(envs, AppEnvVar{
			EnvName: key,
			Value:   parts[1],
		})
	}
	return envs, nil
}

func parseResponseData(resp *APIResponse) map[string]interface{} {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}

	var generic interface{}
	if err := json.Unmarshal(resp.Data, &generic); err != nil {
		return map[string]interface{}{"raw": string(resp.Data)}
	}

	data, ok := generic.(map[string]interface{})
	if !ok {
		return map[string]interface{}{"value": generic}
	}

	normalizeEmbeddedJSON(data, "response")
	normalizeEmbeddedJSON(data, "result")
	return data
}

func normalizeEmbeddedJSON(data map[string]interface{}, key string) {
	raw, ok := data[key].(string)
	if !ok || raw == "" {
		return
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		data[key] = parsed
	}
}

func deepFindStringValue(data interface{}, keys ...string) string {
	switch value := data.(type) {
	case map[string]interface{}:
		for _, key := range keys {
			if s, ok := value[key].(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
		for _, child := range value {
			if found := deepFindStringValue(child, keys...); found != "" {
				return found
			}
		}
	case []interface{}:
		for _, item := range value {
			if found := deepFindStringValue(item, keys...); found != "" {
				return found
			}
		}
	}
	return ""
}

func deepFindBoolValue(data interface{}, keys ...string) (bool, bool) {
	switch value := data.(type) {
	case map[string]interface{}:
		for _, key := range keys {
			if raw, ok := value[key]; ok {
				switch v := raw.(type) {
				case bool:
					return v, true
				case string:
					switch strings.ToLower(strings.TrimSpace(v)) {
					case "true", "yes", "1":
						return true, true
					case "false", "no", "0":
						return false, true
					}
				}
			}
		}
		for _, child := range value {
			if found, ok := deepFindBoolValue(child, keys...); ok {
				return found, true
			}
		}
	case []interface{}:
		for _, item := range value {
			if found, ok := deepFindBoolValue(item, keys...); ok {
				return found, true
			}
		}
	}
	return false, false
}
