package appcfg

import appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

// IsGatewaySharedApp reports whether the Application participates in the shared
// Envoy Gateway path (SRR, HTTPRoute, L4 route-mode=gateway).
//
// Qualifying apps:
//   - v3 installs (app.bytetrade.io/api-version=v3), or
//   - v2 cluster-scoped apps with spec.sharedEntrances (multi-chart shared
//     subchart pilots such as ollamav2 + ollamaserver).
//
// Phase A pilots may use v2 charts without manifest migration; production can
// still prefer v3 for a single {app}-shared namespace.
func IsGatewaySharedApp(app *appv1alpha1.Application) bool {
	if app == nil || len(app.Spec.SharedEntrances) == 0 {
		return false
	}
	return IsV3(app) || IsClusterScoped(app)
}
