// Package restart builds the `market restart` request. Unlike
// stop/resume/uninstall (whose wire format diverges across Olares versions and
// therefore lives in per-version builder sub-packages), the restart endpoint is
// version-agnostic in the SPA: apps/packages/app/src/api/market/private/operations.ts
// declares Resume2RestartRequest with the note "no version, shared by resume and
// restart" and POSTs {app_name, source, computeBinding?} to /apps/restart. We
// mirror that single body here. computeBinding is a 1.12.6+ concept and is only
// written when non-nil (identical to resume's 1.12.6 body), so a restart without
// a GPU binding keeps the minimal {app_name, source} payload.
package restart

import "net/http"

// Build returns the restart request. computeBinding is passed as `any` so this
// package stays free of a dependency on the parent market package (no import
// cycle); pass nil when there is no binding to send.
func Build(appName, source string, computeBinding any) (method, path string, body any) {
	b := map[string]any{
		"app_name": appName,
		"source":   source,
	}
	if computeBinding != nil {
		b["computeBinding"] = computeBinding
	}
	return http.MethodPost, "/apps/restart", b
}
