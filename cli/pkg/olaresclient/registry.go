package olaresclient

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"

	"github.com/beclab/Olares/cli/pkg/utils"
)

// clientFactory builds an OlaresClient for a concrete backend version. The
// FULL backendVersion (including prerelease) is passed in so an implementation
// can branch on daily / alpha if it ever needs to, even though the registry
// keys on the core patch version.
type clientFactory func(backendVersion *semver.Version) (OlaresClient, error)

// registeredClient pairs a core patch version key with its factory.
type registeredClient struct {
	patch   *semver.Version // core Major.Minor.Patch (prerelease stripped)
	factory clientFactory
}

// registeredClients is the ascending-sorted registry of version
// implementations. Populated by each client file's init().
var registeredClients []registeredClient

// defaultFactory builds the fallback client used when the backend version is
// below the lowest registered implementation, or unknown.
var defaultFactory clientFactory

// registerClientFactory registers factory under the core patch version of
// patch, keeping registeredClients sorted ascending so GetClient's floor match
// can scan linearly.
func registerClientFactory(patch *semver.Version, factory clientFactory) {
	registeredClients = append(registeredClients, registeredClient{
		patch:   utils.CoreVersion(patch),
		factory: factory,
	})
	sort.Slice(registeredClients, func(i, j int) bool {
		return registeredClients[i].patch.LessThan(registeredClients[j].patch)
	})
}

// registerDefaultFactory records the fallback factory. Last registration wins;
// there is only one default in practice.
func registerDefaultFactory(factory clientFactory) {
	defaultFactory = factory
}

// GetClient selects the version implementation for backendVersion using a
// floor match on the core (Major.Minor.Patch) version: the registered client
// with the greatest core version that is still <= the backend's core version.
// This gives the two fallback layers the design calls for:
//
//   - selection floor: a backend newer than any known implementation (e.g.
//     1.12.7-20260524 when the CLI only knows up to 1.12.6) resolves to the
//     newest known implementation (1.12.6).
//   - method fallback: handled by embedding inside the chosen implementation
//     (see clientV1_12_6).
//
// A backend below the lowest registered version, or a nil version, resolves to
// the default factory; if none is registered an error is returned.
func GetClient(backendVersion *semver.Version) (OlaresClient, error) {
	if backendVersion == nil {
		return useDefault(backendVersion, "backend version is unknown")
	}
	chosen := selectClient(backendVersion)
	if chosen == nil {
		return useDefault(backendVersion, fmt.Sprintf("backend version %s is below the minimum supported version", backendVersion))
	}
	return chosen.factory(backendVersion)
}

// selectClient runs the floor match (greatest registered core version that is
// <= the backend's core version) and returns the chosen entry, or nil when the
// backend is below the lowest registered version. Shared by GetClient and
// SelectedImplementation so the dispatch decision is computed in exactly one
// place.
func selectClient(backendVersion *semver.Version) *registeredClient {
	if backendVersion == nil {
		return nil
	}
	core := utils.CoreVersion(backendVersion)
	var chosen *registeredClient
	for i := range registeredClients {
		if registeredClients[i].patch.Compare(core) <= 0 {
			chosen = &registeredClients[i] // ascending: keep the last one that fits
			continue
		}
		break
	}
	return chosen
}

// SelectedImplementation reports which registered client implementation
// GetClient would dispatch backendVersion to, as a human-readable label
// ("1.12.6", "1.12.5", or "default"). Used by `settings me version --dispatch`
// to show users exactly which version-specific code path their commands take.
func SelectedImplementation(backendVersion *semver.Version) string {
	chosen := selectClient(backendVersion)
	if chosen == nil {
		return "default"
	}
	return chosen.patch.String()
}

func useDefault(backendVersion *semver.Version, reason string) (OlaresClient, error) {
	if defaultFactory == nil {
		return nil, fmt.Errorf("cannot resolve Olares client: %s and no default client is registered", reason)
	}
	return defaultFactory(backendVersion)
}
