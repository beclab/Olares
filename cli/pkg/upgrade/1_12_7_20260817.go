package upgrade

import (
	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260817 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260817) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260817")
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260817{})
}
