package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_3_20251203 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_3_20251203) Version() *semver.Version {
	return semver.MustParse("1.12.3-20251203")
}

func (u upgrader_1_12_3_20251203) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeNodeExporter()...)
	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_3_20251203{})
}
