// Package uniauth pins app-gateway control-plane posture.
// Install and upgrade complete as the only steady state: no runtime kill switch.
package uniauth

// EntryMTLSDesired is always true in product steady state.
func EntryMTLSDesired() bool {
	return true
}
