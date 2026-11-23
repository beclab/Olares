package uniauth

import (
	"os"
	"testing"
)

func TestEndStateDefaultsOn(t *testing.T) {
	t.Cleanup(func() { _ = os.Unsetenv(EndStateEnv) })
	_ = os.Unsetenv(EndStateEnv)
	if !EndStateEnabled() {
		t.Fatal("default on")
	}
	if !EntryMTLSDesired("") || !AccessPolicyReconcileDesired("") || !PolicyCompilationDesired("", false) {
		t.Fatal("default end-state implies control-plane features")
	}
}

func TestEndStateExplicitOff(t *testing.T) {
	t.Setenv(EndStateEnv, "false")
	if EndStateEnabled() || EntryMTLSDesired("") {
		t.Fatal("explicit off")
	}
}
