package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_1_20250827 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_1_20250827) Version() *semver.Version {
	return semver.MustParse("1.12.1-20250827")
}

func (u upgrader_1_12_1_20250827) PrepareForUpgrade() []task.Interface {
	var preTasks []task.Interface
	preTasks = append(preTasks, upgradeKSCore()...)
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}
