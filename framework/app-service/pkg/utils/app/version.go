package app

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// IsDowngrade reports whether target is an older release than deployed.
// An empty target requests the newest available release and skips comparison.
// Otherwise, both versions must be parseable as semver.
func IsDowngrade(target, deployed string) (bool, error) {
	if target == "" {
		return false, nil
	}

	targetVersion, err := semver.NewVersion(target)
	if err != nil {
		return false, fmt.Errorf("parse target version %q: %w", target, err)
	}

	deployedVersion, err := semver.NewVersion(deployed)
	if err != nil {
		return false, fmt.Errorf("parse deployed version %q: %w", deployed, err)
	}

	return targetVersion.LessThan(deployedVersion), nil
}
