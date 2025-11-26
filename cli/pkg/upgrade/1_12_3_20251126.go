package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_3_20251126 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_3_20251126) Version() *semver.Version {
	return semver.MustParse("1.12.3-20251126")
}

func (u upgrader_1_12_3_20251126) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeKsConfig()...)
	tasks = append(tasks, upgradePrometheusServiceMonitorKubelet()...)
	tasks = append(tasks, upgradeKSCore()...)
	tasks = append(tasks, regenerateKubeFiles()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_3_20251126{})
}
