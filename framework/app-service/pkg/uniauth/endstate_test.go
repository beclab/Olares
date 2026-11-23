package uniauth

import "testing"

func TestFixedHTTPMeshPosture(t *testing.T) {
	if EffectiveHTTPMeshWave() != 4 {
		t.Fatalf("wave=%d", EffectiveHTTPMeshWave())
	}
	if !TrafficGroupNPDesired() {
		t.Fatal("traffic-group NP must always be on")
	}
}
