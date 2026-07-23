package meshinagent

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
)

func TestRenderNginxConfContainsListenAndJWT(t *testing.T) {
	got := RenderNginxConf(NginxConfInput{FailClosed: true})
	for _, want := range []string{
		"listen 15080",
		JWTSecretMountPath + "/token",
		"app-gateway-data.app-gateway.svc",
		"fail-closed",
		"js_set $mesh_in_jwt",
		"Authorization",
		"load_module /usr/lib/nginx/modules/ngx_http_js_module.so",
		`if ($mesh_in_jwt = "") { return 401; }`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderNginxConf missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "ssl_preread") {
		t.Fatal("HTTP inject path must not rely on ssl_preread stream passthrough")
	}
}

func TestBearerJSReadsTokenPath(t *testing.T) {
	got := BearerJS()
	if !strings.Contains(got, JWTSecretMountPath+"/token") {
		t.Fatalf("BearerJS missing token path:\n%s", got)
	}
	if !strings.Contains(got, "readJWT") {
		t.Fatal("BearerJS missing readJWT")
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
	if len(c.Ports) != 1 || c.Ports[0].ContainerPort != HTTPListenPort {
		t.Fatalf("listen port = %#v, want %d", c.Ports, HTTPListenPort)
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
}

func TestInitContainerSpec(t *testing.T) {
	c := InitContainerSpec()
	if c.Name != InitContainerName {
		t.Fatalf("name = %q", c.Name)
	}
	if c.SecurityContext == nil || c.SecurityContext.Capabilities == nil {
		t.Fatal("expected NET_ADMIN capabilities")
	}
	script := strings.Join(c.Command, " ")
	for _, want := range []string{"iptables", "-I OUTPUT", "--dport 80", "REDIRECT", "15080"} {
		if !strings.Contains(script, want) {
			t.Fatalf("init script missing %q in %#v", want, c.Command)
		}
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
