package meshinagent

import (
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

const (
	// HTTPListenPort receives redirected outbound TCP/80 toward the shared gateway.
	HTTPListenPort = 16080
	// HTTPSListenPort receives redirected outbound TCP/443 (SNI preread / CT-1).
	HTTPSListenPort = 16443
	// HTTPSTerminatePort is the loopback HTTPS server after SNI allowlist match.
	HTTPSTerminatePort = 16444

	// SharedHostsFileName is the ConfigMap data key / projected file (N6 line format).
	SharedHostsFileName = "shared-hosts.txt"
)

// NginxConfInput feeds RenderNginxConf.
type NginxConfInput struct {
	HTTPListenPort     int
	HTTPSListenPort    int
	HTTPSTerminatePort int
	GatewayHost        string
	GatewayHTTPPort    int
	JWTTokenPath       string
	CertDir            string
	HostsFile          string
	PlatformDomain     string
	FailClosed         bool
	// EnableHTTPS includes CT-1 stream + ssl terminate. When false, HTTP JWT only.
	EnableHTTPS bool
}

// RenderNginxConf builds nginx+njs config for:
//   - HTTP :16080 JWT inject → gateway:80
//   - when EnableHTTPS: stream :16443 ssl_preread + decideOffload
//     (allowlist → :16444 terminate+JWT; else SNI:443 passthrough)
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
	if in.HostsFile == "" {
		in.HostsFile = HostsMountPath + "/" + SharedHostsFileName
	}
	if in.PlatformDomain == "" {
		in.PlatformDomain = "olares.com"
	}

	failClosedNote := "# fail-closed: empty jwt returns 401"
	if !in.FailClosed {
		failClosedNote = "# fail-open (dev only): empty jwt still proxied"
	}

	locationBody := func(https bool) string {
		var lb strings.Builder
		lb.WriteString("      js_set $mesh_in_jwt main.readJWT;\n")
		if in.FailClosed {
			lb.WriteString("      if ($mesh_in_jwt = \"\") { return 401; }\n")
		}
		if https {
			lb.WriteString("      js_set $mesh_in_host_ok main.checkHost;\n")
			lb.WriteString("      if ($mesh_in_host_ok = \"0\") { return 421; }\n")
		}
		lb.WriteString("      proxy_http_version 1.1;\n")
		lb.WriteString("      proxy_buffering off;\n")
		lb.WriteString(fmt.Sprintf("      proxy_read_timeout %s;\n", constants.MeshInProxyReadTimeout))
		lb.WriteString(fmt.Sprintf("      proxy_send_timeout %s;\n", constants.MeshInProxySendTimeout))
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
	if in.EnableHTTPS {
		b.WriteString("load_module /usr/lib/nginx/modules/ngx_stream_js_module.so;\n")
	}
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
	b.WriteString(fmt.Sprintf("  # hosts: %s\n", in.HostsFile))
	b.WriteString("  map $ssl_server_name $tls_cert_path {\n")
	b.WriteString(fmt.Sprintf("    default %s/tls.crt;\n", in.CertDir))
	b.WriteString("  }\n")
	b.WriteString("  map $ssl_server_name $tls_key_path {\n")
	b.WriteString(fmt.Sprintf("    default %s/tls.key;\n", in.CertDir))
	b.WriteString("  }\n")
	b.WriteString("  server {\n")
	b.WriteString(fmt.Sprintf("    listen %d;\n", in.HTTPListenPort))
	b.WriteString("    server_name _;\n")
	b.WriteString("    location / {\n")
	b.WriteString(locationBody(false))
	b.WriteString("    }\n")
	b.WriteString("  }\n")
	if in.EnableHTTPS {
		b.WriteString("  server {\n")
		b.WriteString(fmt.Sprintf("    listen %d ssl;\n", in.HTTPSTerminatePort))
		b.WriteString("    server_name _;\n")
		b.WriteString("    ssl_certificate $tls_cert_path;\n")
		b.WriteString("    ssl_certificate_key $tls_key_path;\n")
		b.WriteString(fmt.Sprintf("    ssl_certificate_cache max=%d inactive=%s valid=%s;\n",
			constants.MeshInCertCacheMax, constants.MeshInCertCacheInactive, constants.MeshInCertCacheValid))
		b.WriteString("    ssl_protocols TLSv1.2 TLSv1.3;\n")
		b.WriteString("    location / {\n")
		b.WriteString(locationBody(true))
		b.WriteString("    }\n")
		b.WriteString("  }\n")
	}
	b.WriteString("}\n")
	if in.EnableHTTPS {
		b.WriteString("stream {\n")
		b.WriteString("  js_import stream_main from /tmp/mesh-in/bearer.js;\n")
		b.WriteString("  js_set $mesh_in_upstream stream_main.decideOffload;\n")
		// Olares / modern k8s name CoreDNS as Service "coredns" (not legacy "kube-dns").
		b.WriteString("  resolver coredns.kube-system.svc.cluster.local ipv6=off valid=30s;\n")
		b.WriteString("  server {\n")
		b.WriteString(fmt.Sprintf("    listen %d;\n", in.HTTPSListenPort))
		b.WriteString("    ssl_preread on;\n")
		b.WriteString("    proxy_pass $mesh_in_upstream;\n")
		b.WriteString("    proxy_connect_timeout 5s;\n")
		b.WriteString("    proxy_timeout 3600s;\n")
		b.WriteString("  }\n")
		b.WriteString("}\n")
	}
	return b.String()
}
