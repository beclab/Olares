package routecontrol

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// callerUserSystemNamespace resolves user-system-<viewer> for middleware egress.
func callerUserSystemNamespace(ctx context.Context, c client.Client, callerNS string) string {
	if c == nil || callerNS == "" {
		return ""
	}
	var ns corev1.Namespace
	if err := c.Get(ctx, client.ObjectKey{Name: callerNS}, &ns); err == nil {
		if owner := strings.TrimSpace(ns.Labels["bytetrade.io/ns-owner"]); owner != "" {
			return "user-system-" + owner
		}
	}
	// Fallback: litellm-brucedai -> user-system-brucedai
	if i := strings.LastIndex(callerNS, "-"); i > 0 {
		return "user-system-" + callerNS[i+1:]
	}
	return ""
}
