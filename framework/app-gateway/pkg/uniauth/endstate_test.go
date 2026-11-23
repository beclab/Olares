package uniauth

import "testing"

func TestEndStateImpliesAll(t *testing.T) {
	t.Setenv(EndStateEnv, "")
	if EndStateEnabled() || EntryMTLSDesired("") || AccessPolicyReconcileDesired("") {
		t.Fatal("default off")
	}
	t.Setenv(EndStateEnv, "true")
	if !EndStateEnabled() {
		t.Fatal("end-state")
	}
	if !EntryMTLSDesired("") || !AccessPolicyReconcileDesired("") {
		t.Fatal("end-state must imply entry PKI and policy reconcile")
	}
	if !PolicyCompilationDesired("", false) {
		t.Fatal("end-state must imply policy compilation")
	}
}
