// Package v1_12_6 builds the `market stop` request body for Olares 1.12.6+
// (TermiPass PR #1162): POST /apps/stop with {app_name, source, all}. It is a
// pure request builder — it returns (method, path, body) and never touches the
// network, so it has no dependency on the market package (no import cycle).
package v1_12_6

import "net/http"

// Stop returns the 1.12.6 stop request. Unlike 1.12.5 the body carries
// app_name + source (StopRequest extends BaseOperationRequest).
func Stop(appName, source string, all bool) (method, path string, body any) {
	return http.MethodPost, "/apps/stop", map[string]any{
		"app_name": appName,
		"source":   source,
		"all":      all,
	}
}
