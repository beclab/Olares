package routecontrol

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// AppGatewayDataNamespace is the namespace hosting app-gateway-data Service.
	AppGatewayDataNamespace = "app-gateway"

	// DefaultInClusterStrongIdentityServicePort is the strong-identity
	// app-gateway-data Service port consumed by in-cluster callers.
	DefaultInClusterStrongIdentityServicePort int32 = 8081

	// DefaultInClusterHTTPServicePort is the legacy app-gateway-data HTTP service
	// port kept for backward compatibility with callers that still target :80.
	// Deprecated: caller strong HTTP path should use DefaultInClusterHTTPStrongServicePort.
	DefaultInClusterHTTPServicePort int32 = 80

	// DefaultInClusterHTTPStrongServicePort is the caller strong-HTTP service port.
	DefaultInClusterHTTPStrongServicePort int32 = 8082

	// Linkerd skip-port namespace annotation keys.
	LinkerdSkipInboundPortsAnnotation  = "config.linkerd.io/skip-inbound-ports"
	LinkerdSkipOutboundPortsAnnotation = "config.linkerd.io/skip-outbound-ports"

	// Inbound skip values for caller namespace flavors.
	OlaresEnvoyInboundSkipPorts = "15000,15001,15003,15008"
	PureCallerInboundSkipPorts  = "1-65535"
)

// MeshHijackServicePorts returns service ports that must stay mesh-hijacked.
func MeshHijackServicePorts(strongIdentityPort int32) []int32 {
	return []int32{strongIdentityPort, DefaultInClusterHTTPStrongServicePort}
}

// ComputeSkipOutboundPorts builds skip-outbound range string from hijack ports.
func ComputeSkipOutboundPorts(hijackPorts []int32) (string, error) {
	if len(hijackPorts) == 0 {
		return "", fmt.Errorf("hijack ports invalid: empty")
	}

	ports := make([]int32, len(hijackPorts))
	copy(ports, hijackPorts)
	sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })

	uniq := make([]int32, 0, len(ports))
	for i, p := range ports {
		if p < 1 || p > 65535 {
			return "", fmt.Errorf("hijack ports invalid: %v", hijackPorts)
		}
		if i > 0 && p == ports[i-1] {
			return "", fmt.Errorf("hijack ports invalid: duplicate %d", p)
		}
		uniq = append(uniq, p)
	}

	parts := make([]string, 0, len(uniq)+1)
	start := int32(1)
	for _, p := range uniq {
		if start <= p-1 {
			parts = append(parts, fmt.Sprintf("%d-%d", start, p-1))
		}
		start = p + 1
	}
	if start <= 65535 {
		parts = append(parts, fmt.Sprintf("%d-%d", start, int32(65535)))
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("hijack ports invalid: %v", hijackPorts)
	}
	return strings.Join(parts, ","), nil
}
