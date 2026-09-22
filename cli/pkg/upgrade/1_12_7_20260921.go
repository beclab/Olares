package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260921 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260921) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260921")
}

func (u upgrader_1_12_7_20260921) PrepareForUpgrade() []task.Interface {
	tasks := refreshBackupETCDScript()
	return append(tasks, u.upgraderBase.PrepareForUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260921{})
}
