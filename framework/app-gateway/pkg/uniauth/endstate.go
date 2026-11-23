// Package uniauth pins app-gateway control-plane posture.
// Install and upgrade complete as the only steady state: no runtime kill switch.
package uniauth

// PolicyCompilationDesired is always true in product steady state.
func PolicyCompilationDesired() bool {
	return true
}

// EntryMTLSDesired is always true in product steady state.
func EntryMTLSDesired() bool {
	return true
}

// AccessPolicyReconcileDesired is always true in product steady state.
func AccessPolicyReconcileDesired() bool {
	return true
}
