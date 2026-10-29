package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// upgrader_1_12_8_20261019 rolls Prometheus to the image in the embedded
// manifest, which can start after a zero-filled head chunk file.
type upgrader_1_12_8_20261019 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_8_20261019) Version() *semver.Version {
	return semver.MustParse("1.12.8-20261019")
}

func (u upgrader_1_12_8_20261019) UpgradeSystemComponents() []task.Interface {
	return append(upgradePrometheus(), u.upgraderBase.UpgradeSystemComponents()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_8_20261019{})
}
