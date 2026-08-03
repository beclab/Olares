package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_7_20260803 delivers whole-system oes-free steady state
// (PLAN-SYS-DEENVY-OTA-01 / IWO-SYS-DEENVY-01). UpdateOlaresVersion runs only
// after DeenvyCommitGate (accept passed + SteadyStateGate Ready).
type upgrader_1_12_7_20260803 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260803) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260803")
}

func (u upgrader_1_12_7_20260803) PrepareForUpgrade() []task.Interface {
	ver := u.Version().Original()
	tasks := deenvyUpgradeTasks(ver)[:2] // preflight + preload before base prepare
	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func (u upgrader_1_12_7_20260803) UpgradeSystemComponents() []task.Interface {
	ver := u.Version().Original()
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeKubernetesPrometheusRule()...)
	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	// Wait for EG/Linkerd/system-server proxy after platform charts land.
	tasks = append(tasks, deenvyUpgradeTasks(ver)[2:3]...) // WaitDeps only
	return tasks
}

func (u upgrader_1_12_7_20260803) UpgradeUserComponents() []task.Interface {
	ver := u.Version().Original()
	tasks := u.upgraderBase.UpgradeUserComponents()
	// After user charts: cutover RouteMode + enable oes-free gate.
	tasks = append(tasks, deenvyUpgradeTasks(ver)[3:5]...) // Cutover + Rollout
	return tasks
}

func (u upgrader_1_12_7_20260803) UpdateOlaresVersion() []task.Interface {
	ver := u.Version().Original()
	// Accept inventory then commit gate, then write version (PLAN commit point).
	tasks := deenvyUpgradeTasks(ver)[5:] // Accept + CommitGate
	tasks = append(tasks, u.upgraderBase.UpdateOlaresVersion()...)
	return tasks
}

func (u upgrader_1_12_7_20260803) PostUpgrade() []task.Interface {
	ver := u.Version().Original()
	tasks := u.upgraderBase.PostUpgrade()
	tasks = append(tasks, deenvyPostTasks(ver)...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260803{})
}
