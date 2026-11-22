package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// pinnedLinkerdEdge is the control-plane/proxy tag shipped in cluster deploy manifests.
const pinnedLinkerdEdge = "edge-26.9.1"

func TestLinkerdControlPlanePinsEdge2691(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	manifest := filepath.Join(root, ".olares/config/cluster/deploy/linkerd-control-plane.yaml")
	b, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("read linkerd-control-plane.yaml: %v", err)
	}
	body := string(b)
	if strings.Contains(body, "edge-26.5.1") {
		t.Fatal("linkerd-control-plane.yaml still references vulnerable edge-26.5.1")
	}
	if !strings.Contains(body, "linkerdVersion: "+pinnedLinkerdEdge) {
		t.Fatalf("expected linkerdVersion %s in control-plane manifest", pinnedLinkerdEdge)
	}
	if !strings.Contains(body, "cr.l5d.io/linkerd/proxy:"+pinnedLinkerdEdge) {
		t.Fatalf("expected proxy image tag %s", pinnedLinkerdEdge)
	}
}
