package mesh

import (
	"fmt"
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

	LinkerdSkipInboundPortsAnnotation  = "config.linkerd.io/skip-inbound-ports"
	LinkerdSkipOutboundPortsAnnotation = "config.linkerd.io/skip-outbound-ports"
	LinkerdSkipSubnetsAnnotation       = "config.linkerd.io/skip-subnets"
	LinkerdProxyContainer              = "linkerd-proxy"
	LinkerdProxyInboundPolicyEnv       = "LINKERD2_PROXY_INBOUND_DEFAULT_POLICY"

	MTLSModePermissive = "Permissive"
	MTLSModeRequired   = "Required"

	// Linkerd inbound policy values (proxy API).
	InboundPolicyAllUnauthenticated   = "all-unauthenticated"
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
// =4. Use opaque-ports for non-HTTP TCP instead of opting out of the mesh.
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

// CheckHTTPMeshInterception rejects Pod settings that can bypass Linkerd.
func CheckHTTPMeshInterception(pod *corev1.Pod) error {
	if pod.Spec.HostNetwork || pod.Annotations[LinkerdSkipInboundPortsAnnotation] != "" ||
		pod.Annotations[LinkerdSkipOutboundPortsAnnotation] != "" || pod.Annotations[LinkerdSkipSubnetsAnnotation] != "" {
		return fmt.Errorf("HTTP mesh forbids host networking and Linkerd interception exclusions")
	}
	for _, containers := range [][]corev1.Container{pod.Spec.Containers, pod.Spec.InitContainers} {
		for _, container := range containers {
			for _, port := range container.Ports {
				if port.HostPort != 0 {
					return fmt.Errorf("HTTP mesh forbids exposing container %s through hostPort", container.Name)
				}
			}
		}
	}
	return nil
}

// HasHTTPMeshInit checks that the live Pod was created with an intercepted
// HTTP path. Old init rules survive ConfigMap changes until Pod recreation.
func HasHTTPMeshInit(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.InitContainers {
		if container.Name != "linkerd-init" {
			continue
		}
		for i, arg := range container.Args {
			flag, ports, hasValue := strings.Cut(arg, "=")
			if flag != "--inbound-ports-to-ignore" && flag != "--outbound-ports-to-ignore" && flag != "--subnets-to-ignore" {
				continue
			}
			if !hasValue {
				if i+1 >= len(container.Args) {
					return false
				}
				ports = container.Args[i+1]
			}
			for _, port := range strings.Split(ports, ",") {
				switch strings.TrimSpace(port) {
				case "":
				case "3389", "4190", "4191":
					if flag != "--inbound-ports-to-ignore" {
						return false
					}
				default:
					return false
				}
			}
		}
		return true
	}
	return false
}

// ProxyHasInboundPolicy checks the configuration delivered to the running proxy.
func ProxyHasInboundPolicy(pod *corev1.Pod, policy string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name != LinkerdProxyContainer {
			continue
		}
		for _, env := range container.Env {
			if env.Name == LinkerdProxyInboundPolicyEnv && env.ValueFrom == nil && env.Value == policy {
				return true
			}
		}
	}
	return false
}

// ConfigurePodHTTPMesh applies the mesh posture before any sandbox admission
// return, including pods that do not need a legacy application sidecar.
func ConfigurePodHTTPMesh(pod *corev1.Pod, namespace string, wave int, opaquePorts string, eligible, injectMeshIn bool) (bool, error) {
	if !eligible || !ShouldInjectHTTPMesh(injectMeshIn, namespace, wave, pod) {
		return false, nil
	}
	if err := CheckHTTPMeshInterception(pod); err != nil {
		return false, err
	}
	AnnotatePodForHTTPMesh(pod, true, wave, opaquePorts)
	// The injector may have run first when the Pod or namespace was already opted in.
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		if container.Name != LinkerdProxyContainer {
			continue
		}
		for j := range container.Env {
			if container.Env[j].Name == LinkerdProxyInboundPolicyEnv {
				container.Env[j].Value = DesiredInboundPolicy(wave)
				container.Env[j].ValueFrom = nil
			}
		}
	}
	return true, nil
}
