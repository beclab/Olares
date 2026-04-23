package oac

import (
	"github.com/Masterminds/semver/v3"
)

// NewOlaresManifestVersion is the threshold (inclusive) at which the
// OlaresManifest parsing pipeline switches from the legacy helm-template
// path to the literal-parse path. Manifests whose olaresManifest.version
// is at or above this version are considered "new" by
// IsNewOlaresManifestVersion.
const NewOlaresManifestVersion = "0.12.0"

// IsNewOlaresManifestVersion reports whether the given olaresManifest.version
// is at or above the NewOlaresManifestVersion (0.12.0) threshold. Empty or
// malformed versions return false (treated as legacy).
//
// This is the same predicate downstream tooling (e.g. app-service) needs in
// order to branch between legacy-only logic and modern-manifest behaviour
// without re-implementing the semver comparison.
func IsNewOlaresManifestVersion(version string) bool {
	if version == "" {
		return false
	}
	c, err := semver.NewConstraint(">= " + NewOlaresManifestVersion)
	if err != nil {
		return false
	}
	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}
	return c.Check(v)
}
