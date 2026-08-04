package meshoutagent

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

func TestRenderMeshOutNginxConf(t *testing.T) {
	got := RenderMeshOutNginxConf("", []MeshOutRoute{{
		Domain:       "provider.example",
		Paths:        []string{"/api/*"},
		UpstreamHost: "system-server.user-system-alice:28080",
	}})
	for _, want := range []string{
		"listen 15001",
		SATokenMountPath + "/token",
		"Temp-Authorization",
		"system-server.user-system-alice:28080",
		"MESH_OUT_SA_TOKEN_MISSING",
		"location /api",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestContainerSpecNonStub(t *testing.T) {
	c := ContainerSpec()
	if IsStubImage(c.Image) {
		t.Fatalf("stub image forbidden: %s", c.Image)
	}
	if c.Image != DefaultImage {
		t.Fatalf("image = %q want %q", c.Image, DefaultImage)
	}
	if len(c.Ports) != 1 || c.Ports[0].ContainerPort != ListenPort {
		t.Fatalf("ports = %#v", c.Ports)
	}
	if c.SecurityContext == nil || c.SecurityContext.RunAsUser == nil || *c.SecurityContext.RunAsUser != 1652 {
		t.Fatalf("runAsUser = %#v, want 1652", c.SecurityContext)
	}
}

func TestShouldInject(t *testing.T) {
	if ShouldInject(true, []appcfg.ProviderPermission{{AppName: "x"}}) {
		t.Fatal("shared must not inject")
	}
	if ShouldInject(false, nil) {
		t.Fatal("no provider must not inject")
	}
	if !ShouldInject(false, []appcfg.ProviderPermission{{AppName: "x"}}) {
		t.Fatal("provider must inject")
	}
}

func TestInitContainerSpec(t *testing.T) {
	c := InitContainerSpec()
	if c.Name != InitContainerName {
		t.Fatalf("name = %q", c.Name)
	}
	if c.SecurityContext == nil || c.SecurityContext.RunAsUser == nil || *c.SecurityContext.RunAsUser != 0 {
		t.Fatal("iptables init must run as root")
	}
	script := strings.Join(c.Command, " ")
	for _, want := range []string{`NGINX_UID="1652"`, "REDIRECT", "--dport", "15001", "ENVOY_UID"} {
		if !strings.Contains(script, want) {
			t.Fatalf("init script missing %q", want)
		}
	}
}
