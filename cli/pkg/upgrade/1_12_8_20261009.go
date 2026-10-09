package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_8_20261009 moves the overlay gateway parent from the legacy
// bridge to the wired NIC and completes the related CNI migration.
type upgrader_1_12_8_20261009 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8_20261009) Version() *semver.Version {
	return semver.MustParse("1.12.8-20261009")
}

func (u upgrader_1_12_8_20261009) UpgradeSystemComponents() []task.Interface {
	return append(u.upgraderBase.UpgradeSystemComponents(), fixedMACTasks()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20261009{})
}
