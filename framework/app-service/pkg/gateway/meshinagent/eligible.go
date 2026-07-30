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

	// GatewayIPsEnv carries comma/space-separated gateway data-plane ClusterIPs
	// injected by admission (preferred over in-pod DNS for iptables -d).
	GatewayIPsEnv = "MESH_IN_AGENT_GATEWAY_IPS"

	// GatewayDataServiceName is the ClusterIP Service for Shared HTTP/HTTPS data plane.
	GatewayDataServiceName = "app-gateway-data"

	// FailClosedEnv tells the agent to reject traffic when no valid JWT is present.
	FailClosedEnv = "MESH_IN_AGENT_FAIL_CLOSED"

	// LabelSharedCallerOutbound optionally narrows mesh-in to workloads that
	// initiate Shared calls. When absent, non-entrance pods of a caller app inject.
	LabelSharedCallerOutbound = "gateway.olares.io/shared-caller-outbound"
)

// ApplicationDeclaresSharedAccess reports whether the app is a Shared caller
// (persisted decide=true or named edges). Shared provider apps are filtered by callers.
func ApplicationDeclaresSharedAccess(app *appv1alpha1.Application) bool {
	if app == nil {
		return false
	}
	return DeclaresSharedCaller(app.Spec.Settings)
}

// ShouldInject reports whether the mesh-in agent may be considered for a pod.
// Injection follows only persisted decide / named callee facts (DeclaresSharedCaller).
// Live Decide eligibility is no longer evaluated here; the controller persists
// decide via ApplyDecide. isSharedApp is retained for admission call sites.
func ShouldInject(app *appv1alpha1.Application, isSharedApp bool) bool {
	_ = isSharedApp
	return ApplicationDeclaresSharedAccess(app)
}

// AllowOutboundMeshIn reports whether a pod that already passed ShouldInject
// may receive mesh-in. Explicit LabelSharedCallerOutbound wins (false is an
// escape hatch). Non-entrance pods inject. Entrance pods inject only when the
// Application DeclaresSharedCaller (composite Shared caller).
func AllowOutboundMeshIn(isEntrancePod, declaresSharedCaller bool, podLabels map[string]string) bool {
	if podLabels != nil {
		if v, ok := podLabels[LabelSharedCallerOutbound]; ok {
			return strings.EqualFold(strings.TrimSpace(v), "true")
		}
	}
	if !isEntrancePod {
		return true
	}
	return declaresSharedCaller
}

// HasIntentOnly reports needsSharedAccess without named callees.
// Ordinary apps may still inject via Decide; this helper is for diagnostics.
func HasIntentOnly(settings map[string]string) bool {
	if settings == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(settings[SettingNeedsSharedAccess]), "true") {
		return false
	}
	return len(ParseCallees(settings)) == 0
}
