package mesh

import (
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

const (
	// HTTPMeshWaveAnnotation selects progressive namespace waves for full HTTP mesh.
	// 0 = off (legacy mesh-in only), 1 = gateway/mesh, 2 = os-*, 3 = user-space, 4 = all app HTTP.
	HTTPMeshWaveAnnotation = "traffic.olares.io/http-mesh-wave"
	// LinkerdOpaquePortsAnnotation skips non-HTTP ports from protocol detection.
	LinkerdOpaquePortsAnnotation = "config.linkerd.io/opaque-ports"
	// LinkerdDefaultInboundPolicyAnnotation is the Linkerd-native inbound authz default.
	LinkerdDefaultInboundPolicyAnnotation = "config.linkerd.io/default-inbound-policy"
	// LinkerdServerTLSModeAnnotation is an operational label mirroring DesiredMTLSMode
	// (Linkerd enforcement uses default-inbound-policy / Server CRs, not this key alone).
	LinkerdServerTLSModeAnnotation = "traffic.olares.io/mtls-mode"

	MTLSModePermissive = "Permissive"
	MTLSModeRequired   = "Required"

	// Linkerd inbound policy values (proxy API).
	InboundPolicyAllUnauthenticated = "all-unauthenticated"
	InboundPolicyClusterAuthenticated = "cluster-authenticated"
)

// ParseHTTPMeshWave reads wave from annotation or runtime config string (default 0).
func ParseHTTPMeshWave(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	if n > 4 {
		return 4
	}
	return n
}

// NamespaceWaveEligible reports whether ns is included in the current HTTP mesh wave.
func NamespaceWaveEligible(namespace string, wave int) bool {
	if wave <= 0 {
		return false
	}
	group := security.InferTrafficGroup(namespace)
	switch wave {
	case 1:
		return group == security.TrafficGroupGateway || group == security.TrafficGroupMesh ||
			group == security.TrafficGroupL4
	case 2:
		return group == security.TrafficGroupGateway || group == security.TrafficGroupMesh ||
			group == security.TrafficGroupL4 ||
			group == security.TrafficGroupOSFramework || group == security.TrafficGroupOSSystem
	case 3:
		return group != security.TrafficGroupApp
	default: // 4+
		return true
	}
}

// DesiredMTLSMode returns Permissive for early waves and Required for full wave.
func DesiredMTLSMode(wave int) string {
	if wave >= 4 {
		return MTLSModeRequired
	}
	if wave >= 1 {
		return MTLSModePermissive
	}
	return ""
}

// DesiredInboundPolicy returns the Linkerd default-inbound-policy for the wave.
// Waves 1–3 stay all-unauthenticated (mesh inject without failing open clients);
// wave 4 requires cluster mesh identity (mTLS) for inbound HTTP.
func DesiredInboundPolicy(wave int) string {
	if wave >= 4 {
		return InboundPolicyClusterAuthenticated
	}
	if wave >= 1 {
		return InboundPolicyAllUnauthenticated
	}
	return ""
}

// ShouldInjectHTTPMesh combines legacy mesh-in with full-HTTP wave eligibility.
// Product steady state requires HTTP mTLS: inject=disabled does not skip wave
//=4. Use opaque-ports for non-HTTP TCP instead of opting out of the mesh.
func ShouldInjectHTTPMesh(injectMeshIn bool, namespace string, wave int, pod *corev1.Pod) bool {
	_ = pod
	if injectMeshIn {
		return true
	}
	return NamespaceWaveEligible(namespace, wave)
}

// AnnotatePodForHTTPMesh sets inject, Linkerd inbound policy, and optional opaque ports.
func AnnotatePodForHTTPMesh(pod *corev1.Pod, enable bool, wave int, opaquePorts string) {
	if pod == nil || !enable {
		return
	}
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations[LinkerdInjectAnnotation] = LinkerdInjectEnabled
	if mode := DesiredMTLSMode(wave); mode != "" {
		pod.Annotations[LinkerdServerTLSModeAnnotation] = mode
	}
	if policy := DesiredInboundPolicy(wave); policy != "" {
		pod.Annotations[LinkerdDefaultInboundPolicyAnnotation] = policy
	}
	if opaquePorts != "" {
		pod.Annotations[LinkerdOpaquePortsAnnotation] = opaquePorts
	}
}
