package nodestatus

import (
	"context"
	"testing"
)

func stubProbe(context.Context, ProbeInput) (string, Capability, bool) {
	return "", Capability{}, false
}

func TestProbeRegistryRejectsEmptyName(t *testing.T) {
	if err := NewProbeRegistry().Register("", stubProbe); err == nil {
		t.Fatal("empty name registration succeeded")
	}
}

func TestProbeRegistryRejectsNilProbe(t *testing.T) {
	if err := NewProbeRegistry().Register("example", nil); err == nil {
		t.Fatal("nil probe registration succeeded")
	}
}

func TestProbeRegistryRejectsDuplicateName(t *testing.T) {
	reg := NewProbeRegistry()
	if err := reg.Register("example", stubProbe); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register("example", stubProbe); err == nil {
		t.Fatal("duplicate registration succeeded")
	}
}

func TestProbeRegistryFreezeRejectsFurtherRegistration(t *testing.T) {
	reg := NewProbeRegistry()
	if err := reg.Register("example", stubProbe); err != nil {
		t.Fatal(err)
	}
	reg.Freeze()
	if err := reg.Register("other", stubProbe); err == nil {
		t.Fatal("registration after Freeze succeeded")
	}
}

func TestProbeRegistryProbesPreserveRegistrationOrder(t *testing.T) {
	reg := NewProbeRegistry()
	first := func(context.Context, ProbeInput) (string, Capability, bool) {
		return "first", Capability{Supported: true}, true
	}
	second := func(context.Context, ProbeInput) (string, Capability, bool) {
		return "second", Capability{Supported: true}, true
	}
	if err := reg.Register("first", first); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register("second", second); err != nil {
		t.Fatal(err)
	}

	probes := reg.Probes()
	if len(probes) != 2 {
		t.Fatalf("Probes() len = %d, want 2", len(probes))
	}
	name, _, ok := probes[0](context.Background(), ProbeInput{})
	if !ok || name != "first" {
		t.Fatalf("first probe = %q ok=%v", name, ok)
	}
	name, _, ok = probes[1](context.Background(), ProbeInput{})
	if !ok || name != "second" {
		t.Fatalf("second probe = %q ok=%v", name, ok)
	}
}

func TestDefaultProbeRegistryHasBuiltInProbes(t *testing.T) {
	if n := len(DefaultProbeRegistry().Probes()); n < 4 {
		t.Fatalf("default registry has %d probes, want at least the four built-ins", n)
	}
}
