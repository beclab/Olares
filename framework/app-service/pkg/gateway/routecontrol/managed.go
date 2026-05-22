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
	case NetworkPolicyName, security.SharedLinkerdMeshIngressNPName:
		return true
	default:
		return false
	}
}
