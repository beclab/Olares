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

// IsSharedServerApp reports whether the Application is a shared server that
// must participate in the shared in-cluster routing infrastructure: the
// shared-hosts ConfigMap fan-out, the entrance TLS caller-NS replica
// reconciler, and the callee->owner / namespace->owner indexes.
//
// Qualifying apps:
//   - v3 installs (app.bytetrade.io/api-version=v3) that expose shared
//     entrances — v3 is inherently shared, so the v3 marker alone qualifies
//     and these apps no longer need settings.clusterScoped=true; or
//   - legacy v2 cluster-scoped apps (settings.clusterScoped=true), kept for
//     backward compatibility with multi-chart shared pilots (e.g. ollamav2).
//
// This is intentionally additive over the historical clusterScoped-only
// predicate: every app that qualified before (clusterScoped=true) still
// qualifies, and v3 shared servers are newly included.
func IsSharedServerApp(app *appv1alpha1.Application) bool {
	if app == nil {
		return false
	}
	if IsClusterScoped(app) {
		return true
	}
	return IsV3(app) && len(app.Spec.SharedEntrances) > 0
}
