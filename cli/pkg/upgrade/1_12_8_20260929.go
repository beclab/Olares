package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

// upgrader_1_12_8_20260929 installs or upgrades the platform device discovery
// DaemonSet through the same task used by a fresh installation.
type upgrader_1_12_8_20260929 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8_20260929) Version() *semver.Version {
	return semver.MustParse("1.12.8-20260929")
}

func (u upgrader_1_12_8_20260929) UpgradeSystemComponents() []task.Interface {
	tasks := []task.Interface{terminus.GenericDevicePluginTask()}
	return append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20260929{})
}
