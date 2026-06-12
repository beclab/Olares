package resume

import "net/http"

// build1_12_5 is the Olares 1.12.5 resume body (pre TermiPass PR #1162):
// POST /apps/resume with {appName}. source is unused on this version.
func build1_12_5(appName, source string) (method, path string, body any) {
	return http.MethodPost, "/apps/resume", map[string]any{
		"appName": appName,
	}
}
