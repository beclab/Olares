package meshinagent

import (
	"fmt"
	"strings"
)

const (
	// HTTPListenPort receives redirected outbound TCP/80 toward the shared gateway.
	HTTPListenPort = 15080
)

// NginxConfInput feeds RenderNginxConf.
type NginxConfInput struct {
	HTTPListenPort  int
	GatewayHost     string
	GatewayHTTPPort int
	JWTTokenPath    string
	FailClosed      bool
}

// RenderNginxConf builds a runnable nginx+njs config that injects
// Authorization: Bearer <caller JWT> on HTTP traffic to the shared gateway.
func RenderNginxConf(in NginxConfInput) string {
	if in.HTTPListenPort <= 0 {
		in.HTTPListenPort = HTTPListenPort
	}
	if in.GatewayHost == "" {
		in.GatewayHost = "app-gateway-data.app-gateway.svc"
	}
	if in.GatewayHTTPPort <= 0 {
		in.GatewayHTTPPort = 80
	}
	if in.JWTTokenPath == "" {
		in.JWTTokenPath = JWTSecretMountPath + "/token"
	}

	failClosedNote := "# fail-closed: empty jwt returns 401"
	if !in.FailClosed {
		failClosedNote = "# fail-open (dev only): empty jwt still proxied"
	}

	var b strings.Builder
	b.WriteString("load_module /usr/lib/nginx/modules/ngx_http_js_module.so;\n")
	b.WriteString("worker_processes 1;\n")
	b.WriteString("error_log /var/log/nginx/error.log warn;\n")
	b.WriteString("pid /tmp/nginx-mesh-in.pid;\n")
	b.WriteString("events { worker_connections 1024; }\n")
	b.WriteString(failClosedNote + "\n")
	b.WriteString("http {\n")
	b.WriteString("  access_log off;\n")
	b.WriteString("  js_import main from /tmp/mesh-in/bearer.js;\n")
	b.WriteString(fmt.Sprintf("  # jwt path: %s\n", in.JWTTokenPath))
	b.WriteString("  server {\n")
	b.WriteString(fmt.Sprintf("    listen %d;\n", in.HTTPListenPort))
	b.WriteString("    server_name _;\n")
	b.WriteString("    location / {\n")
	b.WriteString("      js_set $mesh_in_jwt main.readJWT;\n")
	if in.FailClosed {
		b.WriteString("      if ($mesh_in_jwt = \"\") { return 401; }\n")
	}
	b.WriteString("      proxy_http_version 1.1;\n")
	b.WriteString("      proxy_set_header Host $host;\n")
	b.WriteString("      proxy_set_header X-Forwarded-Proto $scheme;\n")
	b.WriteString("      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
	b.WriteString("      proxy_set_header Authorization \"Bearer $mesh_in_jwt\";\n")
	b.WriteString("      proxy_pass_request_headers on;\n")
	b.WriteString(fmt.Sprintf("      proxy_pass http://%s:%d;\n", in.GatewayHost, in.GatewayHTTPPort))
	b.WriteString("    }\n")
	b.WriteString("  }\n")
	b.WriteString("}\n")
	return b.String()
}
