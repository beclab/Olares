package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260902 installs/upgrades the AMD device-metrics-exporter
// chart on nodes with a discrete AMD GPU (amd-gpu).
type upgrader_1_12_7_20260902 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260902) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260902")
}

func (u upgrader_1_12_7_20260902) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeAmdDeviceMetricsExporter()...)
	tasks = append(tasks, upgradeAmdDevicePlugin()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260902{})
}
