package upgrade

import (
	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260603 struct {
	upgrader_1_12_7_20260527
}

func (u upgrader_1_12_7_20260603) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260603")
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260603{})
}
