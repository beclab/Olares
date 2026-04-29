// Package disk hosts the business logic for the
// `olares-cli dashboard overview disk` subtree (root + main +
// partitions). The cmd-side area
// (cli/cmd/ctl/dashboard/overview/disk/) is a thin shell — it owns
// cobra wiring, persistent-flag binding, and the area-private
// *Client factory; this package owns every fetcher / aggregator /
// envelope-shape / table-render body so it's independently testable
// against an httptest.Server without dragging cobra into the
// fixture.
//
// Layout choice: the directory tree mirrors the command tree
// (overview/disk/{root,main,partitions}.go) so cmd → pkg navigation
// stays 1:1 across the two trees. helpers.go below is a small set of
// package-private aliases for the frequently-called formatting /
// sampling primitives in the top-level pkgdashboard. Without them
// every call site reads as `pkgdashboard.SampleFloat(...)`; the
// alias keeps call shape close to the pre-refactor cmd-side bodies
// while the source of truth stays in pkgdashboard.
package disk

import pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"

// Frequently-called formatting / sampling primitives. Vars (not
// type aliases) so the binding is zero-cost at runtime; flips
// happen here, not at every call site.
var (
	sampleFloat       = pkgdashboard.SampleFloat
	formatFloat       = pkgdashboard.FormatFloat
	percentString     = pkgdashboard.PercentString
	safeRatio         = pkgdashboard.SafeRatio
	lastSampleFromRow = pkgdashboard.LastSampleFromRow
)
