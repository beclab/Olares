// Package uniauth provides the single end-state switch for full HTTP mesh
// control-plane components hosted in app-gateway.
package uniauth

import (
	"os"
	"strings"
)

const EndStateEnv = "OLARES_UNIAUTH_ENDSTATE"

// EndStateEnabled reports whether operators enabled the full product end-state.
func EndStateEnabled() bool {
	return truthy(os.Getenv(EndStateEnv))
}

// PolicyCompilationDesired is true when end-state or explicit compilation env is on.
func PolicyCompilationDesired(explicitEnv string, cmEnabled bool) bool {
	return EndStateEnabled() || truthy(explicitEnv) || cmEnabled
}

// EntryMTLSDesired is true when end-state or explicit entry PKI env is on.
func EntryMTLSDesired(explicitEnv string) bool {
	return EndStateEnabled() || truthy(explicitEnv)
}

// AccessPolicyReconcileDesired is true when end-state or explicit reconcile env is on.
func AccessPolicyReconcileDesired(explicitEnv string) bool {
	return EndStateEnabled() || truthy(explicitEnv)
}

func truthy(v string) bool {
	v = strings.TrimSpace(v)
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}
