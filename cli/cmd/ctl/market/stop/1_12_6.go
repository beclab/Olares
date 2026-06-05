package stop

import "net/http"

// build1_12_6 is the Olares 1.12.6+ stop body (TermiPass PR #1162):
// POST /apps/stop with {app_name, source, all}.
func build1_12_6(appName, source string, all bool) (method, path string, body any) {
	return http.MethodPost, "/apps/stop", map[string]any{
		"app_name": appName,
		"source":   source,
		"all":      all,
	}
}
