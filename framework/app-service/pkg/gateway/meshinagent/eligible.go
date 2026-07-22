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

	// FailClosedEnv tells the agent to reject traffic when no valid JWT is present.
	FailClosedEnv = "MESH_IN_AGENT_FAIL_CLOSED"
)

// ApplicationDeclaresSharedAccess reports whether named caller→callee edges exist.
// needsSharedAccess alone is NOT sufficient (ARCH Q13 / ADR-IC-08).
func ApplicationDeclaresSharedAccess(app *appv1alpha1.Application) bool {
	if app == nil {
		return false
	}
	return DeclaresSharedCaller(app.Spec.Settings)
}

// ShouldInject reports whether the mesh-in agent should be injected into a pod.
// Shared provider apps and middleware workloads never receive the agent.
func ShouldInject(app *appv1alpha1.Application, isSharedApp bool) bool {
	if isSharedApp || app == nil {
		return false
	}
	return DeclaresSharedCaller(app.Spec.Settings)
}

// HasIntentOnly reports needsSharedAccess without named callees (must not inject).
func HasIntentOnly(settings map[string]string) bool {
	if settings == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(settings[SettingNeedsSharedAccess]), "true") {
		return false
	}
	return len(ParseCallees(settings)) == 0
}
