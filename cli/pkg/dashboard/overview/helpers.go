// Package overview hosts the business logic for the
// `olares-cli dashboard overview` cobra subtree — physical / user /
// ranking / cpu / memory / pods / network leaves plus the default
// fan-out aggregator. The cmd-side area
// (cli/cmd/ctl/dashboard/overview/) is a thin shell: it owns cobra
// wiring, persistent-flag binding, and the area-private
// *Client factory; this package owns every fetch / aggregator /
// envelope-shape / table-render body so it's independently testable
// against an httptest.Server without dragging cobra into the fixture.
//
// Helpers below are package-private lowercase aliases for the
// frequently-called formatting / sampling primitives in the top-level
// pkgdashboard. Without them every call site reads as
// `pkgdashboard.SampleFloat(...)`; the alias keeps the call shape
// close to the pre-refactor cmd-side bodies (which used the same
// lowercase names) while the source of truth stays in pkgdashboard.
package overview

import pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"

// Frequently-called formatting / sampling primitives. Vars (not type
// aliases) so the binding is zero-cost at runtime and unaffected by
// pkgdashboard's exported-name evolution; flips happen here, not at
// every call site.
var (
	sampleFloat       = pkgdashboard.SampleFloat
	formatFloat       = pkgdashboard.FormatFloat
	percentString     = pkgdashboard.PercentString
	safeRatio         = pkgdashboard.SafeRatio
	formatRateAny     = pkgdashboard.FormatRateAny
	renderTemperature = pkgdashboard.RenderTemperature
	lastSampleFromRow = pkgdashboard.LastSampleFromRow
	toFloat           = pkgdashboard.ToFloat
)
