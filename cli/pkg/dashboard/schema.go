package dashboard

import (
	"embed"
	"io/fs"
	"path"
	"sort"
)

//go:embed schemas/*.json
var SchemaFS embed.FS

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
	KindOverviewPhysical       = "dashboard.overview.physical" // 9 cluster metric rows
	KindOverviewUser           = "dashboard.overview.user"     // user CPU / memory quota
	KindOverviewRanking        = "dashboard.overview.ranking"  // workload-grain rank
	KindOverviewCPU            = "dashboard.overview.cpu"      // per-node CPU details
	KindOverviewMemory         = "dashboard.overview.memory"   // per-node memory (physical|swap)
	KindOverviewDisk           = "dashboard.overview.disk"     // sections (main + per-disk partitions)
	KindOverviewDiskMain       = "dashboard.overview.disk.main"
	KindOverviewDiskPart       = "dashboard.overview.disk.partitions"
	KindOverviewPods           = "dashboard.overview.pods"    // per-node running-pod counters
	KindOverviewNetwork        = "dashboard.overview.network" // per-iface system-ifs
	KindOverviewFan            = "dashboard.overview.fan"     // sections (live + curve)
	KindOverviewFanLive        = "dashboard.overview.fan.live"
	KindOverviewFanCurve       = "dashboard.overview.fan.curve"
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
		KindSchemaIndex,
	}
}

// SchemaEntry is one row of the schema index — exported for both tests and
// the cmd-side schema introspection command.
type SchemaEntry struct {
	Path string // user-facing command path, e.g. "overview cpu"
	Kind string // dashboard.* constant
	File string // filename inside the embedded schemas/ FS
}

// LoadSchemaIndex reads the embedded schemas/ directory + augments each
// entry with the user-facing command path. Order: the static table below
// is the source of truth for path↔kind mapping; the FS walk only verifies
// every static entry has a matching file (a missing file returns the entry
// with File="" so `dashboard schema` still surfaces it).
func LoadSchemaIndex() []SchemaEntry {
	static := []SchemaEntry{
		{"dashboard overview", KindOverview, "overview.json"},
		{"dashboard overview physical", KindOverviewPhysical, "overview-physical.json"},
		{"dashboard overview user", KindOverviewUser, "overview-user.json"},
		{"dashboard overview ranking", KindOverviewRanking, "overview-ranking.json"},
		{"dashboard overview cpu", KindOverviewCPU, "overview-cpu.json"},
		{"dashboard overview memory", KindOverviewMemory, "overview-memory.json"},
		{"dashboard overview disk", KindOverviewDisk, "overview-disk.json"},
		{"dashboard overview disk main", KindOverviewDiskMain, "overview-disk-main.json"},
		{"dashboard overview disk partitions", KindOverviewDiskPart, "overview-disk-partitions.json"},
		{"dashboard overview pods", KindOverviewPods, "overview-pods.json"},
		{"dashboard overview network", KindOverviewNetwork, "overview-network.json"},
		{"dashboard overview fan", KindOverviewFan, "overview-fan.json"},
		{"dashboard overview fan live", KindOverviewFanLive, "overview-fan-live.json"},
		{"dashboard overview fan curve", KindOverviewFanCurve, "overview-fan-curve.json"},
		{"dashboard overview gpu list", KindOverviewGPUList, "overview-gpu-list.json"},
		{"dashboard overview gpu tasks", KindOverviewGPUTasks, "overview-gpu-tasks.json"},
		{"dashboard overview gpu get", KindOverviewGPUDetail, "overview-gpu-detail.json"},
		{"dashboard overview gpu task", KindOverviewGPUTaskDet, "overview-gpu-task-detail.json"},
		{"dashboard overview gpu detail", KindOverviewGPUDetailFull, "overview-gpu-detail-full.json"},
		{"dashboard overview gpu task-detail", KindOverviewGPUTaskDetFull, "overview-gpu-task-detail-full.json"},
		{"dashboard applications", KindApplicationsList, "applications.json"},
	}
	have := embeddedSchemaFiles()
	out := make([]SchemaEntry, 0, len(static))
	for _, e := range static {
		if !have[e.File] {
			e.File = ""
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func embeddedSchemaFiles() map[string]bool {
	out := map[string]bool{}
	_ = fs.WalkDir(SchemaFS, "schemas", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		out[path.Base(p)] = true
		return nil
	})
	return out
}

// ReadSchemaFile returns the raw bytes of the named schema file from the
// embedded FS. Used by the cmd-side `dashboard schema <path>` command to
// pretty-print one document. Returned errors come straight from fs — the
// caller wraps them with the user-friendly file path.
func ReadSchemaFile(file string) ([]byte, error) {
	return SchemaFS.ReadFile(path.Join("schemas", file))
}
