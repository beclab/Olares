package apiserver

import "testing"

func TestSharedAppNeedsSidecarPatch(t *testing.T) {
	if sharedAppNeedsSidecarPatch(false, false) {
		t.Fatal("pure Shared callee must stay on label-only path")
	}
	if !sharedAppNeedsSidecarPatch(true, false) {
		t.Fatal("Shared caller with mesh-in must use CreatePatch")
	}
	if !sharedAppNeedsSidecarPatch(false, true) {
		t.Fatal("Shared with mesh-out must use CreatePatch")
	}
}
