package stop

import "net/http"

// build1_12_5 is the Olares 1.12.5 stop body (pre TermiPass PR #1162):
// POST /apps/stop with {appName, all}. source is unused on this version.
func build1_12_5(appName, source string, all bool) (method, path string, body any) {
	return http.MethodPost, "/apps/stop", map[string]any{
		"appName": appName,
		"all":     all,
	}
}
