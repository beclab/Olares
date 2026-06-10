package appcfg

import (
	"fmt"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// IsGatewaySharedApp reports whether the Application participates in the shared
// Envoy Gateway path (SRR + HTTPRoute). Qualifying apps are shared cluster-wide
// installs or v2 cluster-scoped apps that expose spec.sharedEntrances.
func IsGatewaySharedApp(app *appv1alpha1.Application) bool {
	if app == nil || len(app.Spec.SharedEntrances) == 0 {
		return false
	}
	return IsShared(app) || IsClusterScoped(app)
}

// SharedEntranceID returns the first DNS label for v3 shared entrances:
// appid when count==1, appid+index when count>1.
func SharedEntranceID(appid string, entranceIndex, entranceCount int) (string, error) {
	appid = strings.ToLower(strings.TrimSpace(appid))
	if appid == "" {
		return "", fmt.Errorf("shared entrance id: appid is empty")
	}
	if entranceCount < 1 || entranceCount > 10 {
		return "", fmt.Errorf("shared entrance id: entrance count must be in [1,10]")
	}
	if entranceIndex < 0 || entranceIndex >= entranceCount {
		return "", fmt.Errorf("shared entrance id: entrance index out of range")
	}
	if entranceCount == 1 {
		return appid, nil
	}
	return fmt.Sprintf("%s%d", appid, entranceIndex), nil
}

// LogicalHostPattern returns the SRR hostPattern for a shared entrance. One
// logical pattern covers all viewers: <entranceid>.*.<platformDomain>. The
// literal "*" label is the marker route control expands into a Host regex
// match in the generated HTTPRoute.
func LogicalHostPattern(appid string, entranceIndex, entranceCount int, platformDomain string) (string, error) {
	platformDomain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(platformDomain, ".")))
	if platformDomain == "" {
		return "", fmt.Errorf("logical host pattern: platformDomain is empty")
	}
	entranceID, err := SharedEntranceID(appid, entranceIndex, entranceCount)
	if err != nil {
		return "", err
	}
	return entranceID + ".*." + platformDomain, nil
}
