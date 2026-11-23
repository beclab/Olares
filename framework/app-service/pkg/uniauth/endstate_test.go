package uniauth

import (
	"os"
	"testing"
)

func TestEndStateDefaultsOn(t *testing.T) {
	t.Cleanup(func() { _ = os.Unsetenv(EndStateEnv) })
	_ = os.Unsetenv(EndStateEnv)
	if !EndStateEnabled() {
		t.Fatal("unset env must default to steady-state on")
	}
}

func TestEndStateExplicitOff(t *testing.T) {
	t.Setenv(EndStateEnv, "false")
	if EndStateEnabled() {
		t.Fatal("explicit false must disable")
	}
}

func TestEndStateExplicitOn(t *testing.T) {
	t.Setenv(EndStateEnv, "true")
	if !EndStateEnabled() {
		t.Fatal("explicit true")
	}
	if EffectiveHTTPMeshWave("0", func(string) int { return 0 }) != EndStateWave {
		t.Fatal("wave 4")
	}
	if !TrafficGroupNPDesired("") {
		t.Fatal("NP implied")
	}
}
