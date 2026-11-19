package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// The upgrader is selected by the exact target version, so this daily must
// register the first release that installs the NVIDIA base DEB from the CDN.
type upgrader_1_12_9_20261118 struct{ breakingUpgraderBase }

func (u upgrader_1_12_9_20261118) Version() *semver.Version {
	return semver.MustParse("1.12.9-20261118")
}

func (u upgrader_1_12_9_20261118) PreUpgradeNode() []task.Interface {
	return append(migrateNvidiaContainerToolkit(), u.upgraderBase.PreUpgradeNode()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_9_20261118{})
}
