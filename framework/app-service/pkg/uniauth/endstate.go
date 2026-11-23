// Package uniauth provides the product end-state switch for full HTTP mesh.
// New install / upgrade steady-state: end-state is ON unless explicitly disabled.
package uniauth

import (
	"os"
	"strings"
)

const (
	// EndStateEnv is injected from ConfigMap endState (SSOT) on install/upgrade.
	// Set to "false" only for emergency rollback to the pre-uniauth baseline.
	EndStateEnv = "OLARES_UNIAUTH_ENDSTATE"
	// EndStateWave is the mesh wave for product steady-state.
	EndStateWave = 4
)

// EndStateEnabled reports whether the cluster should run the product end-state.
// Resolution: if OLARES_UNIAUTH_ENDSTATE is set, use it; otherwise default true
// (install/upgrade complete == steady state).
func EndStateEnabled() bool {
	if v, ok := os.LookupEnv(EndStateEnv); ok {
		return truthy(v)
	}
	return true
}

// EffectiveHTTPMeshWave returns 4 when end-state is on; otherwise parses waveRaw.
func EffectiveHTTPMeshWave(waveRaw string, parse func(string) int) int {
	if EndStateEnabled() {
		return EndStateWave
	}
	return parse(waveRaw)
}

// TrafficGroupNPDesired is true when end-state is on or the dedicated NP env is set.
func TrafficGroupNPDesired(npEnv string) bool {
	return EndStateEnabled() || truthy(npEnv)
}

func truthy(v string) bool {
	v = strings.TrimSpace(v)
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}
