package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader is what one version contributes to each stage of the upgrade flow.
//
// Every method is bound to exactly one stage, and the stage decides where its
// tasks run — see UpgradeFlow. A task therefore never says where it executes;
// choosing a method is choosing the point in the flow, and the fanout comes
// with it. Adding work to an upgrade means picking the method whose stage
// describes when the work has to happen, not appending to a list.
//
// The methods are grouped below by the stage they feed. Ordering inside one
// stage is the contributor's own business and is preserved as written.
type upgrader interface {
	// --- stage: replace-binaries (every node, one at a time) ---

	// ReplaceBinaries swaps olaresd and olares-cli on the machine.
	ReplaceBinaries() []task.Interface

	// --- stage: import-images (every node) ---

	// ImportImages loads the images this upgrade needs onto the machine.
	ImportImages() []task.Interface

	// --- stage: pre-upgrade-node (every node, one at a time) ---

	// PreUpgradeNode is work each machine does to itself before the cluster
	// is upgraded: its container runtime, its systemd units, its /etc.
	PreUpgradeNode() []task.Interface

	// --- stage: pre-upgrade-admin (control node) ---

	// PrepareForUpgrade is control-node work against the cluster that has to
	// happen before the upgrade: manifests, CRDs, cached facts.
	PrepareForUpgrade() []task.Interface

	// --- stage: upgrade-cluster (control node) ---

	ClearAppChartValues() []task.Interface
	ClearBFLChartValues() []task.Interface
	UpdateChartsInAppService() []task.Interface
	UpgradeSystemComponents() []task.Interface
	UpgradeUserComponents() []task.Interface

	// --- stage: post-upgrade-node (every node) ---

	// PostUpgradeNode is work each machine does to itself afterwards.
	PostUpgradeNode() []task.Interface

	// UpdateReleaseFile stamps the new version onto the machine. It is
	// per-node because /etc/olares/release and the .installed marker are what
	// that node's olaresd reports about itself.
	UpdateReleaseFile() []task.Interface

	// --- stage: post-upgrade-admin (control node) ---

	// UpdateOlaresVersion flips the version in the cluster.
	UpdateOlaresVersion() []task.Interface

	// PostUpgrade checks the cluster came back up.
	PostUpgrade() []task.Interface

	// --- stage: reboot-nodes (every node, one at a time) ---

	// RebootNodes restarts the machines whose upgrade requires it.
	RebootNodes() []task.Interface

	AddedBreakingChange() bool
	NeedRestart() bool
}

// stageTasks asks u for everything it contributes, stage by stage. It is the
// only place the mapping from method to stage is written down.
//
// The methods above are the hooks and the stages are what they feed: a stage
// is a point in the upgrade, and a hook is how one version gets its tasks into
// it. The two are not the same thing and only the stages are named on the
// wire — a node is told which stage to run, never which method produced it.
//
// It reads the methods through the interface rather than off a base struct,
// because that is what makes a version's override the answer: Go resolves an
// embedded struct's own call to its own method, so a base that asked itself
// would never see what the version replaced.
func stageTasks(u upgrader) map[string][]task.Interface {
	return map[string][]task.Interface{
		StageReplaceBinaries: u.ReplaceBinaries(),
		StageImportImages:    u.ImportImages(),
		StagePreUpgradeNode:  u.PreUpgradeNode(),
		StagePreUpgradeAdmin: u.PrepareForUpgrade(),
		StageUpgradeCluster: concat(
			u.ClearAppChartValues(),
			u.ClearBFLChartValues(),
			u.UpdateChartsInAppService(),
			u.UpgradeSystemComponents(),
			u.UpgradeUserComponents(),
		),
		StagePostUpgradeNode: concat(
			u.PostUpgradeNode(),
			u.UpdateReleaseFile(),
		),
		StagePostUpgradeAdmin: concat(
			u.UpdateOlaresVersion(),
			u.PostUpgrade(),
		),
		StageRebootNodes: u.RebootNodes(),
	}
}

// concat joins task groups, keeping the order they were given in. Ordering
// within one stage is the contributor's own business.
func concat(groups ...[]task.Interface) []task.Interface {
	var out []task.Interface
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

type breakingUpgrader interface {
	upgrader
	Version() *semver.Version
}
