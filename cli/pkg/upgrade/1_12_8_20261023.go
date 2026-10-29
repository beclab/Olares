package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_8_20261023 moves discrete Intel GPUs from the plain
// gpu.intel.com/xe device plugin resource to HAMi scheduling.
//
// getUpgraderByVersion matches the target version exactly, so this daily
// stamp must be the first release that carries the migration.
type upgrader_1_12_8_20261023 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8_20261023) Version() *semver.Version {
	return semver.MustParse("1.12.8-20261023")
}

func (u upgrader_1_12_8_20261023) UpgradeSystemComponents() []task.Interface {
	return append(u.upgraderBase.UpgradeSystemComponents(), migrateIntelDGPUToHAMITasks()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20261023{})
}
