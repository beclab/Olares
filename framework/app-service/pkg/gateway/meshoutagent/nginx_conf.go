package meshoutagent

import (
	"fmt"
	"strings"
)

// MeshOutRoute describes one provider domain/path → system-server upstream.
type MeshOutRoute struct {
	Domain       string
	Paths        []string
	UpstreamHost string
}

// SATokenPlaceholder is substituted at container start with the projected token.
const SATokenPlaceholder = "__SA_TOKEN__"

// RenderMeshOutNginxConf builds nginx http{} config for SA Bearer inject.
// Authorization uses a placeholder replaced by the sidecar entrypoint from the
// projected ServiceAccount token file — never a shell backtick literal.
func RenderMeshOutNginxConf(saTokenPath string, routes []MeshOutRoute) string {
	if saTokenPath == "" {
		saTokenPath = SATokenMountPath + "/token"
	}
	var b strings.Builder
	b.WriteString("worker_processes 1;\n")
	b.WriteString("error_log /dev/stderr warn;\n")
	b.WriteString("events { worker_connections 1024; }\n")
	b.WriteString("http {\n")
	b.WriteString(fmt.Sprintf("  # SA token file (runtime substitute): %s\n", saTokenPath))
	b.WriteString("  # fail-closed when token missing (MESH_OUT_SA_TOKEN_MISSING)\n")
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
			b.WriteString(fmt.Sprintf("      proxy_set_header Authorization \"Bearer %s\";\n", SATokenPlaceholder))
			b.WriteString(fmt.Sprintf("      proxy_pass http://%s;\n", host))
			b.WriteString("    }\n")
		}
	}
	b.WriteString("  }\n}\n")
	return b.String()
}

// RenderMeshOutEntrypoint substitutes the projected token into nginx.conf and starts nginx.
func RenderMeshOutEntrypoint(saTokenPath, confPath string) string {
	if saTokenPath == "" {
		saTokenPath = SATokenMountPath + "/token"
	}
	if confPath == "" {
		confPath = ConfMountPath + "/nginx.conf"
	}
	return fmt.Sprintf(`set -eu
TOKEN_FILE="%s"
CONF="%s"
WORK="/tmp/mesh-out-nginx.conf"
if [ ! -s "$TOKEN_FILE" ]; then
  echo "MESH_OUT_SA_TOKEN_MISSING" >&2
  exit 1
fi
TOKEN="$(cat "$TOKEN_FILE")"
if [ -z "$TOKEN" ]; then
  echo "MESH_OUT_SA_TOKEN_MISSING" >&2
  exit 1
fi
sed "s|%s|${TOKEN}|g" "$CONF" > "$WORK"
exec nginx -c "$WORK" -g "daemon off;"
`, saTokenPath, confPath, SATokenPlaceholder)
}
