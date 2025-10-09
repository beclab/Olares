package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_0_20250723 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_0_20250723) Version() *semver.Version {
	return semver.MustParse("1.12.0-20250723")
}

func (u upgrader_1_12_0_20250723) PrepareForUpgrade() []task.Interface {
	var preTasks []task.Interface
	preTasks = append(preTasks, upgradeContainerd()...)
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_0_20250723{})
}
