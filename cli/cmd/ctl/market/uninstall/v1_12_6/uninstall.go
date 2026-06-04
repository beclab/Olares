// Package v1_12_6 builds the `market uninstall` request for Olares 1.12.6+
// (TermiPass PR #1162): DELETE /apps/{name} with a body that now carries
// app_name + source (UninstallRequest extends BaseOperationRequest) alongside
// sync / all / deleteData. It is a pure request builder — it returns
// (method, path, body) and never touches the network, so it has no dependency
// on the market package (no import cycle).
package v1_12_6

import "net/http"

// Uninstall returns the 1.12.6 uninstall request. version is included only
// when the caller supplies one.
func Uninstall(appName, source, version string, all, deleteData bool) (method, path string, body any) {
	b := map[string]any{
		"app_name":   appName,
		"source":     source,
		"sync":       true,
		"all":        all,
		"deleteData": deleteData,
	}
	if version != "" {
		b["version"] = version
	}
	return http.MethodDelete, "/apps/" + appName, b
}
