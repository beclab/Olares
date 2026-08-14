package clusterop

import "testing"

func TestSignatureRequirementRegistryRequiresRegisteredType(t *testing.T) {
	reg := NewSignatureRequirementRegistry()
	if err := reg.Register(Type("example")); err != nil {
		t.Fatal(err)
	}
	if !reg.Requires(Type("example")) {
		t.Fatal("registered type was not required")
	}
	if reg.Requires(Type("other")) {
		t.Fatal("unregistered type was required")
	}
}

func TestSignatureRequirementRegistryTrimsType(t *testing.T) {
	reg := NewSignatureRequirementRegistry()
	if err := reg.Register(Type("  example  ")); err != nil {
		t.Fatal(err)
	}
	if !reg.Requires(Type("example")) {
		t.Fatal("trimmed type was not found")
	}
	if !reg.Requires(Type("  example  ")) {
		t.Fatal("lookup did not trim")
	}
}

func TestSignatureRequirementRegistryRejectsEmptyType(t *testing.T) {
	if err := NewSignatureRequirementRegistry().Register(Type("")); err == nil {
		t.Fatal("empty type registration succeeded")
	}
	if err := NewSignatureRequirementRegistry().Register(Type("   ")); err == nil {
		t.Fatal("whitespace type registration succeeded")
	}
}

func TestSignatureRequirementRegistryRejectsDuplicateType(t *testing.T) {
	reg := NewSignatureRequirementRegistry()
	if err := reg.Register(Type("example")); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(Type("example")); err == nil {
		t.Fatal("duplicate registration succeeded")
	}
}

func TestSignatureRequirementRegistryRejectsRegistrationAfterFreeze(t *testing.T) {
	reg := NewSignatureRequirementRegistry()
	reg.Freeze()
	if err := reg.Register(Type("late")); err == nil {
		t.Fatal("registration after Freeze succeeded")
	}
}

func TestDefaultSignatureRequirementsHasBuiltInTypes(t *testing.T) {
	for _, typ := range []Type{TypeReboot, TypeShutdown, TypeSetSSHPassword} {
		if !RequiresSignature(typ) {
			t.Errorf("RequiresSignature(%q) = false, want true for built-in module", typ)
		}
	}
}
