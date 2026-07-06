package mesh

import (
	"context"
	"testing"
)

func TestShouldSkipEnvoySidecarDefaultFalse(t *testing.T) {
	ResetLinkerdMeshEnabledForTest()
	t.Cleanup(ResetLinkerdMeshEnabledForTest)
	if ShouldSkipEnvoySidecar(context.Background()) {
		t.Fatal("expected envoy sidecar by default before Linkerd steady state")
	}
}

func TestShouldSkipEnvoySidecarWhenMeshEnabled(t *testing.T) {
	ResetLinkerdMeshEnabledForTest()
	t.Cleanup(ResetLinkerdMeshEnabledForTest)
	PrimeLinkerdMeshEnabledForTest(true)
	if !ShouldSkipEnvoySidecar(context.Background()) {
		t.Fatal("expected envoy sidecar skip when linkerd mesh enabled")
	}
}
