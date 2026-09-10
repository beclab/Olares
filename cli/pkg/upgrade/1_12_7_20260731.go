package upgrade

import (
	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260731 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260731) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260731")
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260731{})
}
