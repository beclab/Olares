package app

import (
	"github.com/Masterminds/semver/v3"
	"k8s.io/klog/v2"
)

// IsDowngrade reports whether target is an older release than deployed.
//
// It refuses to guess. An uploaded or DevBox chart carries whatever version
// string its author wrote, so a version that is not semver on either side is
// not treated as a downgrade; neither is an empty target, which means "the
// newest the repo has". utils.MatchVersion cannot be used for this: it reports
// a match for an empty version and a mismatch for anything unparseable, so
// both of its defaults point towards inventing a refusal.
func IsDowngrade(target, deployed string) bool {
	if target == "" || deployed == "" {
		return false
	}

	targetVersion, err := semver.NewVersion(target)
	if err != nil {
		klog.Infof("skipping the downgrade check, target version %q is not semver: %v", target, err)
		return false
	}

	deployedVersion, err := semver.NewVersion(deployed)
	if err != nil {
		klog.Infof("skipping the downgrade check, deployed version %q is not semver: %v", deployed, err)
		return false
	}

	return targetVersion.LessThan(deployedVersion)
}
