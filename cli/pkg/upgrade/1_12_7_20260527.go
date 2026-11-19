package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type upgrader_1_12_7_20260527 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260527) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260527")
}

// PreUpgradeNode makes the shared app directories on each machine, so an app
// scheduled to any node finds them there.
func (u upgrader_1_12_7_20260527) PreUpgradeNode() []task.Interface {
	return append(createAppCommonDir(), u.upgraderBase.PreUpgradeNode()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260527{})
}
