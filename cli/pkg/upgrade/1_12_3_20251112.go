package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_3_20251112 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_3_20251112) Version() *semver.Version {
	return semver.MustParse("1.12.3-20251112")
}

func (u upgrader_1_12_3_20251112) PrepareForUpgrade() []task.Interface {
	return append(regenerateKubeFiles(), u.upgraderBase.PrepareForUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_3_20251112{})
}
