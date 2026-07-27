package appcfg

import (
	"fmt"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// IsGatewaySharedApp reports whether the Application participates in the shared
// Envoy Gateway path (SRR + HTTPRoute). Qualifying apps are shared cluster-wide
// installs (options.shared) or cluster-scoped apps. sharedEntrances are optional;
// apps with only spec.entrances still qualify and receive application-class SRRs.
func IsGatewaySharedApp(app *appv1alpha1.Application) bool {
	if app == nil {
		return false
	}
	return IsShared(app) // || IsClusterScoped(app)
}

// LogicalHostPattern returns the canonical shared gateway host pattern:
// <sharedEntranceID>.shared.<platformDomain>.
func LogicalHostPattern(appid string, entranceIndex, entranceCount int, platformDomain string, isShared bool) (string, error) {
	appid = strings.ToLower(strings.TrimSpace(appid))
	if appid == "" {
		return "", fmt.Errorf("appid is empty")
	}
	if entranceIndex < 0 || entranceIndex >= entranceCount {
		return "", fmt.Errorf("shared entrance index out of range: index=%d count=%d", entranceIndex, entranceCount)
	}
	platformDomain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(platformDomain, ".")))
	if platformDomain == "" {
		return "", fmt.Errorf("platformDomain is empty")
	}
	if isShared {
		return fmt.Sprintf("%s.shared.%s", appv1alpha1.SharedEntranceID(appid, entranceIndex, entranceCount), platformDomain), nil
	}
	return fmt.Sprintf("%s.shared.%s", appv1alpha1.SharedEntranceIDV2(appid, entranceIndex, entranceCount), platformDomain), nil
}
