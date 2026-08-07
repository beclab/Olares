package upgrade

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Platform Envoy allow-list (ADR-SYS-DEENVY-STEADY-01):
// business pods must not run olares-envoy-sidecar; these platform Envoy
// workloads are retained by design (including system-server co-located proxy).
const (
	allowedEnvoyAppL4                = "l4-bfl-proxy"
	allowedEnvoyAppEGData            = "app-gateway-data"
	allowedEnvoyContainerSystemProxy = "proxy"
	allowedEnvoyAppSystemServer      = "systemserver"
)

// isAllowedPlatformEnvoy reports whether a container that looks like Envoy is
// an approved platform data-plane (l4 / EG / system-server proxy), not business oes.
func isAllowedPlatformEnvoy(pod corev1.Pod, c corev1.Container) bool {
	app := ""
	if pod.Labels != nil {
		app = pod.Labels["app"]
		if pod.Labels["envoy.olares.io/allowlist"] == "true" {
			return true
		}
	}
	name := strings.ToLower(c.Name)

	switch {
	case app == allowedEnvoyAppL4 || strings.Contains(pod.Name, allowedEnvoyAppL4):
		return true
	case app == allowedEnvoyAppEGData || pod.Namespace == "os-gateway":
		return true
	case app == allowedEnvoyAppSystemServer && name == allowedEnvoyContainerSystemProxy:
		return true
	case name == "olares-envoy-sidecar" || name == "olares-sidecar-init":
		return false
	}

	img := strings.ToLower(c.Image)
	if strings.Contains(img, "beclab/envoy") || strings.Contains(img, "/envoy:") {
		return false
	}
	return true
}

// isBusinessOESContainer reports a forbidden business-pod Envoy sidecar/init.
func isBusinessOESContainer(pod corev1.Pod, c corev1.Container) bool {
	if isAllowedPlatformEnvoy(pod, c) {
		return false
	}
	name := c.Name
	img := strings.ToLower(c.Image)
	if name == "olares-envoy-sidecar" || name == "olares-sidecar-init" {
		return true
	}
	if strings.Contains(img, "beclab/envoy") {
		return true
	}
	return false
}
