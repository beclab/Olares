package egressagent

import (
	"fmt"
	"strings"
)

// EgressRoute describes one provider domain/path → system-server upstream.
type EgressRoute struct {
	Domain       string
	Paths        []string
	UpstreamHost string
}

// RenderEgressNginxConf builds nginx http{} config for SA Bearer inject (E-1～E-4).
func RenderEgressNginxConf(saTokenPath string, routes []EgressRoute) string {
	if saTokenPath == "" {
		saTokenPath = SATokenMountPath + "/token"
	}
	var b strings.Builder
	b.WriteString("worker_processes 1;\n")
	b.WriteString("events { worker_connections 1024; }\n")
	b.WriteString("http {\n")
	b.WriteString(fmt.Sprintf("  # SA token: %s\n", saTokenPath))
	b.WriteString("  # fail-closed when token missing (EGRESS_SA_TOKEN_MISSING)\n")
	b.WriteString(fmt.Sprintf("  server {\n    listen %d;\n", ListenPort))
	if len(routes) == 0 {
		b.WriteString("    location / {\n")
		b.WriteString("      return 502;\n")
		b.WriteString("    }\n")
	}
	for _, r := range routes {
		host := r.UpstreamHost
		if host == "" {
			host = "system-server.user-system.svc:28080"
		}
		paths := r.Paths
		if len(paths) == 0 {
			paths = []string{"/"}
		}
		for _, p := range paths {
			p = strings.TrimSuffix(p, "*")
			if p == "" {
				p = "/"
			}
			b.WriteString(fmt.Sprintf("    location %s {\n", p))
			if r.Domain != "" {
				b.WriteString(fmt.Sprintf("      # domain match: %s\n", r.Domain))
			}
			b.WriteString("      proxy_set_header Temp-Authorization $http_authorization;\n")
			b.WriteString(fmt.Sprintf("      proxy_set_header Authorization \"Bearer `cat %s`\";\n", saTokenPath))
			b.WriteString(fmt.Sprintf("      proxy_pass http://%s;\n", host))
			b.WriteString("    }\n")
		}
	}
	b.WriteString("  }\n}\n")
	return b.String()
}
