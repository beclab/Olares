package market

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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

func resolveVersionInSource(mc *MarketClient, appName, source string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	appInfo, err := fetchAppInfo(ctx, mc, appName, source)
	if err != nil {
		return "", err
	}

	version, _ := appInfo["version"].(string)
	if version == "" {
		return "", fmt.Errorf("app '%s' version not found in source '%s'", appName, source)
	}
	return version, nil
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
		return nil, fmt.Errorf("app '%s' not found in source '%s'", appName, source)
	}

	appInfo, ok := apps[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse app '%s' info", appName)
	}
	return appInfo, nil
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
