package egressagent

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
)

func TestShouldInjectEgressAgentProviderPod(t *testing.T) {
	perms := []appcfg.ProviderPermission{{AppName: "system-server", ProviderName: "api"}}
	if !ShouldInject(false, perms) {
		t.Fatal("T-EGR-1: provider pod should receive egress agent")
	}
}

func TestShouldInjectSkipsWithoutProviderPermission(t *testing.T) {
	if ShouldInject(false, nil) {
		t.Fatal("T-EGR-2: pod without provider permission must not receive egress agent")
	}
}

func TestShouldInjectSkipsSharedApp(t *testing.T) {
	perms := []appcfg.ProviderPermission{{AppName: "shared", ProviderName: "api"}}
	if ShouldInject(true, perms) {
		t.Fatal("shared inbound-only app must not receive egress agent")
	}
}

func TestContainerSpecFailClosedAndRoutes(t *testing.T) {
	c := ContainerSpec()
	if c.Name != ContainerName {
		t.Fatalf("name = %q", c.Name)
	}
	if c.Ports[0].ContainerPort != ListenPort {
		t.Fatalf("listen port = %d, want %d", c.Ports[0].ContainerPort, ListenPort)
	}
	foundFailClosed := false
	foundHost := false
	for _, env := range c.Env {
		switch env.Name {
		case FailClosedEnv:
			if env.Value == "true" {
				foundFailClosed = true
			}
		case "EGRESS_SYSTEM_SERVER_HOST":
			if env.Value == "system-server.user-system" {
				foundHost = true
			}
		}
	}
	if !foundFailClosed {
		t.Fatalf("missing %s=true", FailClosedEnv)
	}
	if !foundHost {
		t.Fatal("T-EGR-3: missing system-server route env")
	}
}

func TestSATokenVolumeProjected(t *testing.T) {
	v := SATokenVolume()
	if v.Projected == nil || len(v.Projected.Sources) != 1 || v.Projected.Sources[0].ServiceAccountToken == nil {
		t.Fatalf("volume = %#v", v)
	}
}
