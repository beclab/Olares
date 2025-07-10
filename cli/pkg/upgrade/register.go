package upgrade

import "github.com/Masterminds/semver/v3"

// todo: maybe add some code to sort and check uniqueness of upgraders to further ensure robustness
var breakingUpgraders = []breakingUpgrader{
	upgrader_1_12_0_20250702{},
	upgrader_1_12_0_20250705{},
}

func GetUpgradePathFor(base *semver.Version, target *semver.Version) []*semver.Version {
	var path []*semver.Version
	if base == nil || target == nil {
		return path
	}

	if target.LessThanEqual(base) {
		return path
	}

	for _, u := range breakingUpgraders {
		version := u.Version()
		if version.GreaterThanEqual(target) {
			break
		}

		if version.GreaterThan(base) {
			path = append(path, version)
		}
	}

	path = append(path, target)

	return path
}

func getUpgraderByVersion(target *semver.Version) upgrader {
	for _, u := range breakingUpgraders {
		if u.Version().Equal(target) {
			return u
		}
	}
	return upgraderBase{}
}
