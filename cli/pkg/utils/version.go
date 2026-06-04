package utils

import "github.com/Masterminds/semver/v3"

func ParseOlaresVersionString(versionString string) (*semver.Version, error) {
	// todo: maybe some other custom processing only for olares
	return semver.NewVersion(versionString)
}

// CoreVersion returns the Major.Minor.Patch "core" of v with the prerelease
// and build-metadata stripped. It is the canonical key for version-aware
// dispatch: an Olares backend reported as 1.12.6, 1.12.6-20260603 (daily) or
// 1.12.6-alpha1 (prerelease) all share the same core 1.12.6 and therefore the
// same client implementation. Stripping the prerelease also sidesteps the
// Masterminds/semver rule that constraints exclude prerelease versions (and
// that 1.12.6-alpha1 sorts *before* 1.12.6), which makes naive range matching
// silently drop daily / alpha builds.
func CoreVersion(v *semver.Version) *semver.Version {
	if v == nil {
		return nil
	}
	core := semver.New(v.Major(), v.Minor(), v.Patch(), "", "")
	return core
}

// SamePatchLevel reports whether a and b share the same Major.Minor.Patch,
// ignoring any prerelease / build metadata. Mirrors the helper of the same
// intent in pkg/upgrade (samePatchLevelVersion) so version-line semantics stay
// consistent across the codebase.
func SamePatchLevel(a, b *semver.Version) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Major() == b.Major() && a.Minor() == b.Minor() && a.Patch() == b.Patch()
}
