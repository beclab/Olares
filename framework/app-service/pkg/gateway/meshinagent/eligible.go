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

// ApplicationDeclaresSharedAccess reports whether the Application manifest
// declares a dependency on Shared apps (P1 declarative deps contract).
func ApplicationDeclaresSharedAccess(app *appv1alpha1.Application) bool {
	if app == nil || app.Spec.Settings == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(app.Spec.Settings[SettingNeedsSharedAccess]), "true") {
		return true
	}
	if strings.TrimSpace(app.Spec.Settings[SettingSharedAppDeps]) != "" {
		return true
	}
	return strings.TrimSpace(app.Spec.Settings[SettingClusterAppRef]) != ""
}

// ShouldInject reports whether the mesh-in agent should be injected into a pod.
// Shared provider apps and middleware workloads never receive the agent.
func ShouldInject(app *appv1alpha1.Application, isSharedApp bool) bool {
	if isSharedApp || app == nil {
		return false
	}
	return ApplicationDeclaresSharedAccess(app)
}
