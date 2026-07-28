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
