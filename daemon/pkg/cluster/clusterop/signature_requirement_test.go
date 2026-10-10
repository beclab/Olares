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

// Every operation this daemon can be asked to start needs the owner's
// signature.
//
// It is asked of the module registry rather than of a list written here,
// because a list written here is one a new module can be left out of — and
// being left out is not visible: the type simply becomes creatable on an
// access token. That is how the upgrade type was admitted without a signature
// for a while. If some future operation genuinely does not need one, this test
// is where that decision gets written down and argued for.
func TestEveryRegisteredModuleRequiresASignature(t *testing.T) {
	types := DefaultRegistry().Types()
	if len(types) == 0 {
		t.Fatal("no modules are registered, so this proves nothing")
	}
	for _, typ := range types {
		if !RequiresSignature(typ) {
			t.Errorf("operation %q can be created without an owner signature", typ)
		}
	}
}
