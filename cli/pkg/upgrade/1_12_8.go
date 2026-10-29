package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/beclab/Olares/cli/version"
)

var version_1_12_8 = semver.MustParse("1.12.8")

type upgrader_1_12_8 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	if err != nil {
		return version_1_12_8
	}
	if samePatchLevelVersion(version_1_12_8, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_1_12_8
}

func (u upgrader_1_12_8) AddedBreakingChange() bool {
	if u.Version().Equal(version_1_12_8) {
		return true
	}
	return false
}

func (u upgrader_1_12_8) UpgradeSystemComponents() []task.Interface {
	tasks := []task.Interface{terminus.GenericDevicePluginTask()}
	tasks = append(tasks, upgradePrometheus()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	tasks = append(tasks, fixedMACTasks()...)
	tasks = append(tasks, migrateIntelDGPUToHAMITasks()...)
	return tasks
}

// PostUpgradeNode regenerates the kubelet configuration so every node drops the
// soft eviction thresholds; see upgrader_1_12_7_20260824 for why this is
// per-node and placed after the cluster upgrade.
func (u upgrader_1_12_8) PostUpgradeNode() []task.Interface {
	return append(regenerateKubeFilesOnNode(), u.upgraderBase.PostUpgradeNode()...)
}

// PostUpgrade carries the kubeadm half, which only the control node may run.
// See regenerateKubeFilesOnControlNode.
func (u upgrader_1_12_8) PostUpgrade() []task.Interface {
	return append(regenerateKubeFilesOnControlNode(), u.upgraderBase.PostUpgrade()...)
}

func init() {
	registerMainUpgrader(upgrader_1_12_8{})
}
