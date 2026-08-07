package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260803 is crossed on upgrade through EDGE deenvy releases.
// UpdateOlaresVersion best-effort recreates business oes pods after l4 Ready,
// then writes the version; recreate never blocks the upgrade.
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

func (u upgrader_1_12_7_20260803) UpdateOlaresVersion() []task.Interface {
	tasks := deenvyEdgeUpgradeTasks()
	return append(tasks, u.upgraderBase.UpdateOlaresVersion()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260803{})
}
