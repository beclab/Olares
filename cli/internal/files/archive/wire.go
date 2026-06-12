// wire.go: shared path-shape helpers for the archive verbs.
//
// Sources and destinations on /api/archive/<node>/ travel as
// plain UTF-8 strings shaped like `/<fileType>/<extend>/<sub>`
// — the same wire form the cp / paste body uses. The cobra
// layer parses FrontendPath instances and stitches them through
// BuildWirePath; the package never imports the FrontendPath
// type so it stays free of cobra/cmdutil deps.
package archive

import (
	"strings"
)

// BuildWirePath assembles the LarePass-shaped wire path
// `/<fileType>/<extend>/<sub>`. Subpath is taken verbatim: a
// leading '/' is preserved (or added if missing) and a trailing
// '/' is preserved (the backend interprets it as a directory
// marker on some endpoints; archive treats it the same as
// non-trailing-slash for source paths but the consistency is
// useful for log lines / error messages).
//
// Examples:
//
//	BuildWirePath("drive", "Home", "/Documents/foo.pdf") →
//	    "/drive/Home/Documents/foo.pdf"
//	BuildWirePath("drive", "Home", "/Photos/")            →
//	    "/drive/Home/Photos/"
//	BuildWirePath("drive", "Home", "")                    →
//	    "/drive/Home/"
//
// Used by the cobra layer's per-source / per-destination
// builders.
func BuildWirePath(fileType, extend, subPath string) string {
	sub := subPath
	if sub == "" {
		sub = "/"
	}
	if !strings.HasPrefix(sub, "/") {
		sub = "/" + sub
	}
	return "/" + fileType + "/" + extend + sub
}
