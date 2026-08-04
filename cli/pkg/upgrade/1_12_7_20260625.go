package upgrade

import (
	"github.com/beclab/Olares/cli/pkg/core/task"

	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260625 struct {
	upgrader_1_12_7_20260624
}

func (u upgrader_1_12_7_20260625) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260625")
}

func (u upgrader_1_12_7_20260625) UpgradeSystemComponents() []task.Interface {
	tasks := u.upgrader_1_12_7_20260624.UpgradeSystemComponents()
	// backfill the per-mode (multi-mode) node labels for Intel/AMD GPUs so
	// devices upgraded from before the per-mode labeling scheme advertise
	// their GPU mode to the scheduler.
	tasks = append(tasks, labelIntelAMDGPUNode()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260625{})
}
