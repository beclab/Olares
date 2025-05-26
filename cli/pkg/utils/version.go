package utils

import "github.com/Masterminds/semver/v3"

func ParseOlaresVersionString(versionString string) (*semver.Version, error) {
	// todo: maybe some other custom processing only for olares
	return semver.NewVersion(versionString)
}
