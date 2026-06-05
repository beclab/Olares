// Package resume builds the `market resume` request for the detected Olares
// backend version. Build dispatches to the version-specific body builder; the
// per-version wire formats live in 1_12_5.go / 1_12_6.go. These are pure
// request builders — they return (method, path, body) and never touch the
// network, so the package has no dependency on the parent market package (no
// import cycle). Retiring a version later = delete its 1_12_x.go file.
package resume

// Build returns the resume request, choosing the body shape by backend version.
func Build(atLeast126 bool, appName, source string) (method, path string, body any) {
	if atLeast126 {
		return build1_12_6(appName, source)
	}
	return build1_12_5(appName, source)
}
