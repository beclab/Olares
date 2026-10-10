package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

// upgrader_1_12_8_20261010 moves the overlay gateway parent from the legacy
// bridge to the wired NIC and completes the related CNI migration.
//
// getUpgraderByVersion matches the target version exactly, so this daily
// stamp must be the first release that carries the migration.
type upgrader_1_12_8_20261010 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8_20261010) Version() *semver.Version {
	return semver.MustParse("1.12.8-20261010")
}

func (u upgrader_1_12_8_20261010) UpgradeSystemComponents() []task.Interface {
	tasks := []task.Interface{terminus.GenericDevicePluginTask()}
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	tasks = append(tasks, fixedMACTasks()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20261010{})
}
