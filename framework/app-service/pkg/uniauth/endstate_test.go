package uniauth

import "testing"

func TestEndStateEnabled(t *testing.T) {
	t.Setenv(EndStateEnv, "")
	if EndStateEnabled() {
		t.Fatal("default off")
	}
	t.Setenv(EndStateEnv, "true")
	if !EndStateEnabled() {
		t.Fatal("expected on")
	}
}

func TestEffectiveHTTPMeshWave(t *testing.T) {
	parse := func(s string) int {
		if s == "2" {
			return 2
		}
		return 0
	}
	t.Setenv(EndStateEnv, "")
	if EffectiveHTTPMeshWave("2", parse) != 2 {
		t.Fatal("wave passthrough")
	}
	t.Setenv(EndStateEnv, "1")
	if EffectiveHTTPMeshWave("0", parse) != EndStateWave {
		t.Fatal("end-state forces wave 4")
	}
}

func TestTrafficGroupNPDesired(t *testing.T) {
	t.Setenv(EndStateEnv, "")
	if TrafficGroupNPDesired("") {
		t.Fatal("off")
	}
	if !TrafficGroupNPDesired("true") {
		t.Fatal("np env")
	}
	t.Setenv(EndStateEnv, "true")
	if !TrafficGroupNPDesired("") {
		t.Fatal("end-state implies NP")
	}
}
