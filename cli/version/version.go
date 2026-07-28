package version

var VERSION = "0.0.0-development"

// GitCommit and BuildTime are injected at build time via -ldflags.
// They default to "unknown" so a `go build`/`go run` without ldflags
// (or any consumer parsing only the first line of the version output)
// still behaves sensibly.
var (
	GitCommit = "unknown"
	BuildTime = "unknown"
)
