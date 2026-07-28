package cancel

import "net/http"

// build1_12_6 is the Olares 1.12.6+ cancel request: DELETE /apps/{name}/install
// with a body that now carries app_name + source (the 1.12.5 body only sent
// {sync}, which the 1.12.6 backend rejects with "Missing required fields:
// app_name is required"). version is the installed version recorded on the
// per-user state row and is included when known, mirroring the SPA's
// cancelInstalling() which always sends {app_name, source, version}.
func build1_12_6(appName, source, version string) (method, path string, body any) {
	b := map[string]any{
		"app_name": appName,
		"source":   source,
		"sync":     true,
	}
	if version != "" {
		b["version"] = version
	}
	return http.MethodDelete, "/apps/" + appName + "/install", b
}
