package upgrade

import (
	"bytetrade.io/web3os/installer/pkg/core/task"
	"github.com/Masterminds/semver/v3"
)

// versionMatcher checks if the specified version matches its condition
type versionMatcher interface {
	Match(version *semver.Version) bool
}

// explicitVersionMatcher matches the specified version by a range of explicitly
// set version range by min/max version
// and additionally explicitly included/excluded versions
// if any type of condition is not set, that check is omitted
// i.e., if min is not set, there's no limit on the minimum version
// and if no condition is set, the matcher matches all non-nil versions
type explicitVersionMatcher struct {
	min     *semver.Version
	max     *semver.Version
	include []*semver.Version
	exclude []*semver.Version
}

func (m *explicitVersionMatcher) Match(version *semver.Version) bool {
	if version == nil {
		return false
	}
	for _, v := range m.include {
		if v.Equal(version) {
			return true
		}
	}
	for _, v := range m.exclude {
		if v.Equal(version) {
			return false
		}
	}
	if m.min != nil && version.LessThan(m.min) {
		return false
	}
	if m.max != nil && version.GreaterThan(m.max) {
		return false
	}
	return true
}

// todo: do we need to check at least 1.12 in cli?
var anyVersion versionMatcher = &explicitVersionMatcher{}
var atLeasVersion112 versionMatcher = &explicitVersionMatcher{min: semver.New(1, 12, 0, "1", "")}

type upgradeTask struct {
	Task    task.Interface
	Current versionMatcher
	Target  versionMatcher
}

func (t *upgradeTask) Match(current, target *semver.Version) bool {
	return t.Current.Match(current) && t.Target.Match(target)
}
