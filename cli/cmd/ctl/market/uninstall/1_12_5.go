package uninstall

import "net/http"

// build1_12_5 is the Olares 1.12.5 uninstall request (pre TermiPass PR #1162):
// DELETE /apps/{name} with body {sync, all, deleteData}. The app name rides the
// path; source / version are unused on this version.
func build1_12_5(appName, source, version string, all, deleteData bool) (method, path string, body any) {
	return http.MethodDelete, "/apps/" + appName, map[string]any{
		"sync":       true,
		"all":        all,
		"deleteData": deleteData,
	}
}
