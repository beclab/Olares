package meshinagent

import (
	"fmt"
	"strings"
)

const (
	// HTTPListenPort receives redirected outbound TCP/80 toward the shared gateway.
	// Use 16xxx to stay clear of the envoy sidecar band (15000-15008, listener_image 15080)
	// and d2 loopback (15090) / stream (15443).
	HTTPListenPort = 16080
	// HTTPSListenPort receives redirected outbound TCP/443 (SNI preread / CT-1).
	HTTPSListenPort = 16443
	// HTTPSTerminatePort is the loopback HTTPS server after SNI allowlist match.
	HTTPSTerminatePort = 16444
	// HTTPSRejectPort closes non-allowlist SNI (fail-closed; no TPROXY passthrough).
	HTTPSRejectPort = 16445

	// SharedHostsFileName is the ConfigMap data key / projected file (N6 line format).
	SharedHostsFileName = "shared-hosts.txt"
)

// NginxConfInput feeds RenderNginxConf.
type NginxConfInput struct {
	HTTPListenPort     int
	HTTPSListenPort    int
	HTTPSTerminatePort int
	HTTPSRejectPort    int
	GatewayHost        string
	GatewayHTTPPort    int
	JWTTokenPath       string
	CertDir            string
	SNIMapInclude      string
	FailClosed         bool
	// EnableHTTPS includes CT-1 stream + ssl terminate. When false, HTTP JWT only
	// (used when tls Secret is absent so nginx can still start).
	EnableHTTPS bool
}

// RenderNginxConf builds nginx+njs config for:
//   - HTTP :16080 JWT inject → gateway:80 (Linkerd mTLS on the wire)
//   - when EnableHTTPS: stream :16443 ssl_preread → :16444 TLS terminate + JWT → gateway:80
//   - non-allowlist SNI → :16445 reject
func RenderNginxConf(in NginxConfInput) string {
	if in.HTTPListenPort <= 0 {
		in.HTTPListenPort = HTTPListenPort
	}
	if in.HTTPSListenPort <= 0 {
		in.HTTPSListenPort = HTTPSListenPort
	}
	if in.HTTPSTerminatePort <= 0 {
		in.HTTPSTerminatePort = HTTPSTerminatePort
	}
	if in.HTTPSRejectPort <= 0 {
		in.HTTPSRejectPort = HTTPSRejectPort
	}
	if in.GatewayHost == "" {
		in.GatewayHost = DefaultGatewayHost
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
	if in.SNIMapInclude == "" {
		in.SNIMapInclude = "/tmp/mesh-in/sni_map.conf"
	}

	failClosedNote := "# fail-closed: empty jwt returns 401; empty SNI allowlist rejects HTTPS"
	if !in.FailClosed {
		failClosedNote = "# fail-open (dev only): empty jwt still proxied"
	}

	locationBody := func() string {
		var lb strings.Builder
		lb.WriteString("      js_set $mesh_in_jwt main.readJWT;\n")
		if in.FailClosed {
			lb.WriteString("      if ($mesh_in_jwt = \"\") { return 401; }\n")
		}
		lb.WriteString("      proxy_http_version 1.1;\n")
		lb.WriteString("      proxy_set_header Host $host;\n")
		lb.WriteString("      proxy_set_header X-Forwarded-Proto $scheme;\n")
		lb.WriteString("      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		lb.WriteString("      proxy_set_header Authorization \"Bearer $mesh_in_jwt\";\n")
		lb.WriteString("      proxy_pass_request_headers on;\n")
		lb.WriteString(fmt.Sprintf("      proxy_pass http://%s:%d;\n", in.GatewayHost, in.GatewayHTTPPort))
		return lb.String()
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
	b.WriteString(fmt.Sprintf("  # certs: %s\n", in.CertDir))
	// HTTP JWT inject (dport 80 redirect).
	b.WriteString("  server {\n")
	b.WriteString(fmt.Sprintf("    listen %d;\n", in.HTTPListenPort))
	b.WriteString("    server_name _;\n")
	b.WriteString("    location / {\n")
	b.WriteString(locationBody())
	b.WriteString("    }\n")
	b.WriteString("  }\n")
	if in.EnableHTTPS {
		// HTTPS CT-1 terminate then JWT → plaintext gateway:80.
		b.WriteString("  server {\n")
		b.WriteString(fmt.Sprintf("    listen %d ssl;\n", in.HTTPSTerminatePort))
		b.WriteString("    server_name _;\n")
		b.WriteString(fmt.Sprintf("    ssl_certificate %s/tls.crt;\n", in.CertDir))
		b.WriteString(fmt.Sprintf("    ssl_certificate_key %s/tls.key;\n", in.CertDir))
		b.WriteString("    ssl_protocols TLSv1.2 TLSv1.3;\n")
		b.WriteString("    location / {\n")
		b.WriteString(locationBody())
		b.WriteString("    }\n")
		b.WriteString("  }\n")
	}
	b.WriteString("}\n")
	if in.EnableHTTPS {
		b.WriteString("stream {\n")
		b.WriteString("  map $ssl_preread_server_name $mesh_in_backend {\n")
		b.WriteString(fmt.Sprintf("    include %s;\n", in.SNIMapInclude))
		b.WriteString(fmt.Sprintf("    default 127.0.0.1:%d;\n", in.HTTPSRejectPort))
		b.WriteString("  }\n")
		b.WriteString("  server {\n")
		b.WriteString(fmt.Sprintf("    listen %d;\n", in.HTTPSListenPort))
		b.WriteString("    ssl_preread on;\n")
		b.WriteString("    proxy_pass $mesh_in_backend;\n")
		b.WriteString("  }\n")
		b.WriteString("  server {\n")
		b.WriteString(fmt.Sprintf("    listen %d;\n", in.HTTPSRejectPort))
		b.WriteString("    return;\n")
		b.WriteString("  }\n")
		b.WriteString("}\n")
	}
	return b.String()
}
