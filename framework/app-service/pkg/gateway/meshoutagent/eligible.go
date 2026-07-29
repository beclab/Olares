package meshoutagent

import (
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

const (
	// ContainerName is the minimal mesh-out agent replacing outbound :15001 envoy.
	ContainerName = "olares-mesh-out-agent"

	SATokenVolumeName = "mesh-out-sa-token"
	SATokenMountPath  = "/var/run/secrets/olares.io/serviceaccount"

	ListenPort     = 15001
	ListenPortName = "mesh-out-http"

	// FailClosedEnv prevents forwarding when the projected SA token is missing.
	FailClosedEnv = "MESH_OUT_AGENT_FAIL_CLOSED"
)

// HasProviderPermission reports whether the pod should receive outbound
// system-server forwarding (WI-OC-EGRESS-01 §2.1).
func HasProviderPermission(perms []appcfg.ProviderPermission) bool {
	return len(perms) > 0
}

// ShouldInject reports whether the mesh-out agent should replace envoy outbound
// for this workload. Shared inbound-only pods never receive the agent.
func ShouldInject(isSharedApp bool, perms []appcfg.ProviderPermission) bool {
	if isSharedApp {
		return false
	}
	return HasProviderPermission(perms)
}
