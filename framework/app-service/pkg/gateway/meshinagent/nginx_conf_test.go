package meshinagent

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
)

func TestRenderNginxConfContainsListenAndJWT(t *testing.T) {
	got := RenderNginxConf(NginxConfInput{FailClosed: true, EnableHTTPS: true})
	for _, want := range []string{
		"listen 16080",
		"listen 16443",
		"listen 16444 ssl",
		"ssl_preread on",
		"resolver coredns.kube-system.svc.cluster.local",
		"decideOffload",
		"ssl_certificate_cache",
		"proxy_buffering off",
		"proxy_read_timeout 600s",
		"return 421",
		JWTSecretMountPath + "/token",
		"app-gateway-data.os-gateway.svc",
		"proxy_pass http://app-gateway-data.os-gateway.svc:80",
		"fail-closed",
		"js_set $mesh_in_host_ok",
		"js_set $mesh_in_caller_jwt",
		"js_set $mesh_in_auth_deny",
		"main.callerJwt",
		"main.authDeny",
		"main.checkHost",
		callerjwt.CallerJWTHeaderName,
		"load_module /usr/lib/nginx/modules/ngx_http_js_module.so",
		"load_module /usr/lib/nginx/modules/ngx_stream_js_module.so",
		`if ($mesh_in_auth_deny = "1") { return 401; }`,
		"ssl_certificate $tls_cert_path",
		"js_set $tls_cert_path main.pickCert",
		"js_set $mesh_in_tls_mode main.tlsMode",
		"@mesh_in_placeholder",
		"X-Olares-Mesh-In-TLS",
		"mesh_in_tls_placeholder",
		"stream {",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderNginxConf missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, `proxy_set_header Authorization`) {
		t.Fatal("must not overwrite Authorization; app credentials must pass through")
	}
	if strings.Contains(got, `if ($mesh_in_jwt = "") { return 401; }`) {
		t.Fatal("fail-closed must use authDeny (allowlist-gated), not bare empty jwt")
	}
	if strings.Count(got, "js_set $mesh_in_host_ok") < 2 {
		t.Fatal("HTTP and HTTPS terminate locations must both checkHost")
	}
	if strings.Contains(got, "sni_map.conf") {
		t.Fatal("static sni_map must not be used; njs hosts hot-read replaces it")
	}
	if strings.Count(got, "proxy_pass http://app-gateway-data.os-gateway.svc:80") < 2 {
		t.Fatal("expected HTTP and HTTPS terminate servers both proxy to gateway:80")
	}
}

func TestRenderNginxConfHTTPOnlyOmitsSSL(t *testing.T) {
	got := RenderNginxConf(NginxConfInput{FailClosed: true, EnableHTTPS: false})
	if !strings.Contains(got, "listen 16080") {
		t.Fatal("HTTP listen missing")
	}
	for _, ban := range []string{"ssl_preread", "listen 16444 ssl", "ssl_certificate", "stream {"} {
		if strings.Contains(got, ban) {
			t.Fatalf("HTTP-only conf must not contain %q", ban)
		}
	}
}

func TestBearerJSDecideAndHostsHotRead(t *testing.T) {
	got := BearerJS()
	for _, want := range []string{
		JWTSecretMountPath + "/token",
		"readJWT",
		"callerJwt",
		"authDeny",
		"decideOffload",
		"reloadAuthHosts",
		"reloadTLSHosts",
		"AUTH_HOSTS_FILE",
		"TLS_HOSTS_FILE",
		"CACHE_TTL_MS",
		"passthrough",
		"checkHost",
		"jwtEligible",
		"tlsMode",
		"pickCert",
		"pickKey",
		"realCertReady",
		PlaceholderCertDir,
		HostsMountPath + "/" + SharedHostsFileName,
		HostsMountPath + "/" + TLSHostsFileName,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("BearerJS missing %q", want)
		}
	}
	if strings.Contains(got, "V2_GUARD") {
		t.Fatal("entrance class must not be inferred from .shared. domain regex")
	}
}

func TestContainerSpecNonStub(t *testing.T) {
	c := ContainerSpec()
	if IsStubImage(c.Image) {
		t.Fatalf("default image must not be stub: %s", c.Image)
	}
	if c.Image != DefaultImage {
		t.Fatalf("image = %q, want %q", c.Image, DefaultImage)
	}
	if len(c.Ports) != 2 {
		t.Fatalf("ports = %#v, want http+https", c.Ports)
	}
	ports := map[int32]bool{}
	for _, p := range c.Ports {
		ports[p.ContainerPort] = true
	}
	if !ports[HTTPListenPort] || !ports[HTTPSListenPort] {
		t.Fatalf("ports = %#v, want %d and %d", c.Ports, HTTPListenPort, HTTPSListenPort)
	}
	if c.SecurityContext == nil || c.SecurityContext.RunAsUser == nil || *c.SecurityContext.RunAsUser != 1651 {
		t.Fatalf("runAsUser = %#v, want 1651", c.SecurityContext)
	}
	foundFailClosed := false
	for _, env := range c.Env {
		if env.Name == FailClosedEnv && env.Value == "true" {
			foundFailClosed = true
		}
	}
	if !foundFailClosed {
		t.Fatalf("missing %s=true", FailClosedEnv)
	}
	cmd := strings.Join(c.Command, " ")
	if !strings.Contains(cmd, "base64 -d") || !strings.Contains(cmd, "nginx -c") {
		t.Fatalf("start command must materialize conf then exec nginx: %#v", c.Command)
	}
	if strings.Contains(cmd, "sni_map.conf") {
		t.Fatal("start script must not build static sni_map")
	}
	if !strings.Contains(cmd, "ensure-placeholder-cert") || !strings.Contains(cmd, "TLS mode=placeholder") {
		t.Fatalf("start command must ensure placeholder and always enable HTTPS: %#v", c.Command)
	}
	realCheck := `[ -s "$CERT_DIR/tls.crt" ] && [ -s "$CERT_DIR/tls.key" ]`
	if realPos, ensurePos := strings.Index(cmd, realCheck), strings.Index(cmd, "ensure-placeholder-cert -dir"); realPos < 0 || ensurePos < 0 || realPos > ensurePos {
		t.Fatalf("start command must check real certs before generating placeholder: %#v", c.Command)
	}
	if strings.Contains(cmd, "HTTP-only mode") {
		t.Fatal("HTTP-only mode must be removed")
	}
}

func TestInitContainerSpec(t *testing.T) {
	c := InitContainerSpec()
	if c.Name != InitContainerName {
		t.Fatalf("name = %q", c.Name)
	}
	if c.SecurityContext == nil || c.SecurityContext.Capabilities == nil {
		t.Fatal("expected NET_ADMIN capabilities")
	}
	if c.SecurityContext == nil || c.SecurityContext.RunAsUser == nil || *c.SecurityContext.RunAsUser != 0 {
		t.Fatal("iptables init must run as root")
	}
	script := strings.Join(c.Command, " ")
	for _, want := range []string{
		"iptables", "-I OUTPUT", "--dport 80", "REDIRECT", "16080", "os-gateway",
		"--dport 443", "16443",
		`NGINX_UID="1651"`,
		`LINKERD_UID="2102"`,
		`ENVOY_UID="1555"`,
		"! --uid-owner $NGINX_UID",
		"! --uid-owner $LINKERD_UID",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("init script missing %q in %#v", want, c.Command)
		}
	}
	if strings.Contains(script, `--uid-owner "$NGINX_UID" -j RETURN`) ||
		strings.Contains(script, `--uid-owner $NGINX_UID -j RETURN`) {
		t.Fatal("must not install blanket uid RETURN (blocks Linkerd mTLS)")
	}
}

func TestJWTSecretVolumeUsesCallerJWT(t *testing.T) {
	v := JWTSecretVolume()
	if v.Secret == nil {
		t.Fatal("expected secret volume")
	}
	if v.Secret.SecretName != callerjwt.AppJWTSecretName {
		t.Fatalf("secretName = %q, want %q", v.Secret.SecretName, callerjwt.AppJWTSecretName)
	}
	if v.Secret.Optional == nil || *v.Secret.Optional {
		t.Fatal("caller-jwt mount must be required (fail closed)")
	}
}

func TestSharedHostsVolumeProjectsAuthAndTLSHosts(t *testing.T) {
	v := SharedHostsVolume()
	if v.ConfigMap == nil {
		t.Fatal("expected configMap volume")
	}
	if v.ConfigMap.Name != constants.MeshInSharedHostsCMName {
		t.Fatalf("name = %q", v.ConfigMap.Name)
	}
	if len(v.ConfigMap.Items) != 2 {
		t.Fatalf("items = %#v, want 2 keys", v.ConfigMap.Items)
	}
	keys := map[string]bool{}
	for _, it := range v.ConfigMap.Items {
		keys[it.Key] = true
	}
	if !keys[SharedHostsFileName] || !keys[TLSHostsFileName] {
		t.Fatalf("items = %#v, want %s and %s", v.ConfigMap.Items, SharedHostsFileName, TLSHostsFileName)
	}
}

func TestCertsVolumeForViewer(t *testing.T) {
	v := CertsVolumeForViewer("Alice")
	if v.Secret == nil || v.Secret.SecretName != "olares-mesh-in-tls-alice" {
		t.Fatalf("secret = %#v", v.Secret)
	}
	v2 := CertsVolumeForViewer("")
	if v2.Secret == nil || v2.Secret.SecretName != "olares-mesh-in-certs" {
		t.Fatalf("empty viewer secret = %#v", v2.Secret)
	}
}
