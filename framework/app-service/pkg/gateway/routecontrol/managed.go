package routecontrol

import (
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	networkingv1 "k8s.io/api/networking/v1"
)

// IsManagedNetworkPolicy reports whether np is owned by routecontrol and must not
// be deleted by security-controller namespace sweeps.
func IsManagedNetworkPolicy(np *networkingv1.NetworkPolicy) bool {
	if np == nil {
		return false
	}
	if np.Labels[ManagedByLabel] != ManagedByValue {
		return false
	}
	switch np.Name {
	case NetworkPolicyName, security.SharedLinkerdMeshIngressNPName,
		security.AppGatewayInClusterCallerIngressNPName,
		// Legacy caller egress NP names retained in the whitelist so existing
		// clusters do not bounce-create-delete during the upgrade window before
		// CallerReconciler.cleanupCallerResources GCs them. Safe to remove in a
		// future release once all clusters have rolled past v1.0.
		security.CallerToAppGatewayEgressNPName, security.CallerMeshEgressNPName,
		security.CallerDNSEgressNPName, security.CallerMiddlewareEgressNPName:
		return true
	default:
		return false
	}
}
