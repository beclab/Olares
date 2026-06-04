package olaresclient

import "github.com/Masterminds/semver/v3"

// clientDefault is the fallback used when the backend version is unknown or
// below the lowest registered implementation. It behaves like the oldest
// supported line (1.12.5) — the most conservative wire format we still
// understand — maximizing the chance of working against an unexpected backend.
// It is a distinct type (rather than reusing clientV1_12_5 directly) so its
// behavior can be tightened independently later without disturbing the 1.12.5
// path.
type clientDefault struct {
	clientV1_12_5
}

func newClientDefault(backendVersion *semver.Version) (OlaresClient, error) {
	return &clientDefault{clientV1_12_5{baseClient{version: backendVersion}}}, nil
}

func init() {
	registerDefaultFactory(newClientDefault)
}
