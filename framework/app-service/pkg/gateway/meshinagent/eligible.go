package meshinagent

import (
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

const (
	// ContainerName is the mesh-in agent sidecar injected into Shared consumer pods.
	ContainerName = "olares-mesh-in-agent"

	SettingNeedsSharedAccess = "needsSharedAccess"
	SettingSharedAppDeps     = "sharedAppDeps"
	SettingClusterAppRef     = "clusterAppRef"

	JWTSecretVolumeName = "mesh-in-jwt"
	JWTSecretMountPath  = "/var/run/olares/mesh-in-jwt"

	// DefaultGatewayHost is the Shared HTTP data-plane Service (namespace os-gateway).
	DefaultGatewayHost = "app-gateway-data.os-gateway.svc"

	// FailClosedEnv tells the agent to reject traffic when no valid JWT is present.
	FailClosedEnv = "MESH_IN_AGENT_FAIL_CLOSED"
)

// ApplicationDeclaresSharedAccess reports whether the app is a Shared caller
// (persisted decide=true or named edges). Shared provider apps are filtered by callers.
func ApplicationDeclaresSharedAccess(app *appv1alpha1.Application) bool {
	if app == nil {
		return false
	}
	return DeclaresSharedCaller(app.Spec.Settings)
}

// ShouldInject reports whether the mesh-in agent should be injected into a pod.
// Shared provider apps never receive the agent. Prefer persisted decide; else run Decide.
func ShouldInject(app *appv1alpha1.Application, isSharedApp bool) bool {
	if isSharedApp || app == nil {
		return false
	}
	if DeclaresSharedCaller(app.Spec.Settings) {
		return true
	}
	name := strings.TrimSpace(app.Spec.Name)
	if name == "" {
		name = app.Name
	}
	return Decide(name, app.Spec.Settings, DefaultRules()).Inject
}

// HasIntentOnly reports needsSharedAccess without named callees.
// Under B′ eligibility, intent-only apps still inject after Decide; this helper
// remains for diagnostics distinguishing named edges from intent.
func HasIntentOnly(settings map[string]string) bool {
	if settings == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(settings[SettingNeedsSharedAccess]), "true") {
		return false
	}
	return len(ParseCallees(settings)) == 0
}
