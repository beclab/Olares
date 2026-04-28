// Package dashboard implements the `olares-cli dashboard` subtree — the
// AI-agent-oriented mirror of the dashboard SPA's overview / applications
// pages. Every leaf command makes the same HTTP calls the SPA does (against
// the per-user `dashboard.<terminus>` BFF, ks-apiserver-proxy) and renders
// either a tabular human view (`-o table`, default) or a strict JSON envelope
// for agent / scripting use (`-o json`).
//
// File layout:
//
//   - root.go            — command tree assembly + `dashboard schema` cmd.
//   - client.go          — *Client: HTTP wrappers, EnsureUser, RequireAdmin.
//   - output.go          — Envelope / Section / Item shapes + table renderer.
//   - params.go          — shared flag parsing + cross-flag validation.
//   - schema.go          — Kind constants + AllKinds() introspection.
//   - watch.go           — generic Runner (one-shot or --watch ticker).
//   - helpers.go         — shared fetch helpers (cluster monitoring,
//                          workloads, user metric, system-ifs, system-fan,
//                          graphics list / task list / details).
//   - overview.go        — overview tree assembly + default sections envelope.
//   - overview_*.go      — one leaf command per file.
//   - applications.go    — applications tree (default = workloads grid)
//                          + applications pods <namespace>.
//
// Plus child packages:
//
//   - format/   — 1:1 port of @bytetrade/core monitoring.ts +
//                 dashboard utils (cpu/disk/memory/...).
//   - schemas/  — embedded JSON Schema (draft-07) documents, served via
//                 `dashboard schema <command-path>`.
package dashboard

// Kind names every distinct payload shape the CLI can emit. Both `meta.kind`
// in the JSON envelope and the `dashboard schema` introspection table
// reference these constants — adding a new command means adding a Kind here
// AND a JSON Schema under schemas/.
//
// Naming convention: `dashboard.<area>.<verb>` (lowercase, dot-separated).
// The leading `dashboard.` namespace prevents collisions if other CLI
// subtrees ever adopt the same envelope.
const (
	// `dashboard overview` (default action) — sections envelope with three
	// keys: physical / user / ranking. Each section is itself a leaf
	// envelope (physical / user / ranking kinds below).
	KindOverview = "dashboard.overview"

	// Overview leaf commands — one per detail page in the SPA.
	KindOverviewPhysical    = "dashboard.overview.physical"     // 9 cluster metric rows
	KindOverviewUser        = "dashboard.overview.user"         // user CPU / memory quota
	KindOverviewRanking     = "dashboard.overview.ranking"      // workload-grain rank
	KindOverviewCPU         = "dashboard.overview.cpu"          // per-node CPU details
	KindOverviewMemory      = "dashboard.overview.memory"       // per-node memory (physical|swap)
	KindOverviewDisk        = "dashboard.overview.disk"         // sections (main + per-disk partitions)
	KindOverviewDiskMain    = "dashboard.overview.disk.main"    // per-disk main table
	KindOverviewDiskPart    = "dashboard.overview.disk.partitions"
	KindOverviewPods        = "dashboard.overview.pods"         // per-node running-pod counters
	KindOverviewNetwork     = "dashboard.overview.network"      // per-iface system-ifs
	KindOverviewFan         = "dashboard.overview.fan"          // sections (live + curve)
	KindOverviewFanLive     = "dashboard.overview.fan.live"
	KindOverviewFanCurve    = "dashboard.overview.fan.curve"
	KindOverviewGPUList        = "dashboard.overview.gpu.list"
	KindOverviewGPUTasks       = "dashboard.overview.gpu.tasks"
	KindOverviewGPUDetail      = "dashboard.overview.gpu.detail"
	KindOverviewGPUTaskDet     = "dashboard.overview.gpu.task.detail"
	KindOverviewGPUDetailFull  = "dashboard.overview.gpu.detail.full"
	KindOverviewGPUTaskDetFull = "dashboard.overview.gpu.task.detail.full"
	KindOverviewGPUGauges      = "dashboard.overview.gpu.gauges"
	KindOverviewGPUTrends      = "dashboard.overview.gpu.trends"

	// Applications tree.
	KindApplicationsList = "dashboard.applications.list"
	KindApplicationsPods = "dashboard.applications.pods"

	// Schema introspection.
	KindSchemaIndex = "dashboard.schema.index"
)

// AllKinds returns every Kind known to this CLI. Used by the schema-completeness
// unit test to ensure new commands don't forget to register their JSON Schema.
func AllKinds() []string {
	return []string{
		KindOverview,
		KindOverviewPhysical,
		KindOverviewUser,
		KindOverviewRanking,
		KindOverviewCPU,
		KindOverviewMemory,
		KindOverviewDisk,
		KindOverviewDiskMain,
		KindOverviewDiskPart,
		KindOverviewPods,
		KindOverviewNetwork,
		KindOverviewFan,
		KindOverviewFanLive,
		KindOverviewFanCurve,
		KindOverviewGPUList,
		KindOverviewGPUTasks,
		KindOverviewGPUDetail,
		KindOverviewGPUTaskDet,
		KindOverviewGPUDetailFull,
		KindOverviewGPUTaskDetFull,
		KindOverviewGPUGauges,
		KindOverviewGPUTrends,
		KindApplicationsList,
		KindApplicationsPods,
		KindSchemaIndex,
	}
}
