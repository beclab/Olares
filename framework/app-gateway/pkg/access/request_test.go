package access

import "testing"

func TestValidateDecisionTransition(t *testing.T) {
	if err := ValidateDecisionTransition(AccessRequestDecision{Current: RequestPending, Next: RequestApproved}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDecisionTransition(AccessRequestDecision{Current: RequestApproved, Next: RequestDenied}); err == nil {
		t.Fatal("terminal must stick")
	}
}

func TestValidateCapabilityExpansion(t *testing.T) {
	old := []HTTPRule{{Method: "GET", Path: "/state", PathType: PathTypeExact}}
	wider := []HTTPRule{{Method: "GET", Path: "/state", PathType: PathTypePrefix}}
	if err := ValidateCapabilityExpansion(old, wider, false); err == nil {
		t.Fatal("expected re-review requirement")
	}
	if err := ValidateCapabilityExpansion(old, wider, true); err != nil {
		t.Fatal(err)
	}
}
