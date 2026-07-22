package meshinagent

import (
	"strings"
	"testing"
)

func TestRenderNginxConfContainsListenAndJWT(t *testing.T) {
	got := RenderNginxConf(NginxConfInput{FailClosed: true})
	for _, want := range []string{
		"listen 15443",
		JWTSecretMountPath + "/token",
		"app-gateway-data.app-gateway.svc",
		"ssl_preread",
		"fail-closed",
		"js_set $mesh_in_jwt",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderNginxConf missing %q in:\n%s", want, got)
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
	if len(c.Ports) != 1 || c.Ports[0].ContainerPort != 15443 {
		t.Fatalf("listen port = %#v, want 15443", c.Ports)
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
}

func TestInitContainerSpec(t *testing.T) {
	c := InitContainerSpec()
	if c.Name != InitContainerName {
		t.Fatalf("name = %q", c.Name)
	}
	if c.SecurityContext == nil || c.SecurityContext.Capabilities == nil {
		t.Fatal("expected NET_ADMIN capabilities")
	}
}
