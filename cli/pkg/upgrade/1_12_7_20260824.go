package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260824 regenerates the kubelet configuration so an existing
// node picks up the larger system-reserved memory. The reserve is computed from
// the host's RAM and swap at generation time, so regenerating the files is all
// it takes.
//
// getUpgraderByVersion matches the target version exactly, so this only runs for
// a release cut as 1.12.7-20260824. If the release carrying this change ends up
// stamped with another date, rename this file and the version below to match, or
// existing clusters silently keep the old reserve.
type upgrader_1_12_7_20260824 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260824) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260824")
}

// PostUpgradeNode is where the regeneration goes, and both halves of that
// matter.
//
// Node, because the reserve is computed from the RAM and swap of the machine
// the task runs on, and the service file it rewrites is that machine's. Run
// only on the control node it would leave every other node on the old reserve,
// which is the thing this change exists to correct.
//
// Post, because it restarts k3s, and every stage before this one runs tasks
// that exec into pods. Doing it up front made those race the kubelet's re-sync
// and fail with "pod does not exist".
//
// What makes the restart safe for whatever runs next is not where this sits in
// the flow but that the regeneration ends in WaitForKubeAPIServerUp. The guard
// travels with the tasks, so the stage after this one — the version flip,
// which needs the apiserver — cannot begin until the apiserver it needs has
// answered. Anything that moves this work must keep that wait attached to it.
func (u upgrader_1_12_7_20260824) PostUpgradeNode() []task.Interface {
	return append(regenerateKubeFilesOnNode(), u.upgraderBase.PostUpgradeNode()...)
}

// PostUpgrade carries the kubeadm half, which only the control node may run.
// See regenerateKubeFilesOnControlNode.
func (u upgrader_1_12_7_20260824) PostUpgrade() []task.Interface {
	return append(regenerateKubeFilesOnControlNode(), u.upgraderBase.PostUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260824{})
}
