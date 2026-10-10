package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_3_20251127 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_3_20251127) Version() *semver.Version {
	return semver.MustParse("1.12.3-20251127")
}

func (u upgrader_1_12_3_20251127) NeedRestart() bool {
	return true
}

// put GPU driver upgrade step at the very end right before updating the version
func (u upgrader_1_12_3_20251127) UpdateOlaresVersion() []task.Interface {
	return u.upgraderBase.UpdateOlaresVersion()
}

// PostUpgradeNode installs the target NVIDIA driver on the machine it runs
// on: every node has its own GPU and its own kernel to build against.
func (u upgrader_1_12_3_20251127) PostUpgradeNode() []task.Interface {
	return append(upgradeGPUDriver(), u.upgraderBase.PostUpgradeNode()...)
}

// RebootNodes restarts the machines whose driver changed. It runs after the
// version has been flipped, which is what the marker the driver upgrade
// writes depends on: while that marker exists olaresd reports the system as
// upgrading-and-rebooting rather than briefly reporting it as complete.
func (u upgrader_1_12_3_20251127) RebootNodes() []task.Interface {
	return append(rebootAfterGPUDriver(), u.upgraderBase.RebootNodes()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_3_20251127{})
}
