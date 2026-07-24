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
		"resolver kube-dns.kube-system.svc.cluster.local",
		"decideOffload",
		"ssl_certificate_cache",
		"proxy_buffering off",
		"proxy_read_timeout 600s",
		"return 421",
		JWTSecretMountPath + "/token",
		"app-gateway-data.os-gateway.svc",
		"proxy_pass http://app-gateway-data.os-gateway.svc:80",
		"fail-closed",
		"js_set $mesh_in_jwt",
		"Authorization",
		"load_module /usr/lib/nginx/modules/ngx_http_js_module.so",
		"load_module /usr/lib/nginx/modules/ngx_stream_js_module.so",
		`if ($mesh_in_jwt = "") { return 401; }`,
		"ssl_certificate",
		"stream {",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderNginxConf missing %q in:\n%s", want, got)
		}
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
		"decideOffload",
		"reloadHostsIfNeeded",
		"CACHE_TTL_MS",
		"passthrough",
		"checkHost",
		HostsMountPath + "/" + SharedHostsFileName,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("BearerJS missing %q", want)
		}
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
	if !strings.Contains(cmd, "HTTP-only mode") || !strings.Contains(cmd, "CT-1 HTTPS enabled") {
		t.Fatalf("start command must select HTTP-only vs CT-1 by cert presence: %#v", c.Command)
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

func TestSharedHostsVolumeProjectsSharedHostsTxt(t *testing.T) {
	v := SharedHostsVolume()
	if v.ConfigMap == nil {
		t.Fatal("expected configMap volume")
	}
	if v.ConfigMap.Name != constants.MeshInSharedHostsCMName {
		t.Fatalf("name = %q", v.ConfigMap.Name)
	}
	if len(v.ConfigMap.Items) != 1 || v.ConfigMap.Items[0].Key != SharedHostsFileName {
		t.Fatalf("items = %#v, want key %s", v.ConfigMap.Items, SharedHostsFileName)
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
