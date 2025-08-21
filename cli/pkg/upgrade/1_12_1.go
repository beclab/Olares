package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/version"
)

var version_12_1_1 = semver.MustParse("1.12.1")

type upgrader_1_12_1 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_1) Version() *semver.Version {
	cliVersion, err := semver.NewVersion(version.VERSION)
	// tolerate local dev version
	if err != nil {
		return version_12_1_1
	}
	if samePatchLevelVersion(version_12_1_1, cliVersion) && getReleaseLineOfVersion(cliVersion) == mainLine {
		return cliVersion
	}
	return version_12_1_1
}

func (u upgrader_1_12_1) AddedBreakingChange() bool {
	if u.Version().Equal(version_12_1_1) {
		// if this version introduced breaking change
		return true
	}
	if u.Version().Equal(semver.MustParse("1.12.1-alpha.2")) {
		// if this alpha version introduced more breaking change
		// halfway through release
		return true
	}
	return false
}
