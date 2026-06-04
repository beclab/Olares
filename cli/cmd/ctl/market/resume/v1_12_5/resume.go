// Package v1_12_5 builds the `market resume` request body as it existed on
// Olares 1.12.5 (pre TermiPass PR #1162): POST /apps/resume with {appName}. It
// is a pure request builder — it returns (method, path, body) and never
// touches the network, so it has no dependency on the market package (no
// import cycle).
package v1_12_5

import "net/http"

// Resume returns the 1.12.5 resume request. source is accepted for signature
// parity with v1_12_6 but is not part of the 1.12.5 wire format.
func Resume(appName, source string) (method, path string, body any) {
	return http.MethodPost, "/apps/resume", map[string]any{
		"appName": appName,
	}
}
