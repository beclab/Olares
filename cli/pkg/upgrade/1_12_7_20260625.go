package upgrade

import (
	"github.com/beclab/Olares/cli/pkg/core/task"

	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260625 struct {
	upgrader_1_12_7_20260624
}

func (u upgrader_1_12_7_20260625) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260625")
}

// PreUpgradeNode backfills the per-mode (multi-mode) node labels for Intel/AMD
// GPUs, so devices upgraded from before the per-mode labeling scheme advertise
// their GPU mode to the scheduler. Each machine probes its own hardware and
// labels its own Node, so every node runs it.
//
// It runs here rather than after the cluster upgrade because that is where it
// used to be relative to the charts: the GPU plugin this labelling is for is
// upgraded during the cluster stage, and it should find the labels already
// there.
func (u upgrader_1_12_7_20260625) PreUpgradeNode() []task.Interface {
	return append(labelIntelAMDGPUNode(), u.upgrader_1_12_7_20260624.PreUpgradeNode()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260625{})
}
