package terminus

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// envoyGatewayCRDsPresent reports whether Envoy Gateway / Gateway API CRDs are already registered.
func envoyGatewayCRDsPresent(cfg *rest.Config) bool {
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false
	}
	for _, gv := range []string{
		"gateway.envoyproxy.io/v1alpha1",
		"gateway.networking.k8s.io/v1",
	} {
		if _, err := dc.ServerResourcesForGroupVersion(gv); err != nil {
			return false
		}
	}
	return true
}
