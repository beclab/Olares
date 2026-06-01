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
	tasks = append(tasks, upgradeMultus()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260525{})
}
