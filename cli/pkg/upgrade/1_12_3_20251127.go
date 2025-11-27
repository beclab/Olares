package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_3_20251127 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_3_20251127) Version() *semver.Version {
	return semver.MustParse("1.12.3-20251127")
}

func (u upgrader_1_12_3_20251127) NeedRestart() bool {
	return true
}

// put GPU driver upgrade step at the very end right before updating the version
func (u upgrader_1_12_3_20251127) UpdateOlaresVersion() []task.Interface {
	var tasks []task.Interface
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "UpgradeGPUDriver",
			Action: new(upgradeGPUDriverIfNeeded),
		},
	)
	tasks = append(tasks, u.upgraderBase.UpdateOlaresVersion()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "RebootIfNeeded",
			Action: new(rebootIfNeeded),
		},
	)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_3_20251127{})
}
