// Package v1_12_6 builds the `market resume` request body for Olares 1.12.6+
// (TermiPass PR #1162): POST /apps/resume with {app_name, source}. It is a
// pure request builder — it returns (method, path, body) and never touches the
// network, so it has no dependency on the market package (no import cycle).
package v1_12_6

import "net/http"

// Resume returns the 1.12.6 resume request. Unlike 1.12.5 the body carries
// app_name + source (Resume2RestartRequest extends BaseOperationRequest; the
// optional computeBinding is not exposed by the CLI).
func Resume(appName, source string) (method, path string, body any) {
	return http.MethodPost, "/apps/resume", map[string]any{
		"app_name": appName,
		"source":   source,
	}
}
