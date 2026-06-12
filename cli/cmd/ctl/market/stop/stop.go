// Package stop builds the `market stop` request for the detected Olares
// backend version. Build dispatches to the version-specific body builder; the
// per-version wire formats live in 1_12_5.go / 1_12_6.go. These are pure
// request builders — they return (method, path, body) and never touch the
// network, so the package has no dependency on the parent market package (no
// import cycle). Retiring a version later = delete its 1_12_x.go file.
package stop

// Build returns the stop request, choosing the body shape by backend version.
func Build(atLeast126 bool, appName, source string, all bool) (method, path string, body any) {
	if atLeast126 {
		return build1_12_6(appName, source, all)
	}
	return build1_12_5(appName, source, all)
}
