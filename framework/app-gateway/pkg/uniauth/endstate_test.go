package uniauth

import "testing"

func TestFixedControlPlanePosture(t *testing.T) {
	if !PolicyCompilationDesired() || !EntryMTLSDesired() || !AccessPolicyReconcileDesired() {
		t.Fatal("control-plane posture must always be on")
	}
}
