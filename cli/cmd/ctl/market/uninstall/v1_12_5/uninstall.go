// Package v1_12_5 builds the `market uninstall` request as it existed on
// Olares 1.12.5 (pre TermiPass PR #1162): DELETE /apps/{name}; the app name
// rides the path and the body carries only {sync, all, deleteData}. It is a
// pure request builder — it returns (method, path, body) and never touches the
// network, so it has no dependency on the market package (no import cycle).
package v1_12_5

import "net/http"

// Uninstall returns the 1.12.5 uninstall request. source / version are
// accepted for signature parity with v1_12_6 but are not part of the 1.12.5
// wire format.
func Uninstall(appName, source, version string, all, deleteData bool) (method, path string, body any) {
	return http.MethodDelete, "/apps/" + appName, map[string]any{
		"sync":       true,
		"all":        all,
		"deleteData": deleteData,
	}
}
