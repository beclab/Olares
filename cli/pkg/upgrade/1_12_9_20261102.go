package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// This daily must match the first release carrying the xattr migration;
// getUpgraderByVersion selects an exact target version.
type upgrader_1_12_9_20261102 struct{ breakingUpgraderBase }

func (u upgrader_1_12_9_20261102) Version() *semver.Version {
	return semver.MustParse("1.12.9-20261102")
}

func (u upgrader_1_12_9_20261102) PreUpgradeNode() []task.Interface {
	return append(juiceFSXattrTasks(), u.upgraderBase.PreUpgradeNode()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_9_20261102{})
}
