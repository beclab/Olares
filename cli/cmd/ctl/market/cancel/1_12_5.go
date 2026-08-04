package cancel

import "net/http"

// build1_12_5 is the Olares 1.12.5 cancel request: DELETE /apps/{name}/install
// with body {sync}. The app name rides the path; source / version are unused on
// this version.
func build1_12_5(appName, source, version string) (method, path string, body any) {
	return http.MethodDelete, "/apps/" + appName + "/install", map[string]any{
		"sync": true,
	}
}
