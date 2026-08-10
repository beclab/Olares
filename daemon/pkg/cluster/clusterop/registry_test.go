package clusterop

import (
	"context"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

type registryTestModule struct{ typ Type }

func (m registryTestModule) Type() Type { return m.typ }

func (registryTestModule) Validate(CreateRequest) error { return nil }

func (registryTestModule) Phase(Operation) (nodestatus.Phase, bool) {
	return "", false
}

func (registryTestModule) Run(context.Context, Runtime, RunRequest) Outcome {
	return Outcome{Status: StatusSucceeded}
}

type registryTypedNilModule struct{}

func (*registryTypedNilModule) Type() Type { return Type("typed-nil") }

func (*registryTypedNilModule) Validate(CreateRequest) error { return nil }

func (*registryTypedNilModule) Phase(Operation) (nodestatus.Phase, bool) {
	return "", false
}

func (*registryTypedNilModule) Run(context.Context, Runtime, RunRequest) Outcome {
	return Outcome{Status: StatusSucceeded}
}

func TestRegistryParseRegisteredType(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(registryTestModule{typ: Type("example")}); err != nil {
		t.Fatal(err)
	}
	got, err := reg.Parse("example")
	if err != nil || got != Type("example") {
		t.Fatalf("Parse() = %q, %v", got, err)
	}
}

func TestRegistryRejectsWhitespaceWrappedType(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(registryTestModule{typ: Type("reboot")}); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.Parse(" reboot "); err == nil {
		t.Fatal("whitespace-wrapped type was accepted")
	}
}

func TestRegistryRejectsDuplicateType(t *testing.T) {
	reg := NewRegistry()
	module := registryTestModule{typ: Type("example")}
	if err := reg.Register(module); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(module); err == nil {
		t.Fatal("duplicate registration succeeded")
	}
}

func TestRegistryRejectsEmptyType(t *testing.T) {
	if err := NewRegistry().Register(registryTestModule{}); err == nil {
		t.Fatal("empty type registration succeeded")
	}
}

func TestRegistryRejectsNilModule(t *testing.T) {
	var module OperationModule
	if err := NewRegistry().Register(module); err == nil {
		t.Fatal("nil module registration succeeded")
	}
}

func TestRegistryRejectsTypedNilModule(t *testing.T) {
	var module *registryTypedNilModule
	if err := NewRegistry().Register(module); err == nil {
		t.Fatal("typed-nil module registration succeeded")
	}
}

func TestRegistryRejectsRegistrationAfterFreeze(t *testing.T) {
	reg := NewRegistry()
	reg.Freeze()
	if err := reg.Register(registryTestModule{typ: Type("late")}); err == nil {
		t.Fatal("registration after Freeze succeeded")
	}
}
