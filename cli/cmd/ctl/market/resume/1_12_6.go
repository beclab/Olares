package resume

import "net/http"

// build1_12_6 is the Olares 1.12.6+ resume body (TermiPass PR #1162):
// POST /apps/resume with {app_name, source}.
func build1_12_6(appName, source string) (method, path string, body any) {
	return http.MethodPost, "/apps/resume", map[string]any{
		"app_name": appName,
		"source":   source,
	}
}
