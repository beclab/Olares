package upgrade

import (
	"github.com/Masterminds/semver/v3"
)

type upgrader_1_12_0_20250705 struct {
	upgraderBase
}

func (u upgrader_1_12_0_20250705) Version() *semver.Version {
	return semver.MustParse("1.12.0-20250705")
}
