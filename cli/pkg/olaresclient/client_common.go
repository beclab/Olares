package olaresclient

import "github.com/Masterminds/semver/v3"

// baseClient holds state common to every version implementation. Concrete
// version clients embed it (directly or transitively) to satisfy VersionAware.
type baseClient struct {
	// version is the FULL detected backend version (with any prerelease),
	// not the core patch the registry keyed on.
	version *semver.Version
}

// Version implements VersionAware.
func (c baseClient) Version() *semver.Version { return c.version }
