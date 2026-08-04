package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260709 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260709) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260709")
}

func (u upgrader_1_12_7_20260709) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeNodeExporter()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)

	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260709{})
}
