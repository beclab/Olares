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
	// LinkerdServerTLSModeAnnotation documents desired MeshTLS mode for the wave.
	LinkerdServerTLSModeAnnotation = "traffic.olares.io/mtls-mode"
	MTLSModePermissive             = "Permissive"
	MTLSModeRequired               = "Required"
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
		return group == security.TrafficGroupGateway || group == security.TrafficGroupMesh
	case 2:
		return group == security.TrafficGroupGateway || group == security.TrafficGroupMesh ||
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

// ShouldInjectHTTPMesh combines legacy mesh-in with full-HTTP wave eligibility.
// Explicit linkerd.io/inject=disabled always wins. Non-HTTP-only pods can pass
// opaquePorts to keep TCP opaque while still injecting for HTTP ports.
func ShouldInjectHTTPMesh(injectMeshIn bool, namespace string, wave int, pod *corev1.Pod) bool {
	if pod != nil && pod.Annotations != nil && pod.Annotations[LinkerdInjectAnnotation] == LinkerdInjectDisabled {
		return false
	}
	if injectMeshIn {
		return true
	}
	return NamespaceWaveEligible(namespace, wave)
}

// AnnotatePodForHTTPMesh sets inject and mtls-mode annotations for the wave.
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
	if opaquePorts != "" {
		pod.Annotations[LinkerdOpaquePortsAnnotation] = opaquePorts
	}
}
