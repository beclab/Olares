package meshoutagent

import (
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

// RoutesFromPermissions builds mesh-out nginx routes from legacy provider permissions.
// Domains mirror envoy outbound virtual_host domains: providerName.ns and optional Domain.
// Upstream is always system-server in the caller's user-system namespace (fail-closed if unknown).
func RoutesFromPermissions(perms []appcfg.ProviderPermission, podNamespace string) []MeshOutRoute {
	if len(perms) == 0 {
		return nil
	}
	upstream := systemServerUpstream(podNamespace)
	out := make([]MeshOutRoute, 0, len(perms))
	for _, p := range perms {
		ns := strings.TrimSpace(p.Namespace)
		if ns == "" {
			ns = "user-system"
		}
		name := strings.TrimSpace(p.ProviderName)
		if name == "" {
			name = strings.TrimSpace(p.AppName)
		}
		if name == "" {
			continue
		}
		domain := fmt.Sprintf("%s.%s", name, ns)
		out = append(out, MeshOutRoute{
			Domain:       domain,
			Paths:        []string{"/"},
			UpstreamHost: upstream,
		})
	}
	if len(out) == 0 {
		// Keep a fail-closed default location so nginx still starts.
		out = append(out, MeshOutRoute{Paths: []string{"/"}, UpstreamHost: upstream})
	}
	return out
}

// RoutesFromPermissionCfg builds routes from resolved PermissionCfg (Domain + Paths).
func RoutesFromPermissionCfg(cfgs []appcfg.PermissionCfg, podNamespace string) []MeshOutRoute {
	if len(cfgs) == 0 {
		return RoutesFromPermissions(nil, podNamespace)
	}
	upstream := systemServerUpstream(podNamespace)
	out := make([]MeshOutRoute, 0, len(cfgs))
	for _, c := range cfgs {
		domain := strings.TrimSpace(c.Domain)
		paths := c.Paths
		if len(paths) == 0 {
			paths = []string{"/"}
		}
		if domain == "" && c.ProviderPermission != nil {
			ns := strings.TrimSpace(c.Namespace)
			if ns == "" {
				ns = "user-system"
			}
			name := strings.TrimSpace(c.ProviderName)
			if name == "" {
				name = strings.TrimSpace(c.AppName)
			}
			if name != "" {
				domain = fmt.Sprintf("%s.%s", name, ns)
			}
		}
		out = append(out, MeshOutRoute{
			Domain:       domain,
			Paths:        paths,
			UpstreamHost: upstream,
		})
	}
	return out
}

func systemServerUpstream(podNamespace string) string {
	owner := ownerFromPodNamespace(podNamespace)
	if owner == "" {
		return "system-server.user-system.svc:28080"
	}
	return fmt.Sprintf("system-server.user-system-%s.svc:28080", owner)
}

func ownerFromPodNamespace(ns string) string {
	ns = strings.TrimSpace(ns)
	for _, prefix := range []string{"user-space-", "user-system-"} {
		if strings.HasPrefix(ns, prefix) {
			return strings.TrimPrefix(ns, prefix)
		}
	}
	// app-<name>-<owner> style: take trailing segment after last '-' only when
	// namespace embeds user-space pattern elsewhere; fallback empty.
	return ""
}
