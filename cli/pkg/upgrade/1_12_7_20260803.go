package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260803 delivers EDGE deenvy accept (l4 PEP + zero business oes).
// UpdateOlaresVersion: Accept recreates residual oes pods then gates; CommitGate
// writes SteadyState Ready. Webhook stops injecting once l4 Deployment is Ready
// (new install / post-upgrade) without waiting for Accept chicken-egg.
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
	ver := u.Version().Original()
	tasks := deenvyEdgeUpgradeTasks(ver)
	tasks = append(tasks, u.upgraderBase.UpdateOlaresVersion()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260803{})
}
