// Package uniauth pins the product security posture for full HTTP mesh.
// Install and upgrade complete as the only steady state: no runtime kill switch.
package uniauth

const (
	// HTTPMeshWave is the fixed product wave: all HTTP namespaces + cluster-authenticated.
	HTTPMeshWave = 4
)

// EffectiveHTTPMeshWave always returns the product wave (HTTP mTLS required).
func EffectiveHTTPMeshWave() int {
	return HTTPMeshWave
}

// TrafficGroupNPDesired is always true in product steady state.
func TrafficGroupNPDesired() bool {
	return true
}
