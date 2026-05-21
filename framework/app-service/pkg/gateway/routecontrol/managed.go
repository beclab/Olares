package routecontrol

import networkingv1 "k8s.io/api/networking/v1"

// IsManagedNetworkPolicy reports whether np is owned by routecontrol and must not
// be deleted by security-controller namespace sweeps.
func IsManagedNetworkPolicy(np *networkingv1.NetworkPolicy) bool {
	if np == nil {
		return false
	}
	return np.Name == NetworkPolicyName && np.Labels[ManagedByLabel] == ManagedByValue
}
