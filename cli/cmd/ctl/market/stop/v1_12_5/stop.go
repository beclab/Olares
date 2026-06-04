// Package v1_12_5 builds the `market stop` request body as it existed on
// Olares 1.12.5 (pre TermiPass PR #1162): POST /apps/stop with {appName, all}.
// It is a pure request builder — it returns (method, path, body) and never
// touches the network, so it has no dependency on the market package (no
// import cycle).
package v1_12_5

import "net/http"

// Stop returns the 1.12.5 stop request. source is accepted for signature
// parity with v1_12_6 but is not part of the 1.12.5 wire format.
func Stop(appName, source string, all bool) (method, path string, body any) {
	return http.MethodPost, "/apps/stop", map[string]any{
		"appName": appName,
		"all":     all,
	}
}
