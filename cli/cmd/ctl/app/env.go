package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	sysv1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	appserviceapi "github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
)

type envValidationError struct {
	AppName       string
	MissingValues []string
	MissingRefs   []string
	InvalidValues []string
}

func (e *envValidationError) Error() string {
	return formatEnvValidationError(e)
}

func decodeAppEnvSpecs(raw interface{}) []sysv1alpha1.AppEnvVar {
	if raw == nil {
		return nil
	}

	payload, err := json.Marshal(raw)
	if err != nil {
		return nil
	}

	var specs []sysv1alpha1.AppEnvVar
	if err := json.Unmarshal(payload, &specs); err != nil {
		return nil
	}
	return specs
}

func formatEnvValidationError(e *envValidationError) string {
	var b strings.Builder
	b.WriteString("environment variable requirements not met\n")

	if len(e.MissingValues) > 0 {
		b.WriteString(fmt.Sprintf("\n  Missing required values: %s\n", strings.Join(e.MissingValues, ", ")))
	}

	if len(e.MissingRefs) > 0 {
		b.WriteString(fmt.Sprintf("\n  Missing referenced values: %s\n", strings.Join(e.MissingRefs, ", ")))
	}

	if len(e.InvalidValues) > 0 {
		b.WriteString(fmt.Sprintf("\n  Invalid values: %s\n", strings.Join(e.InvalidValues, ", ")))
	}

	b.WriteString("\nRun 'olares-cli app get ")
	b.WriteString(e.AppName)
	b.WriteString("' to inspect the declared envs, then use --env KEY=VALUE to provide or correct values.")
	return b.String()
}

func formatAppEnvDetails(specs []sysv1alpha1.AppEnvVar) string {
	if len(specs) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Envs:\n")
	for _, spec := range specs {
		writeAppEnvDetail(&b, spec)
	}
	return strings.TrimRight(b.String(), "\n")
}

func writeAppEnvDetail(b *strings.Builder, spec sysv1alpha1.AppEnvVar) {
	status := "optional"
	if spec.Required {
		status = "required"
	}

	b.WriteString(fmt.Sprintf("  - %s", spec.EnvName))
	if spec.Type != "" {
		b.WriteString(fmt.Sprintf("  (%s, type: %s)", status, spec.Type))
	} else {
		b.WriteString(fmt.Sprintf("  (%s)", status))
	}
	b.WriteString("\n")

	if spec.Title != "" && spec.Title != spec.EnvName {
		b.WriteString(fmt.Sprintf("      title: %s\n", spec.Title))
	}
	if spec.Description != "" {
		b.WriteString(fmt.Sprintf("      description: %s\n", spec.Description))
	}
	if spec.ValueFrom != nil && strings.TrimSpace(spec.ValueFrom.EnvName) != "" {
		b.WriteString(fmt.Sprintf("      referenced from: %s\n", spec.ValueFrom.EnvName))
	} else if spec.Default != "" {
		b.WriteString(fmt.Sprintf("      default: %s\n", spec.Default))
	}
	writeEnvConstraints(b, spec)
}

func writeEnvConstraints(b *strings.Builder, spec sysv1alpha1.AppEnvVar) {
	if len(spec.Options) > 0 {
		b.WriteString(fmt.Sprintf("      options: %s\n", formatOptionsInline(spec.Options)))
	}
	if spec.RemoteOptions != "" {
		remoteOpts, err := tryFetchRemoteOptions(spec.RemoteOptions)
		if err == nil && len(remoteOpts) > 0 {
			b.WriteString(fmt.Sprintf("      options (remote): %s\n", formatOptionsInline(remoteOpts)))
		} else {
			b.WriteString(fmt.Sprintf("      options: (fetch from) %s\n", spec.RemoteOptions))
		}
	}
	if spec.Regex != "" {
		b.WriteString(fmt.Sprintf("      pattern: %s\n", spec.Regex))
	}
}

func formatOptionsInline(options []sysv1alpha1.EnvValueOptionItem) string {
	const maxShow = 10
	items := make([]string, 0, len(options))
	for i, opt := range options {
		if i >= maxShow {
			items = append(items, fmt.Sprintf("... and %d more", len(options)-maxShow))
			break
		}
		if opt.Title != "" && opt.Title != opt.Value {
			items = append(items, fmt.Sprintf("%s (%s)", opt.Value, opt.Title))
		} else {
			items = append(items, opt.Value)
		}
	}
	return strings.Join(items, ", ")
}

func tryFetchRemoteOptions(endpoint string) ([]sysv1alpha1.EnvValueOptionItem, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []sysv1alpha1.EnvValueOptionItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// parseServerEnvError extracts env validation details from a failed API response.
// This handles the case where the market service wraps app-service's 422 response.
func parseServerEnvError(resp *APIResponse, appName string) *envValidationError {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}

	data := parseResponseData(resp)
	checkResult := extractServerEnvCheckResult(data)
	if checkResult == nil {
		return nil
	}

	result := &envValidationError{
		AppName:       appName,
		MissingValues: envNames(checkResult.MissingValues),
		MissingRefs:   envNames(checkResult.MissingRefs),
		InvalidValues: envNames(checkResult.InvalidValues),
	}

	if len(result.MissingValues) == 0 && len(result.MissingRefs) == 0 && len(result.InvalidValues) == 0 {
		return nil
	}
	return result
}

func extractServerEnvCheckResult(data map[string]interface{}) *appserviceapi.AppEnvCheckResult {
	if data == nil {
		return nil
	}

	checkPayload := data
	if backendResp, ok := data["backend_response"].(map[string]interface{}); ok {
		backendData, ok := backendResp["data"].(map[string]interface{})
		if !ok {
			return nil
		}
		checkPayload = backendData
	}

	checkType, _ := checkPayload["type"].(string)
	if checkType != appserviceapi.CheckTypeAppEnv {
		return nil
	}

	if nested, ok := checkPayload["Data"].(map[string]interface{}); ok {
		checkPayload = nested
	}

	payload, err := json.Marshal(checkPayload)
	if err != nil {
		return nil
	}

	var result appserviceapi.AppEnvCheckResult
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil
	}
	return &result
}

func envNames(envs []sysv1alpha1.AppEnvVar) []string {
	names := make([]string, 0, len(envs))
	for _, env := range envs {
		if strings.TrimSpace(env.EnvName) == "" {
			continue
		}
		names = append(names, env.EnvName)
	}
	return names
}
