package calleragent

import (
	"fmt"
	"strings"
)

// NginxConfInput feeds RenderNginxConf.
type NginxConfInput struct {
	ListenPort      int
	GatewayHost     string
	GatewayHTTPPort int
	JWTTokenPath    string
	CertDir         string
	SharedHostsFile string
	FailClosed      bool
}

// RenderNginxConf builds a minimal nginx+njs config for HTTP Bearer inject and
// HTTPS CT-1 termination on ListenPort (default 15443).
func RenderNginxConf(in NginxConfInput) string {
	if in.ListenPort <= 0 {
		in.ListenPort = listenPort
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
	if in.CertDir == "" {
		in.CertDir = CertsMountPath
	}
	if in.SharedHostsFile == "" {
		in.SharedHostsFile = HostsMountPath + "/hosts"
	}

	failClosedNote := "# fail-closed: missing jwt refuses upstream"
	if !in.FailClosed {
		failClosedNote = "# fail-open (dev only)"
	}

	var b strings.Builder
	b.WriteString("worker_processes 1;\n")
	b.WriteString("events { worker_connections 1024; }\n")
	b.WriteString(failClosedNote + "\n")
	b.WriteString("http {\n")
	b.WriteString("  js_import main from bearer.js;\n")
	b.WriteString(fmt.Sprintf("  # jwt path: %s\n", in.JWTTokenPath))
	b.WriteString(fmt.Sprintf("  # shared hosts: %s\n", in.SharedHostsFile))
	b.WriteString("  server {\n")
	b.WriteString("    listen 15080;\n")
	b.WriteString("    location / {\n")
	b.WriteString("      js_set $caller_jwt main.readJWT;\n")
	b.WriteString("      proxy_set_header Authorization \"Bearer $caller_jwt\";\n")
	b.WriteString(fmt.Sprintf("      proxy_pass http://%s:%d;\n", in.GatewayHost, in.GatewayHTTPPort))
	b.WriteString("    }\n")
	b.WriteString("  }\n")
	b.WriteString("}\n")
	b.WriteString("stream {\n")
	b.WriteString(fmt.Sprintf("  # CT-1 certs: %s\n", in.CertDir))
	b.WriteString(fmt.Sprintf("  server { listen %d; ssl_preread on; proxy_pass %s:%d; }\n",
		in.ListenPort, in.GatewayHost, in.GatewayHTTPPort))
	b.WriteString("}\n")
	return b.String()
}
