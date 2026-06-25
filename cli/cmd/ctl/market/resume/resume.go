// Package resume builds the `market resume` request for the detected Olares
// backend version. Build dispatches to the version-specific body builder; the
// per-version wire formats live in 1_12_5.go / 1_12_6.go. These are pure
// request builders — they return (method, path, body) and never touch the
// network, so the package has no dependency on the parent market package (no
// import cycle). Retiring a version later = delete its 1_12_x.go file.
package resume

// Build returns the resume request, choosing the body shape by backend
// version. computeBinding is a 1.12.6+ concept (the selected devices); it is
// passed as `any` so this package stays free of a dependency on the parent
// market package (no import cycle) and is only ever written into the 1.12.6
// body. Pass nil when there is no binding to send.
func Build(atLeast126 bool, appName, source string, computeBinding any) (method, path string, body any) {
	if atLeast126 {
		return build1_12_6(appName, source, computeBinding)
	}
	return build1_12_5(appName, source)
}
