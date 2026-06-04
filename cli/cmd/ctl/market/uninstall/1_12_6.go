package uninstall

import "net/http"

// build1_12_6 is the Olares 1.12.6+ uninstall request (TermiPass PR #1162):
// DELETE /apps/{name} with a body that now carries app_name + source alongside
// sync / all / deleteData. version is included only when supplied.
func build1_12_6(appName, source, version string, all, deleteData bool) (method, path string, body any) {
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
