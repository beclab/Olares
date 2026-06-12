package upgrade

import (
	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_7_20260605 struct {
	upgrader_1_12_7_20260513
}

func (u upgrader_1_12_7_20260605) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260605")
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260605{})
}
