package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260527 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260527) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260527")
}

func (u upgrader_1_12_7_20260527) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, createAppCommonDir()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260527{})
}
