// Package uniauth provides the single end-state switch for full HTTP mesh.
package uniauth

import (
	"os"
	"strings"
)

const (
	// EndStateEnv turns on the product end-state in one shot (default off).
	EndStateEnv = "OLARES_UNIAUTH_ENDSTATE"
	// EndStateWave is the mesh wave used when end-state is enabled (all HTTP + Required inbound).
	EndStateWave = 4
)

// EndStateEnabled reports whether operators enabled the full end-state.
func EndStateEnabled() bool {
	return truthy(os.Getenv(EndStateEnv))
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
