package resume

import "net/http"

// build1_12_6 is the Olares 1.12.6+ resume body (TermiPass PR #1162):
// POST /apps/resume with {app_name, source, computeBinding?}. computeBinding
// is only added when non-nil so a resume without a GPU binding keeps the
// minimal {app_name, source} body.
func build1_12_6(appName, source string, computeBinding any) (method, path string, body any) {
	b := map[string]any{
		"app_name": appName,
		"source":   source,
	}
	if computeBinding != nil {
		b["computeBinding"] = computeBinding
	}
	return http.MethodPost, "/apps/resume", b
}
