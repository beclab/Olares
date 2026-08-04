package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260803 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260803) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260803")
}

func (u upgrader_1_12_7_20260803) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeKubernetesPrometheusRule()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)

	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260803{})
}
