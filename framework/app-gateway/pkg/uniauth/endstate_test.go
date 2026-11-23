package uniauth

import "testing"

func TestFixedControlPlanePosture(t *testing.T) {
	if !EntryMTLSDesired() {
		t.Fatal("control-plane posture must always be on")
	}
}
