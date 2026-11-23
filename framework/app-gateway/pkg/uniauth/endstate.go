// Package uniauth provides the product end-state switch for app-gateway control plane.
// New install / upgrade steady-state: end-state is ON unless explicitly disabled.
package uniauth

import (
	"os"
	"strings"
)

const EndStateEnv = "OLARES_UNIAUTH_ENDSTATE"

// EndStateEnabled reports whether the cluster should run the product end-state.
// If OLARES_UNIAUTH_ENDSTATE is set, use it; otherwise default true.
func EndStateEnabled() bool {
	if v, ok := os.LookupEnv(EndStateEnv); ok {
		return truthy(v)
	}
	return true
}

// PolicyCompilationDesired is true when end-state or explicit compilation is on.
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
