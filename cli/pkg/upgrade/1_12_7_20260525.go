package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260525 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260525) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260525")
}

func (u upgrader_1_12_7_20260525) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeMultusCluster()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

// PreUpgradeNode installs the multus DHCP daemon on each machine.
func (u upgrader_1_12_7_20260525) PreUpgradeNode() []task.Interface {
	return append(upgradeMultusNode(), u.upgraderBase.PreUpgradeNode()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260525{})
}
