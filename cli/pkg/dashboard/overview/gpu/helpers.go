// Package gpu hosts the business logic for the
// `olares-cli dashboard overview gpu` subtree (root + list + tasks
// + get + task + detail + task-detail). The cmd-side area
// (cli/cmd/ctl/dashboard/overview/gpu/) is a thin shell — it owns
// cobra wiring, persistent-flag binding, and the area-private
// *Client factory; this package owns every fetcher / aggregator /
// envelope-shape / table-render body so it's independently testable
// against an httptest.Server without dragging cobra into the
// fixture.
//
// Layout choice: the directory tree mirrors the command tree
// (overview/gpu/{root,list,tasks,get,task,detail,task_detail,
// specs}.go). specs.go is the per-page query catalogue (gauge +
// trend specs, fan-out runners, display copies, window resolver) —
// kept in one file because the GPU detail and task detail pages
// share every primitive.
//
// helpers.go below is a small set of package-private aliases for
// the frequently-called formatting / sampling primitives in the
// top-level pkgdashboard. Without them every call site reads as
// `pkgdashboard.GPUVRAMHuman(...)`; the alias keeps call shape
// close to the pre-refactor cmd-side bodies while the source of
// truth stays in pkgdashboard.
package gpu

import pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"

// Frequently-called formatting / labelling / sampling primitives.
// Vars (not type aliases) so the binding is zero-cost at runtime;
// flips happen here, not at every call site.
var (
	formatFloat       = pkgdashboard.FormatFloat
	percentDirect     = pkgdashboard.PercentDirect
	renderTemperature = pkgdashboard.RenderTemperature
	toFloat           = pkgdashboard.ToFloat
	gpuVRAMHuman      = pkgdashboard.GPUVRAMHuman
	gpuModeLabel      = pkgdashboard.GPUModeLabel
	gpuHealthLabel    = pkgdashboard.GPUHealthLabel
	firstAnyInArray   = pkgdashboard.FirstAnyInArray
)
