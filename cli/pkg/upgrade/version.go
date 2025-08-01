package upgrade

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"
	"strings"
)

type releaseLine string

var (
	mainLine  = releaseLine("main")
	dailyLine = releaseLine("daily")

	dailyUpgraders = []breakingUpgrader{
		upgrader_1_12_0_20250702{},
		upgrader_1_12_0_20250723{},
		upgrader_1_12_0_20250730{},
	}
	mainUpgraders = []breakingUpgrader{}
)

func getReleaseLineOfVersion(v *semver.Version) releaseLine {
	preRelease := v.Prerelease()
	mainLinePrereleasePrefixes := []string{"alpha", "beta", "rc"}
	if preRelease == "" {
		return mainLine
	}
	for _, prefix := range mainLinePrereleasePrefixes {
		if strings.HasPrefix(preRelease, prefix) {
			return mainLine
		}
	}
	return dailyLine
}

func check(base *semver.Version, target *semver.Version) error {
	if base == nil {
		return fmt.Errorf("base version is nil")
	}
	baseReleaseLine := getReleaseLineOfVersion(base)

	cliVersion, err := utils.ParseOlaresVersionString(version.VERSION)
	if err != nil {
		return fmt.Errorf("invalid olares-cli version :\"%s\"", version.VERSION)
	}
	cliReleaseLine := getReleaseLineOfVersion(cliVersion)
	if baseReleaseLine != cliReleaseLine {
		return fmt.Errorf("incompatible base release line: %s and olares-cli release line: %s", baseReleaseLine, cliReleaseLine)
	}

	if target != nil {
		if !target.GreaterThan(base) {
			return fmt.Errorf("base version: %s, target version: %s, no need to upgrade", base, target)
		}

		targetReleaseLine := getReleaseLineOfVersion(target)
		if targetReleaseLine != baseReleaseLine {
			return fmt.Errorf("unable to upgrade to %s on %s release line from %s on %s release line", target, targetReleaseLine, base, baseReleaseLine)
		}
		switch targetReleaseLine {
		case mainLine:
			if !sameMajorLevelVersion(base, target) {
				return fmt.Errorf("upgrade on %s rlease line can only be performed across same major level version", baseReleaseLine)
			}
		case dailyLine:
			if !sameMinorLevelVersion(base, target) {
				return fmt.Errorf("upgrade on %s rlease line can only be performed across same patch version", baseReleaseLine)
			}
		}

		if target.GreaterThan(cliVersion) {
			return fmt.Errorf("target version: %s, cli version: %s, please upgrade olares-cli first!", target, cliVersion)
		}
	}

	if base.GreaterThan(cliVersion) {
		return fmt.Errorf("base version: %s, cli version: %s, please upgrade olares-cli first!", base, cliVersion)
	}

	return nil
}

func GetUpgradePathFor(base *semver.Version, target *semver.Version) ([]*semver.Version, error) {
	if err := check(base, target); err != nil {
		return nil, err
	}
	var path []*semver.Version
	var releaseLineUpgraders []breakingUpgrader
	var versionFilter func(v *semver.Version) bool
	line := getReleaseLineOfVersion(base)
	if target == nil {
		cliVersion, err := utils.ParseOlaresVersionString(version.VERSION)
		if err != nil {
			return path, fmt.Errorf("invalid olares-cli version :\"%s\"", version.VERSION)
		}
		if getReleaseLineOfVersion(cliVersion) == line && cliVersion.GreaterThan(base) {
			target = cliVersion
		}
	}
	switch line {
	case mainLine:
		releaseLineUpgraders = mainUpgraders
		versionFilter = func(v *semver.Version) bool {
			if !v.GreaterThan(base) {
				return false
			}
			if target != nil && !v.LessThan(target) {
				return false
			}
			return true
		}
	case dailyLine:
		releaseLineUpgraders = dailyUpgraders
		versionFilter = func(v *semver.Version) bool {
			if !v.GreaterThan(base) {
				return false
			}
			if target != nil && !v.LessThan(target) {
				return false
			}
			return true
		}
	}

	for _, u := range releaseLineUpgraders {
		v := u.Version()
		if versionFilter(v) {
			path = append(path, v)
		}
	}

	if target != nil {
		path = append(path, target)
	}

	return path, nil
}

func getUpgraderByVersion(target *semver.Version) upgrader {
	for _, upgraders := range [][]breakingUpgrader{
		dailyUpgraders,
		mainUpgraders,
	} {

		for _, u := range upgraders {
			if u.Version().Equal(target) {
				return u
			}
		}
	}
	return upgraderBase{}
}

func samePatchLevelVersion(a, b *semver.Version) bool {
	return a.Major() == b.Major() && a.Minor() == b.Minor() && a.Patch() == b.Patch()
}

func sameMinorLevelVersion(a, b *semver.Version) bool {
	return a.Major() == b.Major() && a.Minor() == b.Minor()
}

func sameMajorLevelVersion(a, b *semver.Version) bool {
	return a.Major() == b.Major()
}
