package app

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
	defaultLocalSource   = "cli"
)

var localSources = map[string]bool{
	"upload": true,
	"studio": true,
	"cli":    true,
}

func resolveCatalogSource(opts *AppOptions) string {
	if s := strings.TrimSpace(opts.Source); s != "" {
		return s
	}
	return defaultCatalogSource
}

func resolveLocalSource(opts *AppOptions) string {
	if s := strings.TrimSpace(opts.Source); s != "" {
		return s
	}
	return defaultLocalSource
}

func resolveFromSource(opts *AppOptions) string {
	if s := strings.TrimSpace(opts.FromSource); s != "" {
		return s
	}
	return defaultCatalogSource
}

func validateLocalSource(source string) error {
	if !localSources[source] {
		return fmt.Errorf("invalid local source '%s': must be one of upload, studio, cli", source)
	}
	return nil
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

func appSupportsClone(appInfo map[string]interface{}) bool {
	supported, ok := deepFindBoolValue(appInfo, "allowMultipleInstall")
	return ok && supported
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
		result.User = mc.user
	}

	data := parseResponseData(resp)
	result.TargetApp = deepFindStringValue(data, "app_name", "appName", "uid")
	if result.TargetApp == result.App {
		result.TargetApp = ""
	}

	return result
}

func finishOperation(opts *AppOptions, _ *MarketClient, result OperationResult) error {
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
