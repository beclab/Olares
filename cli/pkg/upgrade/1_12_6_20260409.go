package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_6_20260409 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_6_20260409) Version() *semver.Version {
	return semver.MustParse("1.12.6-20260409")
}

func (u upgrader_1_12_6_20260409) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeKsConfig()...)
	tasks = append(tasks, upgradeKubernetesPrometheusRule()...)
	tasks = append(tasks, upgradeNodeExporterServiceMonitor()...)
	tasks = append(tasks, upgradeNodeExporter()...)
	tasks = append(tasks, upgradeKSCore()...)

	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_6_20260409{})
}
