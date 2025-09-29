package upgrade

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"
	"sort"
	"strconv"
	"strings"
)

type VersionSpec struct {
	Version                   string                    `json:"version"`
	Major                     uint64                    `json:"major"`
	Minor                     uint64                    `json:"minor"`
	Patch                     uint64                    `json:"patch"`
	ReleaseType               string                    `json:"releaseType"`
	ReleaseNum                int                       `json:"releaseNum"`
	PreRelease                bool                      `json:"prerelease"`
	AddedBreakingChange       bool                      `json:"addedBreakingChange"`
	MinimumUpgradableVersions MinimumVersionConstraints `json:"minimumUpgradableVersions"`
}

type VersionConstraints map[string]*semver.Version
type MinimumVersionConstraints VersionConstraints

func (c MinimumVersionConstraints) SatisfiedBy(base *semver.Version) (bool, error) {
	if base == nil {
		return false, nil
	}
	var minVersion *semver.Version
	prerelease := base.Prerelease()
	if prerelease == "" {
		minVersion = c[releaseTypeStable]
	} else {
		prereleaseType, _, err := parsePrereleaseVersion(prerelease)
		if err != nil {
			return false, fmt.Errorf("invalid version '%s': %v", base, err)
		}
		minVersion = c[prereleaseType]
	}
	return minVersion != nil && minVersion.LessThanEqual(base), nil
}

type releaseLine string

var (
	mainLine  = releaseLine("main")
	dailyLine = releaseLine("daily")

	// the versions when current upgrade framework is introduced
	minUpgradableStableVersion = semver.MustParse("1.12.0")
	minUpgradableDailyVersion  = semver.MustParse("1.12.0-0")

	releaseTypeStable = "stable"
	releaseTypeDaily  = "daily"
	releaseTypeRC     = "rc"
	releaseTypeBeta   = "beta"
	releaseTypeAlpha  = "alpha"
	prereleaseSep     = "."

	dailyUpgraders = []breakingUpgrader{
		upgrader_1_12_0_20250702{},
		upgrader_1_12_0_20250723{},
		upgrader_1_12_0_20250730{},
		upgrader_1_12_1_20250826{},
		upgrader_1_12_2_20250929{},
	}
	mainUpgraders = []breakingUpgrader{
		upgrader_1_12_1{},
	}
)

func getReleaseLineOfVersion(v *semver.Version) releaseLine {
	preRelease := v.Prerelease()
	mainLinePrereleasePrefixes := []string{releaseTypeRC, releaseTypeBeta, releaseTypeAlpha}
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

func Check(base *semver.Version, target *semver.Version) error {
	if base == nil {
		return fmt.Errorf("base version is nil")
	}

	cliVersion, err := utils.ParseOlaresVersionString(version.VERSION)
	if err != nil {
		return fmt.Errorf("invalid olares-cli version :\"%s\"", version.VERSION)
	}

	if target != nil {
		if !target.GreaterThan(base) {
			return fmt.Errorf("base version: %s, target version: %s, no need to upgrade", base, target)
		}
		if !target.Equal(cliVersion) {
			return fmt.Errorf("target version: %s is not the same with cli version: %s", target, cliVersion)
		}
	}

	currentVersionSpec, err := CurrentVersionSpec()
	if err != nil {
		return fmt.Errorf("failed to get current version's upgrade spec: %v", err)
	}

	satisfied, err := currentVersionSpec.MinimumUpgradableVersions.SatisfiedBy(base)
	if err != nil {
		return err
	}

	if !satisfied {
		return fmt.Errorf("cannot upgrade to version '%s' from '%s': version constraints not satified: %v", target, base, currentVersionSpec.MinimumUpgradableVersions)
	}

	return nil
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

func parsePrereleaseVersion(prereleaseVersion string) (string, int, error) {
	if !strings.Contains(prereleaseVersion, prereleaseSep) {
		n, err := strconv.Atoi(prereleaseVersion)
		if err != nil {
			return "", 0, fmt.Errorf("invalid prereleaseVersion: %s", prereleaseVersion)
		}
		return releaseTypeDaily, n, nil
	}
	parts := strings.Split(prereleaseVersion, prereleaseSep)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid prereleaseVersion: %s", prereleaseVersion)
	}
	tStr, nStr := parts[0], parts[1]
	if tStr != releaseTypeRC && tStr != releaseTypeBeta && tStr != releaseTypeAlpha {
		return "", 0, fmt.Errorf("invalid prereleaseVersion: %s", prereleaseVersion)
	}
	n, err := strconv.Atoi(nStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid prereleaseVersion: %s", prereleaseVersion)
	}
	return tStr, n, nil
}

func formatPrereleaseVersion(releaseType string, releaseNum int) string {
	if releaseType == releaseTypeDaily {
		return strconv.Itoa(releaseNum)
	}
	return fmt.Sprintf("%s%s%s", releaseType, prereleaseSep, strconv.Itoa(releaseNum))
}

func CurrentVersionSpec() (spec *VersionSpec, err error) {
	v, err := semver.NewVersion(version.VERSION)
	if err != nil {
		return nil, fmt.Errorf("current version '%s' is invalid: %v", version.VERSION, err)
	}
	spec = &VersionSpec{}
	spec.Version, spec.Major, spec.Minor, spec.Patch = v.Original(), v.Major(), v.Minor(), v.Patch()
	if v.Prerelease() != "" {
		spec.PreRelease = true
		spec.ReleaseType, spec.ReleaseNum, err = parsePrereleaseVersion(v.Prerelease())
		if err != nil {
			return nil, err
		}
	} else {
		spec.ReleaseType = releaseTypeStable
	}
	u := getUpgraderByVersion(v)
	spec.AddedBreakingChange = u.AddedBreakingChange()
	if spec.ReleaseType == releaseTypeDaily {
		lastBreakingVersion := getLastBreakingVersion(dailyUpgraders, v)
		if lastBreakingVersion == nil {
			lastBreakingVersion = minUpgradableDailyVersion
		}
		spec.MinimumUpgradableVersions = MinimumVersionConstraints{releaseTypeDaily: lastBreakingVersion}
	} else {
		lastBreakingVersion := getLastBreakingVersion(mainUpgraders, v)
		if lastBreakingVersion == nil {
			lastBreakingVersion = minUpgradableStableVersion
		}
		// all mainline release types support upgrade from stable release
		spec.MinimumUpgradableVersions = MinimumVersionConstraints{
			releaseTypeStable: semver.New(lastBreakingVersion.Major(), lastBreakingVersion.Minor(), lastBreakingVersion.Patch(), "", ""),
		}
		switch spec.ReleaseType {
		// both stable and rc release types support upgrade from stable/rc release
		case releaseTypeStable, releaseTypeRC:
			spec.MinimumUpgradableVersions[releaseTypeRC] = semver.New(lastBreakingVersion.Major(), lastBreakingVersion.Minor(), lastBreakingVersion.Patch(), formatPrereleaseVersion(releaseTypeRC, 0), "")
		case releaseTypeAlpha:
			// alpha release type supports upgrade from the last alpha release
			// if it exists
			// and no breaking change is added to the current version
			if !spec.AddedBreakingChange && spec.ReleaseNum > 0 {
				spec.MinimumUpgradableVersions[releaseTypeAlpha] = semver.New(spec.Major, spec.Minor, spec.Patch, formatPrereleaseVersion(releaseTypeAlpha, spec.ReleaseNum-1), "")
			}
		}
	}

	return spec, nil
}

func getLastBreakingVersion(upgraders []breakingUpgrader, current *semver.Version) *semver.Version {
	sort.Slice(upgraders, func(i, j int) bool {
		return upgraders[i].Version().LessThan(upgraders[j].Version())
	})
	for i := len(upgraders) - 1; i >= 0; i-- {
		if !upgraders[i].AddedBreakingChange() {
			continue
		}
		if upgraders[i].Version().GreaterThanEqual(current) {
			continue
		}
		return upgraders[i].Version()
	}
	return nil
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
