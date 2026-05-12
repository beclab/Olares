package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_6_20260512 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_6_20260512) Version() *semver.Version {
	return semver.MustParse("1.12.6-20260512")
}

func (u upgrader_1_12_6_20260512) NeedRestart() bool {
	return true
}

func (u upgrader_1_12_6_20260512) PrepareForUpgrade() []task.Interface {
	return u.upgraderBase.PrepareForUpgrade()
}

func (u upgrader_1_12_6_20260512) UpdateOlaresVersion() []task.Interface {
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
	registerDailyUpgrader(upgrader_1_12_6_20260512{})
}
