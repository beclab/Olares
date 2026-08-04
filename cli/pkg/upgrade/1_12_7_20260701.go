package upgrade

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/bootstrap/patch"
	"github.com/beclab/Olares/cli/pkg/core/task"

	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260701 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260701) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260701")
}

func (u upgrader_1_12_7_20260701) UpgradeSystemComponents() []task.Interface {
	tasks := append([]task.Interface{
		&task.LocalTask{
			Name:   "PatchNfsScript",
			Action: new(patch.PatchNfsScriptTask),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	})
	tasks = append(u.upgraderBase.UpgradeSystemComponents(), tasks...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260701{})
}
