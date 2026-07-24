package security

import netv1 "k8s.io/api/networking/v1"

const (
	// RouteControlManagedByLabel marks NetworkPolicies owned by app-service
	// routecontrol reconcilers (not security-controller templates).
	RouteControlManagedByLabel = "app.kubernetes.io/managed-by"
	RouteControlManagedByValue = "app-service"
	// RouteControlComponentLabel distinguishes routecontrol NP from other
	// app-service writers (webhooks, etc.).
	RouteControlComponentLabel = "app.kubernetes.io/component"
	RouteControlComponentValue = "route-control"
	// CallerJWTComponentValue marks NetworkPolicies owned by the caller JWT issuer.
	CallerJWTComponentValue = "caller-jwt"
)

// IsRouteControlManagedNP reports whether np is owned by routecontrol and must
// not be deleted by security-controller namespace sweeps.
func IsRouteControlManagedNP(np *netv1.NetworkPolicy) bool {
	if np == nil || np.Labels == nil {
		return false
	}
	return np.Labels[RouteControlManagedByLabel] == RouteControlManagedByValue &&
		np.Labels[RouteControlComponentLabel] == RouteControlComponentValue
}

// IsCallerJWTManagedNP reports whether np is owned by the caller JWT issuer
// (JWKS ingress allow-list for Envoy Gateway).
func IsCallerJWTManagedNP(np *netv1.NetworkPolicy) bool {
	if np == nil || np.Labels == nil {
		return false
	}
	return np.Labels[RouteControlManagedByLabel] == RouteControlManagedByValue &&
		np.Labels[RouteControlComponentLabel] == CallerJWTComponentValue
}

// IsAppServiceManagedExternalNP reports whether np is owned by an app-service
// controller other than security-controller templates.
func IsAppServiceManagedExternalNP(np *netv1.NetworkPolicy) bool {
	return IsRouteControlManagedNP(np) || IsCallerJWTManagedNP(np)
}
